package engine

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/pobearm/workflow/util"
	"time"
)

const (
	// StatusStart 发起
	StatusStart = "started"
	// StatusFinished 结束
	StatusFinished = "finished"
	// StatusAbandon 丢弃
	StatusAbandon = "abandoned"
)

// NextStatuInfo 流程下一步的信息
type NextStatuInfo struct {
	// StepName 步骤名称
	StepName string
	// Users 流程参与人
	Users []*FlowUser
	// IsFree 是否自由流程
	IsFree bool
	// SelectType 如何选择人员 0表示只能选Users数组内的人 1表示可以选择所有的人
	SelectType bool
}

// Workflow 流程引擎对象
type Workflow struct {
	// FlowDef 流程的定义
	FlowDef *Flow
	// Fcase 流程实例信息
	Fcase *FlowCase
	// Appdata 流程节点运算需要的业务数据, 用json格式传递.
	Appdata string
	// FlowDefSRV 流程定义管理服务
	FlowDefSRV FlowDefineService
	// OrgSRV 组织架构/人员管理服务
	OrgSRV OrgService
	// FlowCaseSRV 流程实例服务
	CaseSRV FlowCaseService
}

// New_Workflow 创建一个新的工作流对象.
// connstr: db连接串, 这个是多租户模式需要, 一个流程服务可以服务于多个租户数据库, 租户按数据库隔离.
// 如果非多租户模式使用, db连接串写在服务配置文件.
func NewWorkflow(connstr string) (*Workflow, error) {
	fds, err := ServiceFactory{}.NewFlowDefineService(connstr)
	if err != nil {
		return nil, err
	}
	ogs, err := ServiceFactory{}.NewOrgService(connstr)
	if err != nil {
		return nil, err
	}
	css, err := ServiceFactory{}.NewCaseService(connstr)
	if err != nil {
		return nil, err
	}
	wf := &Workflow{
		FlowDefSRV: fds,
		OrgSRV:     ogs,
		CaseSRV:    css,
	}
	return wf, nil
}

// LoadWorkflow 加载一个已经存在的流程
// caseid: 流程实例id
// appdata: 流程节点运算需要的数据
func (w *Workflow) LoadWorkflow(caseid string, appdata string) error {
	w.Appdata = appdata
	fc, err := w.CaseSRV.LoadFlowCase(caseid)
	if err != nil {
		return err
	}
	w.Fcase = fc

	wfd, err := w.FlowDefSRV.GetFlowByVersionNo(fc.CaseInfo.FlowID, fc.CaseInfo.VersionNo)
	if err != nil {
		return err
	}
	w.FlowDef = wfd
	return nil
}

// CreateWorkflow 创建一个新的流程实例, 返回caseid
func (w *Workflow) CreateWorkflow(caseid, flowid, appdata string, user FlowUser) (string, string, error) {
	w.Appdata = appdata
	//找到流程定义
	wfd, err := w.FlowDefSRV.GetFlow(flowid)
	if err != nil {
		return "", "", err
	}
	w.FlowDef = wfd

	if caseid == "" {
		caseid = uuid.New().String()
	}
	statusname := ""
	for _, s := range w.FlowDef.FlowStatus {
		if s.Sequence == 0 {
			statusname = s.Name
			break
		}
	}
	if statusname == "" {
		return "", "", errors.New("get sequence 0 status name error")
	}
	w.Fcase = NewStartFlowCase(user.Userid, user.UserName, caseid, flowid, statusname, wfd.VersionNo)

	//保存到数据库
	serialnumberTemp := ""
	versionno := wfd.VersionNo
	if serialnumberTemp, err = w.CaseSRV.SaveNewCase(w.Fcase, versionno); err != nil {
		return "", "", err
	}
	return serialnumberTemp, caseid, nil
}

//
func (w *Workflow) PreCreateWorkflow(flowid string, appdata string, userid, username string) (string, error) {
	w.Appdata = appdata
	//找到流程定义
	wfd, err := w.FlowDefSRV.GetFlow(flowid)
	if err != nil {
		return "", err
	}
	w.FlowDef = wfd

	caseid := uuid.New().String()
	statusname := ""
	for _, s := range w.FlowDef.FlowStatus {
		if s.Sequence == 0 {
			statusname = s.Name
			break
		}
	}
	if statusname == "" {
		return "", errors.New("get sequence 0 status name error")
	}
	w.Fcase = NewStartFlowCase(userid, username, caseid, flowid, statusname, wfd.VersionNo)

	return caseid, nil
}

