package engine

import (
	"errors"
	"github.com/jteeuwen/go-pkg-xmlx"
)

// Transition 迁移, 对应流程图上节点之间的线条
type Transition struct {
	TrueTo     string       //如果有condition节点, 条件运算结果为true时去到的目标步骤名称
	FalseTo    string       //如果有condition节点, 条件运算结果为false时去掉的目标步骤名称
	Conditions []*Condition //条件集合
}

// NewTransition 根据xml节点信息创建迁移对象.
func NewTransition(n *xmlx.Node) (*Transition, error) {
	t := &Transition{}
	t.TrueTo = n.As("", "trueto")
	t.FalseTo = n.As("", "falseto")
	//如没有condition节点, 默认选择trueto
	if cc := n.SelectNode("", "conditions"); cc != nil {
		ca := cc.SelectNodes("", "condition")
		t.Conditions = make([]*Condition, len(ca))
		for i, cn := range ca {
			if c, err := NewCondition(cn); err != nil {
				return nil, err
			} else {
				t.Conditions[i] = c
			}
		}
	}
	return t, nil
}

// NextStatus 根据appdata和流程节点定义, 计算出下一步的步骤名称
// nextstep 返回下一步去到的步骤
func (t *Transition) NextStatus(appdata string) (nextstep string, err error) {
	c := len(t.Conditions)
	if c == 0 { //如没有condition节点, 默认选择trueto
		return t.TrueTo, nil
	}
	fr := true
	for _, cd := range t.Conditions {
		if result, err := cd.Eval(appdata); err != nil {
			return "", errors.New("condition eval error")
		} else {
			if cd.Logic == LogicAnd {
				fr = fr && result
			} else if cd.Logic == LogicOr {
				fr = fr || result
			} else {
				return "", errors.New("not supported condition logic operator")
			}
		}
	}
	if fr {
		return t.TrueTo, nil
	}
	return t.FalseTo, nil
}
