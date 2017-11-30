package workflow

import (
	"errors"
	"fmt"
	"time"
)

const (
	Status_Start    = "发起"
	Status_Finished = "通过"
	Status_Abandon  = "不通过"
)

//流程下一步的信息
type NextStatuInfo struct {
	StepName   string
	Users      []*FlowUser
	IsFree     bool
	SelectType bool //选择人员类型 1表示可以选择所有的人，0表示只能选Users数组的人
}

type Workflow struct {
	FlowDef    *Flow
	Fcase      *FlowCase
	Appdata    string
	FlowProvid FlowProvider
	OrgProvid  OrgProvider
}

func New_Workflow(connstr string) (*Workflow, error) {
	fp, err := New_FlowPgProvider(connstr)
	if err != nil {
		return nil, err
	}
	op, err := New_OrgPgProvider(connstr)
	if err != nil {
		return nil, err
	}
	wf := &Workflow{
		FlowProvid: fp,
		OrgProvid:  op,
	}
	return wf, nil
}

//加载一个已经存在的流程
func (w *Workflow) LoadWorkflow(caseid string, appdata string) error {
	w.Appdata = appdata
	fc, err := w.FlowProvid.LoadFlowCase(caseid)
	if err != nil {
		return err
	}
	w.Fcase = fc

	wfd, err := w.FlowProvid.GetFlowByVersionNo(fc.CaseInfo.FlowId, fc.CaseInfo.VersionNo)
	if err != nil {
		return err
	}
	w.FlowDef = wfd
	return nil
}

//加载流程数据, 并解析xml
// func (w *Workflow) GetFlow(flowid string) (*Flow, error) {
// 	return w.FlowProvid.GetFlow(flowid)
// }

//创建一个新的流程实例, 返回caseid
func (w *Workflow) CreateWorkflow(caseid, flowid string, appdata string,
	userid, username string) (string, string, error) {

	w.Appdata = appdata
	//找到流程定义
	wfd, err := w.FlowProvid.GetFlow(flowid)
	if err != nil {
		return "", "", err
	}
	w.FlowDef = wfd

	log.Info("before.CreateWorkflow.caseid", caseid)
	if caseid == "" {
		caseid = uuid.NewV4().String()
	}
	log.Info("after.CreateWorkflow.caseid", caseid)

	log.Info("workflow.CreateWorkflow.w.FlowDef", w.FlowDef)
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
	w.Fcase = New_StartFlowCase(userid, username, caseid, flowid, statusname, wfd.VersionNo)
	//保存到数据库
	serialnumberTemp := ""
	versionno := wfd.VersionNo
	if serialnumberTemp, err = w.FlowProvid.SaveNewCase(w.Fcase, versionno); err != nil {
		return "", "", err
	}
	log.Debugf("workflow.CreateWorkflow", "%#v", w)
	return serialnumberTemp, caseid, nil
}

//
func (w *Workflow) PreCreateWorkflow(flowid string, appdata string, userid, username string) (string, error) {
	w.Appdata = appdata
	//找到流程定义
	wfd, err := w.FlowProvid.GetFlow(flowid)
	if err != nil {
		return "", err
	}
	w.FlowDef = wfd

	log.Info("workflow.CreateWorkflow.w.FlowDef.begin")
	log.Info("workflow.CreateWorkflow.w.FlowDef", wfd)
	log.Info("workflow.CreateWorkflow.w.FlowDef.end")

	caseid := uuid.NewV4().String()
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
	w.Fcase = New_StartFlowCase(userid, username, caseid, flowid, statusname, wfd.VersionNo)

	log.Debugf("workflow.PreCreateWorkflow", "%#v", w)
	return caseid, nil
}