//预运算,计算一下步,以及下一步的处理人供用户选择
//返回, 下一步骤的名称, 处理人列表, 如果步骤名称等于"结束""作废", 特殊处理
func (w *Workflow) PreRun(itemid int32, choice string) (*NextStatuInfo, error) {

	if itemid > 0 && choice == "重新发起" {
		if w.Fcase.CaseItems[itemid-1].Choice == "退回" {
			return w.PreRun(itemid, "")
		}
	}

	//找到当前步骤
	ci, ok := w.Fcase.CaseItems[itemid]
	if !ok {
		return nil, fmt.Errorf("not found caseitem: %v in %s", itemid, w.Fcase.CaseInfo.CaseID)
	}
	if ci.StepStatus == StepStatusFinish {
		return nil, errors.New("step is finished: " + ci.StepName)
	}
	//检查一下流程的步骤名称,是否与流程状态吻合
	if w.Fcase.CaseInfo.Step != ci.StepName {
		return nil, errors.New("flow status not equal caseitem stepname")
	}
	//找到当前状态的定义
	cs, find := w.FlowDef.FlowStatus[w.Fcase.CaseInfo.Step]
	if !find {
		return nil, errors.New("not find flow status: " + w.Fcase.CaseInfo.Step)
	}
	//获得下一步应该去到的状态
	ns_name, err := cs.NextStatus(choice, w.Appdata)
	if err != nil {
		return nil, err
	}
	ns, find := w.findNextStatus(ns_name, cs.Sequence+1)
	if !find {
		//ns = w.findStatusSeq(seq)
		return nil, fmt.Errorf(`not find status name: "%s", seq: "%v"`,
			w.Fcase.CaseInfo.Step, cs.Sequence+1)
	}
	nsInfo := &NextStatuInfo{
		StepName:   ns.Name,
		IsFree:     ns.IsFree(),
		SelectType: true,
	}

	//计算下一步的处理人
	//如果步骤名称等于"通过","不通过", 下一步处理人为空
	if ns.Name == StatusFinished || ns.Name == StatusAbandon {
		nsInfo.Users = make([]*FlowUser, 0, 0)
		nsInfo.SelectType = false
		return nsInfo, nil
	} else if ns.Name == StatusStart {
		//步骤名称等于"发起", 处理人就是发起人
		us := make([]*FlowUser, 0, 1)
		startItem := w.Fcase.CaseItems[0]
		u := &FlowUser{
			UserName: startItem.HandleUserName,
			Userid:   startItem.HandleUserid,
		}
		us = append(us, u)
		nsInfo.Users = us
		nsInfo.SelectType = false
		return nsInfo, nil
	} else {
		if ns.Partici == nil {
			return nsInfo, errors.New("step not defined participant")
		}
		us, err := ns.Partici.FindUser(w.OrgSRV, w.Fcase)
		if err != nil {
			return nil, err
		}
		if us != nil && len(us) > 0 {
			nsInfo.SelectType = false
		}
		nsInfo.Users = us
		return nsInfo, nil
	}
}

func (w *Workflow) findNextStatus(ns_name string, seq int) (*Status, bool) {
	if ns_name != "" {
		s, ok := w.FlowDef.FlowStatus[ns_name]
		return s, ok
	} else {
		//处理ns_name为空字符串的情况, (当status没有定义choice的时候为空,或者transition中的定义为空)
		//此时status自己计算不出下一步要去的步骤, 默认去到sequence+1的状态.
		return w.findStatusSeq(seq)
	}
}

//根据sequence来找到下一步的状态
func (w *Workflow) findStatusSeq(seq int) (*Status, bool) {
	for _, s := range w.FlowDef.FlowStatus {
		if s.Sequence == seq {
			return s, true
		}
	}
	return nil, false
}

