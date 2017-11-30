package workflow

const (
	f_datetime = "2006-01-02 15:04:05"
)

type FlowProvider interface {
	//获得流程的定义（最新版本）属性:flowid, name, descript, flowxml
	GetFlow(flowid string) (flow *Flow, err error)
	//获取特定版本的流程定义
	GetFlowByVersionNo(flowid string, versionno int32) (flow *Flow, err error)
	//加载一个流程的完整信息
	LoadFlowCase(caseid string) (fc *FlowCase, err error)
	//保存一个新的流程实例
	SaveNewCase(fc *FlowCase, versionno int32) (string, error)
	//在一个事务中提交流程数据,case,当前item,下一步item
	ComitFlow(c *Case, ci *CaseItem, ni *CaseItem) error
	//找到步骤处理人的代理人
	FindAgent(userid string) (user *FlowUser, find bool)
	//记录步骤进入,退出的消息
	StepHandled(ca *Case, ci *CaseItem, ni *CaseItem) error

	WriteBackSendTime(caseinfo *CaseInfo) error
}
