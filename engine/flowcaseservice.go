package engine

// FlowCaseService 流程实例服务
type FlowCaseService interface {

	// LoadFlowCase 加载一个流程的完整信息
	LoadFlowCase(caseid string) (fc *FlowCase, err error)

	// SaveNewCase 保存一个新的流程实例
	SaveNewCase(fc *FlowCase, versionno int32) (string, error)

	// ComitFlow 在一个事务中提交流程数据,case,当前item,下一步item
	ComitFlow(c *Case, ci *CaseItem, ni *CaseItem) error

	// FindAgent 找到步骤处理人的代理人
	FindAgent(userid string) (user *FlowUser, find bool)

	// StepHandled 记录步骤进入,退出的消息
	StepHandled(ca *Case, ci *CaseItem, ni *CaseItem) error

	// WriteBackSendTime 更新发送时间??
	WriteBackSendTime(caseinfo *CaseInfo) error
}