//用户做出了选择, 并且选择了下一步的处理人后,提交流程.
//itemid: 当前流程实例的步骤id, caseitem.
//choice: 用户审批的选择, 如 [同意][不同意].
//mark:   审批时填写的备注.
//user:   用户选择的人下一步处理人.
//返回值
//string  进入的状态名称 stepname.
//error   错误
func (w *Workflow) Run(itemid int32, choice, mark string, user *FlowUser) (string, error) {
	//找到当前步骤
	ci, ok := w.Fcase.CaseItems[itemid]
	if !ok {
		return "", fmt.Errorf("not found caseitem: %v in %s", itemid, w.Fcase.CaseInfo.CaseID)
	}
	if ci.StepStatus == StepStatusFinish {
		return "", errors.New("step is finished: " + ci.StepName)
	}
	//检查一下流程的步骤名称,是否与流程状态吻合
	if w.Fcase.CaseInfo.Step != ci.StepName {
		return "", errors.New("flow status not equal caseitem stepname")
	}
	//找到当前状态的定义
	cs, find := w.FlowDef.FlowStatus[ci.StepName]
	if !find {
		return "", errors.New("not find flow status: " + w.Fcase.CaseInfo.Step)
	}
	//获得下一步应该去到的状态
	ns_name, err := cs.NextStatus(choice, w.Appdata)
	if err != nil {
		return "", err
	}
	//----------------------------------------------------------------------------
	return w.runIntoStep(ns_name, choice, mark, cs, ci, user)
}

func (w *Workflow) runIntoStep(ns_name, choice, mark string, cs *Status,
	ci *CaseItem, user *FlowUser) (string, error) {

	ns, find := w.findNextStatus(ns_name, cs.Sequence+1)
	if !find {
		return "", fmt.Errorf(`not find status name: "%s", seq: "%v"`,
			w.Fcase.CaseInfo.Step, cs.Sequence+1)
	}
	//处理当前步骤的数据, 但还没有更新到数据库
	if err := w.HandlCurrentStep(choice, mark, ci, cs); err != nil {
		return "", err
	}
	//处理下一步骤的数据, 但还没有更新到数据库
	//自由流程时, 如果没有选择人, 就直接跳到[通过]
	if ns.IsFree() && user == nil {
		ns = w.FlowDef.FlowStatus[StatusFinished]
	}

	//如果选择项是不通过（中止） 跳转[不通过]
	if choice == StatusAbandon {
		ns = w.FlowDef.FlowStatus[StatusAbandon]
	}

	ni, err := w.HandNextStep(int32(ci.ItemID+1), ns, user, w.Appdata)
	if err != nil {
		return "", err
	}

	//如果,一下步去到[通过], case.status=1, 到[不通过], case.status=2
	if ni.StepName == StatusFinished {
		w.Fcase.CaseInfo.Status = 1
	}
	if ni.StepName == StatusAbandon {
		w.Fcase.CaseInfo.Status = 2
	}
	//更新到数据库
	if err := w.CaseSRV.ComitFlow(w.Fcase.CaseInfo, ci, ni); err != nil {
		return "", err
	}
	//----------------------------------------------------------------------
	//执行当前步骤的退出处理器
	if info, err := cs.OnExit(w.Appdata, w.Fcase, ci.ItemID); err != nil {
		//处理器, 处理失败, 只是记录日志, 不影响流程提交
		ci.SysExitInfo = err.Error()
	} else {
		ci.SysExitInfo = info
	}
	//执行下一个步骤的进入处理器(消息推送是一个处理器)
	if info, err := ns.OnEnter(w.Appdata, w.Fcase, ni.ItemID); err != nil {
		//处理器, 处理失败, 只是记录日志, 不影响流程提交
		ni.SysEnterInfo = err.Error()
	} else {
		ni.SysEnterInfo = info
	}
	//记录步骤OnExit, OnEnter的处理结果
	if err := w.CaseSRV.StepHandled(w.Fcase.CaseInfo, ci, ni); err != nil {
		return "", err
	}

	//push msg todo
	//回写发送时间
	caseinfo := &CaseInfo{
		CaseID: w.Fcase.CaseInfo.CaseID,
		ItemID: ni.ItemID,
	}
	err = w.CaseSRV.WriteBackSendTime(caseinfo)
	if err != nil {
		return "", err
	}

	return ni.StepName, nil
}

