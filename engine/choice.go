package engine

import "github.com/jteeuwen/go-pkg-xmlx"

// Choice 流程步骤的选择项, 如: 同意 / 不同意
type Choice struct {
	Index int         //选项的序号
	Name  string      //选择项的值
	Trans *Transition //选项对应的状态变化
}

// NewChoice 从xml的描述中,创建choice对象
func NewChoice(n *xmlx.Node) (*Choice, error) {
	c := &Choice{}
	c.Name = n.As("", "name")
	//迁移的描述节点
	tn := n.SelectNode("", "transition")
	//创建迁移对象
	if t, err := NewTransition(tn); err != nil {
		return nil, err
	} else {
		c.Trans = t
	}
	return c, nil
}

// ByIndex 实现choice按index排序?
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
