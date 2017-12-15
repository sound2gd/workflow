package engine

// FlowDefineService 流程定义服务接口
type FlowDefineService interface {
	//保存一个新的流程定义
	AddFlow(flow *FlowInfo, appid string) error

	//删除一个流程定义
	DeleteFlow(flowid string) error

	//修改一个流程定义
	UpdateFlow(flow *FlowInfo) error

	//启用流程
	EnableFlow(flow *FlowInfo) error

	//停用流程
	DisableFlow(flow *FlowInfo) error

	// GetFlow 获得流程的定义（最新版本）属性:flowid, name, descript, flowxml
	GetFlow(flowid string) (flow *Flow, err error)

	// GetFlowByVersionNo 获取特定版本的流程定义
	GetFlowByVersionNo(flowid string, versionno int32) (flow *Flow, err error)
}