//跳转到步骤, 如: 通过, 不通过
func (w *Workflow) JumpToStep(itemid int32, stepname, choice, mark string, user *FlowUser) error {
	//找到当前步骤
	ci, ok := w.Fcase.CaseItems[itemid]
	if !ok {
		return fmt.Errorf("not found caseitem: %v in %s", itemid, w.Fcase.CaseInfo.CaseID)
	}
	//检查一下流程的步骤名称,是否与流程状态吻合
	if w.Fcase.CaseInfo.Step != ci.StepName {
		return errors.New("flow status not equal caseitem stepname")
	}
	//找到当前状态的定义
	cs, find := w.FlowDef.FlowStatus[ci.StepName]
	if !find {
		return errors.New("not find flow status: " + w.Fcase.CaseInfo.Step)
	}
	//todo: 应该调用runIntoStep, 是同样的逻辑?
	//----------------------------------------------------------------------
	//获得下一步应该去到的状态
	ns, find := w.FlowDef.FlowStatus[stepname]
	if !find {
		return errors.New("do not find step: " + stepname)
	}
	//处理当前步骤的数据, 但还没有更新到数据库
	if err := w.HandlCurrentStep(choice, mark, ci, cs); err != nil {
		return err
	}
	//处理下一步骤的数据, 但还没有更新到数据库
	ni, err := w.HandNextStep(int32(ci.ItemID+1), ns, user, w.Appdata)
	if err != nil {
		return err
	}
	//如果,一下步去到[通过], case.status=1, 到[不通过], case.status=2
	if ni.StepName == StatusFinished {
		w.Fcase.CaseInfo.Status = 1
	}
	if ni.StepName == StatusAbandon {
		w.Fcase.CaseInfo.Status = 2

		choiceName := "中止"
		ci.Choice = choiceName
	}
	//最后一次更新到数据库
	if err := w.CaseSRV.ComitFlow(w.Fcase.CaseInfo, ci, ni); err != nil {
		return err
	}
	//----------------------------------------------------------------------
	//执行当前步骤的退出处理器
	if info, err := cs.OnExit(w.Appdata, w.Fcase, ci.ItemID); err != nil {
		//处理器, 处理失败, 只是记录日志, 不影响流程提交
		ci.SysExitInfo = err.Error()
	} else {
		ci.SysExitInfo = info
	}
	//执行下一个步骤的进入处理器(消息推送是一个处理器)
	if info, err := ns.OnEnter(w.Appdata, w.Fcase, ni.ItemID); err != nil {
		//处理器, 处理失败, 只是记录日志, 不影响流程提交
		ni.SysEnterInfo = err.Error()
	} else {
		ni.SysEnterInfo = info
	}
	//记录步骤OnExit, OnEnter的处理结果
	if err := w.CaseSRV.StepHandled(w.Fcase.CaseInfo, ci, ni); err != nil {
		return err
	}

	//push msg todo
	//回写发送时间
	caseinfo := &CaseInfo{
		CaseID: w.Fcase.CaseInfo.CaseID,
		ItemID: w.Fcase.CaseInfo.ItemID,
	}
	err = w.CaseSRV.WriteBackSendTime(caseinfo)
	if err != nil {
		return err
	}
	return nil
}

// HandlCurrentStep 处理当前步骤的数据
func (w *Workflow) HandlCurrentStep(choice, mark string, ci *CaseItem, s *Status) error {
	if ci.StepStatus == StepStatusFinish {
		return errors.New("step already finished")
	}
	ci.Choice = choice
	ci.Mark = mark
	ci.HandleTime = time.Now().Format(util.DatetimeHMS)
	ci.StepStatus = StepStatusFinish
	return nil
}

// HandNextStep 处理下一步骤的数据
func (w *Workflow) HandNextStep(nid int32, ns *Status, user *FlowUser, appdata string) (*CaseItem, error) {
	var ci *CaseItem
	//如果步骤名称等于"通过""不通过", 下一步处理人, 为空
	if ns.Name == StatusFinished || ns.Name == StatusAbandon {
		//新建下一步的caseitem
		ci = NewCaseItem(nid, ns.Name, "", "")
		//直接进入结束状态, 不产生待办
		ci.StepStatus = StepStatusFinish
		ci.HandleTime = time.Now().Format(util.DatetimeHMS)
		//修改case的数据
		w.Fcase.CaseInfo.EndTime = time.Now()
		//即使是返回发起人, prerun的时候也已经找到人了, 所以这里不用再判断了
	} else {
		if user == nil {
			return nil, errors.New("step user is nil")
		}
		//新建下一步的caseitem
		ci = NewCaseItem(nid, ns.Name, user.Userid, user.UserName)
		//判断是否有设置代理人, 如果有就标记代理人
		agent, ok := w.CaseSRV.FindAgent(user.Userid)
		if ok {
			ci.SetAgent(agent.Userid, agent.UserName)
		}
		ci.StepStatus = StepStatusNew
	}
	//修改case的数据
	w.Fcase.CaseInfo.Step = ns.Name
	w.Fcase.CaseItems[ci.ItemID] = ci
	return ci, nil
}
