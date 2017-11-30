package workflow

import (
	"errors"
	//"fmt"
	"github.com/jteeuwen/go-pkg-xmlx"
	"strings"
)

const (
	ParitciType_Userid    uint32 = iota //指定用户
	ParitciType_Roledept                //部门+角色
	ParticiType_Dept                    //部门
	ParticiType_Role                    //角色
	ParticiType_Free                    //自由选择
	ParticiType_AnyUserid               //通讯录所有人的人
	ParticiType_Creator                 //流程发起人
	ParticiType_StepUsers               //参与过流程的人
)

const (
	DeptGet_Creator        uint32 = iota //流程发起人所在部门
	DeptGet_StepUser                     //步骤处理人所在部门
	DeptGet_Deptid                       //指定部门
	DeptGet_StepUserParent               //步骤处理人所在部门的父级部门
)

type FlowUser struct {
	Userid    string `json:"userid,omitempty"`
	UserName  string `json:"username,omitempty"`
	HeadPhoto string `json:"headphoto,omitempty"`
}

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

func New_Participant(n *xmlx.Node) (*Participant, error) {
	p := &Participant{}
	pt := n.As("", "ptype")
	switch pt {
	case "userid":
		p.ParticiType = ParitciType_Userid
	case "roledept":
		p.ParticiType = ParitciType_Roledept
	case "dept":
		p.ParticiType = ParticiType_Dept
	case "role":
		p.ParticiType = ParticiType_Role
	case "free":
		p.ParticiType = ParticiType_Free
	case "anyuserid":
		p.ParticiType = ParticiType_AnyUserid
	case "creator":
		p.ParticiType = ParticiType_Creator
	case "stepusers":
		p.ParticiType = ParticiType_StepUsers
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
			p.DeptGetType = DeptGet_Creator
		case "stepuser":
			p.DeptGetType = DeptGet_StepUser
		case "deptid":
			p.DeptGetType = DeptGet_Deptid
			p.Deptid = dn.S("", "deptid")
		case "stepuserparent":
			p.DeptGetType = DeptGet_StepUserParent
		default:
			return nil, errors.New("partici dtype not supported")
		}
	}
	return p, nil
}

func (p *Participant) FindUser(opd OrgProvider, fcase *FlowCase) ([]*FlowUser, error) {
	//fmt.Println("particiType: ", p.ParticiType)
	if p.ParticiType == ParitciType_Userid {
		//直接返回设定的userid
		return opd.GetUser(p.Userid)

	} else if p.ParticiType == ParitciType_Roledept {

		switch p.DeptGetType {
		case DeptGet_Deptid: //指定部门,的角色的人.
			return opd.FindUser(p.Role, p.Deptid)
		case DeptGet_Creator: //流程创建人所在的部门.
			if dpid, err := opd.FindUserDept(fcase.CaseInfo.CreatorId); err != nil {
				return nil, err
			} else {
				return opd.FindUser(p.Role, dpid)
			}
		case DeptGet_StepUser: //上一个步骤处理人所在部门
			caseItemCount := len(fcase.CaseItems)
			lastUserid := fcase.CaseItems[int32(caseItemCount-1)].HandleUserid
			if dpid, err := opd.FindUserDept(lastUserid); err != nil {
				return nil, err
			} else {
				return opd.FindUser(p.Role, dpid)
			}
		case DeptGet_StepUserParent: //上一个步骤处理人所在部门的父部门
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

	} else if p.ParticiType == ParticiType_Dept {

		switch p.DeptGetType {
		case DeptGet_Deptid: //指定部门,的角色的人.
			return opd.FindUserByDept(p.Deptid)
		case DeptGet_Creator: //流程创建人所在的部门.
			if dpid, err := opd.FindUserDept(fcase.CaseInfo.CreatorId); err != nil {
				return nil, err
			} else {
				return opd.FindUserByDept(dpid)
			}
		case DeptGet_StepUser: //上一个步骤处理人所在部门
			caseItemCount := len(fcase.CaseItems)
			lastUserid := fcase.CaseItems[int32(caseItemCount-1)].HandleUserid
			if dpid, err := opd.FindUserDept(lastUserid); err != nil {
				return nil, err
			} else {
				return opd.FindUserByDept(dpid)
			}
		case DeptGet_StepUserParent: //上一个步骤处理人所在部门的父部门
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

	} else if p.ParticiType == ParticiType_Role {

		if u, err := opd.FindUserByRole(p.Role); err != nil {
			return nil, err
		} else {
			return u, nil
		}

	} else if p.ParticiType == ParticiType_Free {

		us := make([]*FlowUser, 0, 0)
		//自由选择, 返回空的用户列表
		return us, nil

	} else if p.ParticiType == ParticiType_AnyUserid {

		us := make([]*FlowUser, 0, 0)
		return us, nil

	} else if p.ParticiType == ParticiType_Creator {

		//流程创建人
		return opd.GetUser(fcase.CaseInfo.CreatorId)

	} else if p.ParticiType == ParticiType_StepUsers {

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
