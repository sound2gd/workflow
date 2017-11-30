package workflow

import (
	"time"
)

// CaseInfo,DTO对象,流程事务
type CaseInfo struct {
	CaseId         string    `json:"caseid"`         //流程实例id
	ItemId         int32     `json:"itemid"`         //流程当前步骤id
	FlowId         string    `json:"flowid"`         //流程id
	Name           string    `json:"flowname"`       //流程名称
	Creator        string    `json:"creator"`        //流程发起人账号
	Creatorname    string    `json:"creatorname"`    //流程发起人姓名
	Createtime     time.Time `json:"createtime"`     //流程发起时间
	Handleuserid   string    `json:"handleuserid"`   //步骤原处理人(有代理人)
	Handleusername string    `json:"handleusername"` //步骤原处理人姓名
	Handletime     string    `json:"handletime"`     //处理时间
	ChoiceItems    string    `json:"choiceitems"`    //审核选项
	Stepname       string    `json:"stepname"`       //当前步骤名称
	Stepstatus     int32     `json:"stepstatus"`     //当前步骤的状态，0:未处理 1:已读 2:已处理
	Status         int32     `json:"status"`         //状态,0:审批中 1:通过 2:不通过
	Appid          string    `json:"appid"`          //流程关联的业务对象(记录在crm_t_entityreg)
	Bizid1         string    `json:"bizid1"`         //业务主键1
	Bizid2         string    `json:"bizid2"`         //业务主键2
	SendTime       string    `json:"sendtime"`       //发送时间
	SerialNumber   string    `json:"serialnumber"`   //流水号
	Choice         string    `json:"choice"`         //审核
	PluginId       string    `json:"pluginid"`       //插件id
	FlowStatus     int32     `json:"flowstatus"`     //流程状态 1启用0停用
}

//代办事务
type CaseList struct {
	Items      []*CaseInfo
	TotalItems int32
}

//流程列表
type FlowList struct {
	Items      []*FlowInfo
	TotalItems int32
}

//流程的信息
type FlowInfo struct {
	FlowId         string    `json:"flowid"`
	Name           string    `json:"flowname"`
	Descript       string    `json:"descript"`
	FlowXml        string    `json:"flowxml"`
	StepCount      int32     `json:"stepcount"`
	CreateTime     time.Time `json:"createtime"`
	Creator        string    `json:"creator"`
	Status         int32     `json:"status"`
	UpdateTime     string    `json:"updatetime"`
	Updator        string    `json:"updator"`
	FlowType       int32     `json:"flowtype"`
	AppId          string    `json:"appid"`
	EntityType     int32     `json:"entitytype"`     //1系统对象2插件对象
	FlowCategory   int32     `json:"flowcategory"`   //1表示固定流程，0表示自由流程
	PluginStatus   int32     `json:"pluginstatus"`   //插件状态 1在用
	PVersionStatus int32     `json:"pversionstatus"` //插件版本
	PowerControl   int32     `json:"powercontrol"`   //权限控制
}

type FlowService interface {
	//获取用户的代办列表---flowname查询条件
	GetTodoCases(flowname, usernumber string, pageindex, pagesize int32) (*CaseList, error)

	//获取用户事务列表
	GetMyCases(usernumber string, finishstate, filter, pageindex, pagesize int32,
		flowid, keyword, begintime, endtime, createtime, handletime, sorttype string) (*CaseList, error)

	//
	GetWorkFlows(status, flowname string, pageindex, pagesize int32) (*FlowList, error)

	//
	GetWorkFlowsForMobile(status, flowname string, pageindex, pagesize int32) (*FlowList, error)

	//
	GetWorkFlowsForWeb(status, flowname string,
		pageindex, pagesize int32) (*FlowList, error)

	//获取流程定义列表
	WorkFlows(status, flowname string, pageindex, pagesize int32,
		dynamic_sql string) (*FlowList, error)

	//
	GetWorkFlowDetail(flowid string) (*FlowInfo, error)

	//流程实例详情
	GetCaseDetail(caseid string) (*FlowCaseList, error)

	//新发起一个流程, 返回caseid
	AddCase(enterprise, caseid, flowid, flowname, usernumber, username, biz_1, biz_2,
		appid, handeruserid, handerusername string, copyuser []int,
		appdata, remark string) (string, string, error)

	//预新发起一个流程, 返回步骤和人
	PreAddCase(flowid, usernumber, username, appdata string) (*NextStatuInfo, error)

	//预提交, 选择审批选项, 返回下一步去到的步骤和可选审批人
	PreCommitCase(caseid, choice string, itemid int32,
		appdata string) (nsif *NextStatuInfo, err error)

	//处理待办项, 返回进入的状态名称
	CommitCase(enterprise, usernumber, caseid, choice, remark string, itemid int32,
		flowuser *FlowUser,
		appdata string) (string, error)

	//作废流程实列
	AbandonCase(enterprise, usernumber, caseid, choice, remark string, itemid int32,
		appdata string) error
	//结束流程实列
	FinishCase(caseid, choice, remark string, itemid int32,
		appdata string) error
	//流程实列, 退回到发起人
	SendbackCase(enterprise, usernumber, caseid, choice, remark string, itemid int32,
		appdata string) error

	//流程实列, 退回给上一个步骤
	FallbackCase(caseid, choice, remark string, itemid int32,
		appdata string) error

	//标记流程步骤为已读
	Readed(itemid int32, caseid, usernumber string) error

	//设置代理人
	SetAgent(userid, agentid string) error

	//取消代理人
	UnsetAgent(userid string) error

	//动态获取审批选项
	GetDynamicSel(flowid, stepname string) ([]*Choice, error)

	//
	WBStepStatus(itemid int32, caseid, usernumber string)
}
