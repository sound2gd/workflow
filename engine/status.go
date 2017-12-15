package engine

import (
	"errors"
	"github.com/jteeuwen/go-pkg-xmlx"
)

// Status 流程状态, 也就是去到的环节
type Status struct {
	Name     string             //状态的名字, 也就是流程的步骤名称StepName
	Sequence int                //状态的默认顺序
	Partici  *Participant       //状态的处理人定义
	Choices  map[string]*Choice //状态处理时给用户的选择项,如:同意;不同意;
	//EnterHandler FlowHandler        //进入状态的时的触发处理器
	//ExitHandler  FlowHandler        //退出状态的时的触发处理器
}

// NewStatus 根据xml流程定义创建节点对象
func NewStatus(n *xmlx.Node) (*Status, error) {
	s := &Status{}
	s.Name = n.S("", "name")
	s.Sequence = n.I("", "sequence")
	if pn := n.SelectNode("", "participant"); pn != nil {
		if p, err := NewParticipant(pn); err != nil {
			return nil, err
		} else {
			s.Partici = p
		}
	}
	if cc := n.SelectNode("", "choices"); cc != nil {
		cs := cc.SelectNodes("", "choice")
		s.Choices = make(map[string]*Choice)
		for index, cn := range cs {
			if c, err := NewChoice(cn); err != nil {
				return nil, err
			} else {
				c.Index = index
				s.Choices[c.Name] = c
			}
		}
	}
	//if enn := n.SelectNode("", "enter"); enn != nil {
	// TODO: 进入节点时,可以出发任务
	//}
	//if exn := n.SelectNode("", "exit"); exn != nil {
	// TODO: 退出节点时,可以出发任务
	//}
	return s, nil
}

// NextStatus 找到下一步
// choice: 选择, 同意/不同意
// appdata: 业务表单的数据.
func (s *Status) NextStatus(choice string, appdata string) (nextStep string, err error) {
	//如果没有定义选择项, 返回空字符
	if len(s.Choices) == 0 {
		return "", nil
	}
	//找到选择的结果
	if c, find := s.Choices[choice]; find {
		if next, err := c.Trans.NextStatus(appdata); err != nil {
			return "", err
		} else {
			return next, nil
		}
	} else {
		//如果没有找到选择项, 报错
		return "", errors.New("not found choice: " + choice)
	}
}

// OnEnter 进入这个状态时触发任务
func (s *Status) OnEnter(appdata string, fcase *FlowCase, itemid int32) (string, error) {
	//TODO: 未实现
	return "", nil
}

// OnExit 进入这个状态时触发任务
func (s *Status) OnExit(appdata string, fcase *FlowCase, itemid int32) (string, error) {
	//TODO: 未实现
	return "", nil
}

// IsFree 确定是否是自由选择人员的步骤
func (s *Status) IsFree() bool {
	if s.Partici != nil && s.Partici.ParticiType == ParticiTypeFree {
		return true
	}
	return false
}
