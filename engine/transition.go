package workflow

import (
	"errors"
	// "fmt"
	"github.com/jteeuwen/go-pkg-xmlx"
)

//迁移
type Transition struct {
	TrueTo     string       //如果有condition节点, 根据条件结果选择trueto或者falseto-
	FalseTo    string       //如果有condition节点, 根据条件结果选择trueto或者falseto-
	Conditions []*Condition //条件集合
}

func New_Transition(n *xmlx.Node) (*Transition, error) {
	t := &Transition{}
	t.TrueTo = n.As("", "trueto")
	t.FalseTo = n.As("", "falseto")
	//如没有condition节点, 默认选择trueto
	if cc := n.SelectNode("", "conditions"); cc != nil {
		ca := cc.SelectNodes("", "condition")
		t.Conditions = make([]*Condition, len(ca))
		for i, cn := range ca {
			if c, err := New_Condition(cn); err != nil {
				return nil, err
			} else {
				t.Conditions[i] = c
			}
		}
	}
	return t, nil
}

//计算出下一步的状态名称
func (t *Transition) NextStatus(appdata string) (string, error) {
	//fmt.Println("Transition  NextStatus")
	//fmt.Println(t)
	c := len(t.Conditions)
	if c == 0 { //如没有condition节点, 默认选择trueto
		//fmt.Println("condition len:", c)
		return t.TrueTo, nil
	} else {
		fr := true
		for _, cd := range t.Conditions {
			if r, err := cd.Eval(appdata); err != nil {
				log.Error("transition.NextStatus", err.Error())
				return "", errors.New("condition eval error")
			} else {
				if cd.Logic == Logic_And {
					fr = fr && r
				} else if cd.Logic == Logic_Or {
					fr = fr || r
				} else {
					return "", errors.New("not supported condition logic operator")
				}
			}
		}
		log.Debug("workflow.transition.NextStatus", "condi eval result: ")
		if fr {
			return t.TrueTo, nil
		} else {
			return t.FalseTo, nil
		}
	}
}