//---------------------------------------------
//预运算,计算一下步,以及下一步的处理人供用户选择
//返回, 下一步骤的名称, 处理人列表, 如果步骤名称等于"结束""作废", 特殊处理
func (w *Workflow) PreRun(itemid int32, choice string) (*NextStatuInfo, error) {
	log.Debug("workflow.PreRun")

	if itemid > 0 && choice == "重新发起" {
		log.Debug(w.Fcase.CaseItems[itemid-1].Choice)
		if w.Fcase.CaseItems[itemid-1].Choice == "退回" {
			log.Debug(choice)
			return w.PreRun(itemid, "")
		}
	}

	//找到当前步骤
	ci, ok := w.Fcase.CaseItems[itemid]
	if !ok {
		return nil, fmt.Errorf("not found caseitem: %v in %s", itemid, w.Fcase.CaseInfo.CaseId)
	}
	if ci.StepStatus == StepStatus_Finish {
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
	if ns.Name == Status_Finished || ns.Name == Status_Abandon {
		nsInfo.Users = make([]*FlowUser, 0, 0)
		nsInfo.SelectType = false
		return nsInfo, nil
	} else if ns.Name == Status_Start {
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
		us, err := ns.Partici.FindUser(w.OrgProvid, w.Fcase)
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
		return "", fmt.Errorf("not found caseitem: %v in %s", itemid, w.Fcase.CaseInfo.CaseId)
	}
	if ci.StepStatus == StepStatus_Finish {
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

	log.Debug("workflow.runIntoStep")
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
		ns = w.FlowDef.FlowStatus[Status_Finished]
	}

	//如果选择项是不通过（中止） 跳转[不通过]
	if choice == Status_Abandon {
		ns = w.FlowDef.FlowStatus[Status_Abandon]
	}

	ni, err := w.HandNextStep(int32(ci.ItemId+1), ns, user, w.Appdata)
	if err != nil {
		return "", err
	}

	//如果,一下步去到[通过], case.status=1, 到[不通过], case.status=2
	if ni.StepName == Status_Finished {
		w.Fcase.CaseInfo.Status = 1
	}
	if ni.StepName == Status_Abandon {
		w.Fcase.CaseInfo.Status = 2
	}
	//更新到数据库
	if err := w.FlowProvid.ComitFlow(w.Fcase.CaseInfo, ci, ni); err != nil {
		return "", err
	}
	//----------------------------------------------------------------------
	//执行当前步骤的退出处理器
	if info, err := cs.OnExit(w.Appdata, w.Fcase, ci.ItemId); err != nil {
		//处理器, 处理失败, 只是记录日志, 不影响流程提交
		ci.SysExitInfo = err.Error()
	} else {
		ci.SysExitInfo = info
	}
	//执行下一个步骤的进入处理器(消息推送是一个处理器)
	if info, err := ns.OnEnter(w.Appdata, w.Fcase, ni.ItemId); err != nil {
		//处理器, 处理失败, 只是记录日志, 不影响流程提交
		ni.SysEnterInfo = err.Error()
	} else {
		ni.SysEnterInfo = info
	}
	//记录步骤OnExit, OnEnter的处理结果
	if err := w.FlowProvid.StepHandled(w.Fcase.CaseInfo, ci, ni); err != nil {
		return "", err
	}

	//push msg todo
	//回写发送时间
	caseinfo := &CaseInfo{
		CaseId: w.Fcase.CaseInfo.CaseId,
		ItemId: ni.ItemId,
	}
	err = w.FlowProvid.WriteBackSendTime(caseinfo)
	if err != nil {
		log.Error("workflow.runIntoStep.WriteBackSendTime", err.Error())
	}

	return ni.StepName, nil
}

//跳转到步骤, 如: 通过, 不通过
func (w *Workflow) JumpToStep(itemid int32, stepname, choice, mark string, user *FlowUser) error {
	log.Debug("workflow.JumpToStep")
	//找到当前步骤
	ci, ok := w.Fcase.CaseItems[itemid]
	if !ok {
		return fmt.Errorf("not found caseitem: %v in %s", itemid, w.Fcase.CaseInfo.CaseId)
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
	ni, err := w.HandNextStep(int32(ci.ItemId+1), ns, user, w.Appdata)
	if err != nil {
		return err
	}
	//如果,一下步去到[通过], case.status=1, 到[不通过], case.status=2
	if ni.StepName == Status_Finished {
		w.Fcase.CaseInfo.Status = 1
	}
	if ni.StepName == Status_Abandon {
		w.Fcase.CaseInfo.Status = 2

		choiceName := "中止"
		ci.Choice = choiceName
	}
	//最后一次更新到数据库
	if err := w.FlowProvid.ComitFlow(w.Fcase.CaseInfo, ci, ni); err != nil {
		return err
	}
	//----------------------------------------------------------------------
	//执行当前步骤的退出处理器
	if info, err := cs.OnExit(w.Appdata, w.Fcase, ci.ItemId); err != nil {
		//处理器, 处理失败, 只是记录日志, 不影响流程提交
		ci.SysExitInfo = err.Error()
	} else {
		ci.SysExitInfo = info
	}
	//执行下一个步骤的进入处理器(消息推送是一个处理器)
	if info, err := ns.OnEnter(w.Appdata, w.Fcase, ni.ItemId); err != nil {
		//处理器, 处理失败, 只是记录日志, 不影响流程提交
		ni.SysEnterInfo = err.Error()
	} else {
		ni.SysEnterInfo = info
	}
	//记录步骤OnExit, OnEnter的处理结果
	if err := w.FlowProvid.StepHandled(w.Fcase.CaseInfo, ci, ni); err != nil {
		return err
	}

	//push msg todo
	//回写发送时间
	caseinfo := &CaseInfo{
		CaseId: w.Fcase.CaseInfo.CaseId,
		ItemId: w.Fcase.CaseInfo.ItemId,
	}
	err = w.FlowProvid.WriteBackSendTime(caseinfo)
	if err != nil {
		log.Error("workflow.JumpToStep.WriteBackSendTime", err.Error())
	}
	return nil
}

//处理当前步骤的数据
func (w *Workflow) HandlCurrentStep(choice, mark string, ci *CaseItem, s *Status) error {
	if ci.StepStatus == StepStatus_Finish {
		return errors.New("step already finished")
	}
	ci.Choice = choice
	ci.Mark = mark
	ci.HandleTime = time.Now().Format(f_datetime)
	ci.StepStatus = StepStatus_Finish
	return nil
}

//处理下一步骤的数据
func (w *Workflow) HandNextStep(nid int32, ns *Status, user *FlowUser, appdata string) (*CaseItem, error) {
	var ci *CaseItem
	//如果步骤名称等于"通过""不通过", 下一步处理人, 为空
	if ns.Name == Status_Finished || ns.Name == Status_Abandon {
		//新建下一步的caseitem
		ci = New_CaseItem(nid, ns.Name, "", "")
		//直接进入结束状态, 不产生待办
		ci.StepStatus = StepStatus_Finish
		ci.HandleTime = time.Now().Format(f_datetime)
		//修改case的数据
		w.Fcase.CaseInfo.EndTime = time.Now()
		//即使是返回发起人, prerun的时候也已经找到人了, 所以这里不用再判断了
	} else {
		if user == nil {
			return nil, errors.New("step user is nil")
		}
		//新建下一步的caseitem
		ci = New_CaseItem(nid, ns.Name, user.Userid, user.UserName)
		//判断是否有设置代理人, 如果有就标记代理人
		agent, ok := w.FlowProvid.FindAgent(user.Userid)
		if ok {
			ci.SetAgent(agent.Userid, agent.UserName)
		}
		ci.StepStatus = StepStatus_New
	}
	//修改case的数据
	w.Fcase.CaseInfo.Step = ns.Name
	w.Fcase.CaseItems[ci.ItemId] = ci
	return ci, nil
}
