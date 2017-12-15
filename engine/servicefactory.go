package engine

// ServiceFactory 服务工厂
type ServiceFactory struct {
}

// NewFlowDefineService 创建流程定义服务
func (sf ServiceFactory) NewFlowDefineService(conn string) (FlowDefineService, error) {
	return nil, nil
}

// NewOrgService 创建组织服务
func (sf ServiceFactory) NewOrgService(conn string) (OrgService, error) {
	return nil, nil
}

// NewCaseService 创建组织服务
func (sf ServiceFactory) NewCaseService(conn string) (FlowCaseService, error) {
	return nil, nil
}
