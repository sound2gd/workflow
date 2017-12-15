package engine

import (
	"errors"
	"github.com/jteeuwen/go-pkg-xmlx"
	"github.com/pobearm/workflow/util"
	"strings"
)

// 流程参与人的寻找方案
const (
	ParitciTypeUserid    uint32 = iota //指定用户
	ParitciTypeRoledept                //部门+角色
	ParticiTypeDept                    //部门
	ParticiTypeRole                    //角色
	ParticiTypeFree                    //自由选择
	ParticiTypeAnyUserid               //通讯录所有人的人
	ParticiTypeCreator                 //流程发起人
	ParticiTypeStepUsers               //参与过流程的人
)

// 部门定义方案
const (
	DeptGetCreator        uint32 = iota //流程发起人所在部门
	DeptGetStepUser                     //步骤处理人所在部门
	DeptGetDeptid                       //指定部门
	DeptGetStepUserParent               //步骤处理人所在部门的父级部门
)

// Participant 流程步骤参与人
type Participant struct {
	//寻找步骤的处理人,userid: 指定的用户id为处理人,roledept:使用部门+角色寻找
	//dept:部门内的人;  role: 角色的所有人; free: 自由选人;
	//creator:流程发起人; stepusers:参与过流程的人;
	ParticiType uint32
	Userid      string //指定的用户处理
	Role        string //指定的角色
	Deptid      string //特定的部门, DType="deptid"时生效
	//如何获取部门,
	//creator: 流程的创建人所在部门,stepuser: 流程步骤的处理人人所在部门,
	//deptid: 特定的部门, stepuserparent: 流程步骤的处理人人所在部门的父部门
	DeptGetType uint32
}

// NewParticipant 根据xml定义,构造流程参与人
func NewParticipant(n *xmlx.Node) (*Participant, error) {
	p := &Participant{}
	pt := n.As("", "ptype")
	switch pt {
	case "userid":
		p.ParticiType = ParitciTypeUserid
	case "roledept":
		p.ParticiType = ParitciTypeRoledept
	case "dept":
		p.ParticiType = ParticiTypeDept
	case "role":
		p.ParticiType = ParticiTypeRole
	case "free":
		p.ParticiType = ParticiTypeFree
	case "anyuserid":
		p.ParticiType = ParticiTypeAnyUserid
	case "creator":
		p.ParticiType = ParticiTypeCreator
	case "stepusers":
		p.ParticiType = ParticiTypeStepUsers
	default:
		return nil, errors.New("not supported ptype")
	}
	if un := n.SelectNode("", "userid"); un != nil {
		p.Userid = un.GetValue()
	}
	if un := n.SelectNode("", "role"); un != nil {
		p.Role = un.GetValue()
	}
	if dn := n.SelectNode("", "dept"); dn != nil {
		dt := dn.As("", "dtype")
		switch dt {
		case "creator":
			p.DeptGetType = DeptGetCreator
		case "stepuser":
			p.DeptGetType = DeptGetStepUser
		case "deptid":
			p.DeptGetType = DeptGetDeptid
			p.Deptid = dn.S("", "deptid")
		case "stepuserparent":
			p.DeptGetType = DeptGetStepUserParent
		default:
			return nil, errors.New("partici dtype not supported")
		}
	}
	return p, nil
}

// FindUser 使用组织服务,
func (p *Participant) FindUser(opd OrgService, fcase *FlowCase) ([]*FlowUser, error) {
	//fmt.Println("particiType: ", p.ParticiType)
	if p.ParticiType == ParitciTypeUserid {
		//直接返回设定的userid
		return opd.GetUser(p.Userid)

	} else if p.ParticiType == ParitciTypeRoledept {

		switch p.DeptGetType {
		case DeptGetDeptid: //指定部门,的角色的人.
			return opd.FindUser(p.Role, p.Deptid)
		case DeptGetCreator: //流程创建人所在的部门.
			if dpid, err := opd.FindUserDept(fcase.CaseInfo.CreatorID); err != nil {
				return nil, err
			} else {
				return opd.FindUser(p.Role, dpid)
			}
		case DeptGetStepUser: //上一个步骤处理人所在部门
			caseItemCount := len(fcase.CaseItems)
			lastUserid := fcase.CaseItems[int32(caseItemCount-1)].HandleUserid
			if dpid, err := opd.FindUserDept(lastUserid); err != nil {
				return nil, err
			} else {
				return opd.FindUser(p.Role, dpid)
			}
		case DeptGetStepUserParent: //上一个步骤处理人所在部门的父部门
			caseItemCount := len(fcase.CaseItems)
			lastUserid := fcase.CaseItems[int32(caseItemCount-1)].HandleUserid
			if ppid, err := opd.FindUserParentDept(lastUserid); err != nil {
				return nil, err
			} else {
				return opd.FindUser(p.Role, ppid)
			}
		default:
			return nil, errors.New("unsupported Participant DeptGetType")
		}

	} else if p.ParticiType == ParticiTypeDept {

		switch p.DeptGetType {
		case DeptGetDeptid: //指定部门,的角色的人.
			return opd.FindUserByDept(p.Deptid)
		case DeptGetCreator: //流程创建人所在的部门.
			if dpid, err := opd.FindUserDept(fcase.CaseInfo.CreatorID); err != nil {
				return nil, err
			} else {
				return opd.FindUserByDept(dpid)
			}
		case DeptGetStepUser: //上一个步骤处理人所在部门
			caseItemCount := len(fcase.CaseItems)
			lastUserid := fcase.CaseItems[int32(caseItemCount-1)].HandleUserid
			if dpid, err := opd.FindUserDept(lastUserid); err != nil {
				return nil, err
			} else {
				return opd.FindUserByDept(dpid)
			}
		case DeptGetStepUserParent: //上一个步骤处理人所在部门的父部门
			caseItemCount := len(fcase.CaseItems)
			lastUserid := fcase.CaseItems[int32(caseItemCount-1)].HandleUserid
			if ppid, err := opd.FindUserParentDept(lastUserid); err != nil {
				return nil, err
			} else {
				return opd.FindUserByDept(ppid)
			}
		default:
			return nil, errors.New("unsupported Participant DeptGetType")
		}

	} else if p.ParticiType == ParticiTypeRole {

		if u, err := opd.FindUserByRole(p.Role); err != nil {
			return nil, err
		} else {
			return u, nil
		}

	} else if p.ParticiType == ParticiTypeFree {

		us := make([]*FlowUser, 0, 0)
		//自由选择, 返回空的用户列表
		return us, nil

	} else if p.ParticiType == ParticiTypeAnyUserid {

		us := make([]*FlowUser, 0, 0)
		return us, nil

	} else if p.ParticiType == ParticiTypeCreator {

		//流程创建人
		return opd.GetUser(fcase.CaseInfo.CreatorID)

	} else if p.ParticiType == ParticiTypeStepUsers {

		//参与过流程的人
		users := make([]string, 0, len(fcase.CaseItems))
		for _, ci := range fcase.CaseItems {
			if !util.StringInSlice(ci.HandleUserid, users) {
				users = append(users, ci.HandleUserid)
			}
		}
		uids := strings.Join(users, ",")
		return opd.GetUser(uids)

	} else {
		return nil, errors.New("unsupported particiType")
	}
}
