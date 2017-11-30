package workflow

import "github.com/jteeuwen/go-pkg-xmlx"

//流程步骤的选择项
type Choice struct {
	Index     int
	Name      string            //选择项的值
	Trans     *Transition       //选项对应的状态变化
	AllEdit   bool              //允许编辑所有字段
	DataItems []*ChoiceDataItem //审批时需要填写的数据
}

type ChoiceResp struct {
	Name      string            //选择项的值
	DataItems []*ChoiceDataItem //审批时需要填写的数据
}

func New_Choice(n *xmlx.Node) (*Choice, error) {
	c := &Choice{}
	c.Name = n.As("", "name")
	//fmt.Println(c)
	//迁移
	tn := n.SelectNode("", "transition")
	if t, err := New_Transition(tn); err != nil {
		return nil, err
	} else {
		c.Trans = t
	}
	//数据填写项
	if dns := n.SelectNode("", "dataitems"); dns != nil {
		c.AllEdit = dns.Ab("", "alledit")
		items := dns.SelectNodes("", "item")
		c.DataItems = make([]*ChoiceDataItem, len(items))
		for i, in := range items {
			if d, err := New_ChoiceDataItem(in); err != nil {
				return nil, err
			} else {
				c.DataItems[i] = d
			}
		}
	}
	return c, nil
}
