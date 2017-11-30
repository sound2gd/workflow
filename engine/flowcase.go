package workflow

import (
	"time"
)

//流程实例的数据
type FlowCase struct {
	CaseInfo  *Case               //`json:"case"`      //流程实例信息
	CaseItems map[int32]*CaseItem //`json:"caseitems"` //流程的步骤记录
}

type FlowCaseList struct {
	CaseInfo  *Case       //`json:"case"`      //流程实例信息
	CaseItems []*CaseItem //`json:"caseitems"` //流程的步骤记录
}

type Case struct {
	CaseId         string    `json:"caseid"` //实例id
	ItemId         int32     `json:"itemid"` //当前实例顺序
	AppId          string    `json:"appid"`
	BizId1         string    `json:"bizid1"`
	BizId2         string    `json:"bizid2"`
	FlowId         string    `json:"flowid"`         //流程定义id
	FlowName       string    `json:"flowname"`       //流程名称
	CreatorId      string    `json:"creatorid"`      //创建人id
	CreatorName    string    `json:"creatorname"`    //创建人名字
	Step           string    `json:"stepname"`       //当前状态
	Status         int32     `json:"status"`         //状态 0:审批中 1:通过 2:不通过
	CreateTime     time.Time `json:"createtime"`     //创建时间
	EndTime        time.Time `json:"endtime"`        //结束时间
	CopyUser       []int     `json:"copyuser"`       //抄送人
	AppData        string    `json:"appdata"`        //表单json数据
	SendTime       string    `json:"sendtime"`       //发送时间
	ChoiceItems    string    `json:"choiceitems"`    //审核选项
	SerialNumber   string    `json:"serialnumber"`   //流水号
	HandleUserid   string    `json:"handleuserid"`   //步骤处理人
	HandleUserName string    `json:"handleusername"` //处理人名字
	HandleTime     string    `json:"handletime"`     //处理时间
	StepStatus     int32     `json:"stepstatus"`     //处理状态 0未读1已读2已处理
	PluginId       string    `json:"pluginid"`
	VersionNo      int32     `json:"versionno"`
}

func New_StartFlowCase(creatorid, creatorname, caseid, flowid, step string, versionno int32) *FlowCase {
	c := &Case{
		CaseId:      caseid,
		FlowId:      flowid,
		CreatorId:   creatorid,
		CreatorName: creatorname,
		Step:        step,
		Status:      0,
		CreateTime:  time.Now(),
		VersionNo:   versionno,
	}
	ci := New_CaseItem(0, step, creatorid, creatorname)
	cis := make(map[int32]*CaseItem)
	cis[ci.ItemId] = ci
	return &FlowCase{CaseInfo: c, CaseItems: cis}
}

//流程实例的任务数据
type CaseItem struct {
	ItemId         int32     `json:"itemid"`         //步骤id
	HandleUserid   string    `json:"handleuserid"`   //步骤处理人
	HandleUserName string    `json:"handleusername"` //处理人名字
	StepName       string    `json:"stepname"`       //这个步骤的状态名字
	Choice         string    `json:"choice"`         //用户的选择结果
	Mark           string    `json:"mark"`           //处理人的备注
	CreateTime     time.Time `json:"createtime"`     //发起时间
	HandleTime     string    `json:"handletime"`     //处理时间
	AgentUserid    string    `json:"agentuserid"`    //代理人id
	AgentUserName  string    `json:"agentusername"`  //代理人名字
	StepStatus     int32     `json:"stepstatus"`     //流程步骤的状态
	SysEnterInfo   string    `json:"sysenterinfo"`   //进入步骤,系统信息记录
	SysExitInfo    string    `json:"sysexitinfo"`    //离开步骤,系统信息记录
	ChoiceItems    string    `json:"choiceitems"`    //审核选项
}

//流程步骤的状态
const (
	StepStatus_New int32 = iota
	StepStatus_Read
	StepStatus_Finish
)

func New_CaseItem(itemid int32, stepname, userid, username string) *CaseItem {
	return &CaseItem{
		ItemId:         itemid,
		StepName:       stepname,
		HandleUserid:   userid,
		HandleUserName: username,
		CreateTime:     time.Now(),
		StepStatus:     StepStatus_New,
	}
}

//设置代理人
func (c *CaseItem) SetAgent(userid, username string) {
	c.AgentUserid = userid
	c.AgentUserName = username
}
