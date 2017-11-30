package workflow

import (
	"errors"
	//"fmt"
	"github.com/jteeuwen/go-pkg-xmlx"
)

//流程状态, 也就是去到的环节
type Status struct {
	Name         string             //状态的名字
	Sequence     int                //状态的默认顺序
	Partici      *Participant       //状态的处理人定义
	Choices      map[string]*Choice //状态处理时给用户的选择项,如:同意;不同意;
	EnterHandler FlowHandler        //进入状态的时的触发处理器
	ExitHandler  FlowHandler        //退出状态的时的触发处理器
}

type ByIndex map[int]*Choice

func (s ByIndex) Len() int {
	return len(s)
}
func (s ByIndex) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByIndex) Less(i, j int) bool {
	return s[i].Index < s[j].Index
}

func New_Status(n *xmlx.Node) (*Status, error) {
	s := &Status{}
	s.Name = n.S("", "name")
	s.Sequence = n.I("", "sequence")
	//fmt.Println(s)
	if pn := n.SelectNode("", "participant"); pn != nil {
		if p, err := New_Participant(pn); err != nil {
			return nil, err
		} else {
			s.Partici = p
		}
	}
	//fmt.Println("--1--")
	if cc := n.SelectNode("", "choices"); cc != nil {
		cs := cc.SelectNodes("", "choice")
		s.Choices = make(map[string]*Choice)
		for index, cn := range cs {
			if c, err := New_Choice(cn); err != nil {
				return nil, err
			} else {
				c.Index = index
				log.Info("New_Status.choice", c)
				s.Choices[c.Name] = c
			}
		}
	}
	//fmt.Println("--2--")
	if enn := n.SelectNode("", "enter"); enn != nil {
		//todo: 目前只支持apihandler
		if eh, err := New_ApiHandler(enn); err != nil {
			return nil, err
		} else {
			s.EnterHandler = eh
		}
	}
	if exn := n.SelectNode("", "exit"); exn != nil {
		//todo: 目前只支持apihandler
		if eh, err := New_ApiHandler(exn); err != nil {
			return nil, err
		} else {
			s.ExitHandler = eh
		}
	}
	//fmt.Println("--3--")
	return s, nil
}

//找到下一步
func (s *Status) NextStatus(choice string, appdata string) (string, error) {
	// fmt.Println(choice)
	// fmt.Println("NextStatus")
	// fmt.Println(s)
	// fmt.Println(appdata)
	//如果没有定义选择项, 返回空字符
	if len(s.Choices) == 0 {
		return "", nil
	}
	// fmt.Println("status name: ", s.Name)
	// fmt.Println("choice: ", choice)
	//找到选择的结果
	if c, find := s.Choices[choice]; find {
		// fmt.Println(c)
		// fmt.Println(find)
		if next, err := c.Trans.NextStatus(appdata); err != nil {
			//fmt.Println("next err: ", err.Error())
			return "", err
		} else {
			//fmt.Println("next: ", next)
			return next, nil
		}
	} else {
		//如果没有找到选择项, 报错
		return "", errors.New("not found choice: " + choice)
	}
}

//进入这个状态时出发的任务
func (s *Status) OnEnter(appdata string, fcase *FlowCase, itemid int32) (string, error) {
	if s.EnterHandler == nil {
		return "", nil
	}
	return s.EnterHandler.Execute(appdata, fcase, itemid)
}

//进入这个状态时出发的任务
func (s *Status) OnExit(appdata string, fcase *FlowCase, itemid int32) (string, error) {
	if s.ExitHandler == nil {
		return "", nil
	}
	return s.ExitHandler.Execute(appdata, fcase, itemid)
}

//确定是否是自由选择人员的步骤
func (s *Status) IsFree() bool {
	if s.Partici != nil && s.Partici.ParticiType == ParticiType_Free {
		return true
	}
	return false
}
