package workflow

// DefineService 流程定义服务接口
type DefineService interface {
	//流程定义详情
	GetFlow(flowid string) (*FlowInfo, error)

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
}
