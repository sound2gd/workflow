package workflow

type OrgProvider interface {
	//根据角色,  部门id找到用户
	FindUser(role, departid string) (us []*FlowUser, err error)
	//找到用户直属部门
	FindUserDept(userid string) (deptid string, err error)
	//找到用户直属部门的父部门
	FindUserParentDept(userid string) (deptid string, err error)
	//根据用户id,获取用户
	GetUser(userid string) (us []*FlowUser, err error)
	//根据部门找到所有用户
	FindUserByDept(departid string) (us []*FlowUser, err error)
	//根据角色找到所有用户
	FindUserByRole(role string) (us []*FlowUser, err error)
}
