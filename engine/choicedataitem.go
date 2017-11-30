package workflow

import "github.com/jteeuwen/go-pkg-xmlx"

//流程步骤的选择项
type ChoiceDataItem struct {
	Name string //审批时,需要填写的数据项, 对应于[业务对象]的属性名称
	Must bool   //是否必填
}

func New_ChoiceDataItem(n *xmlx.Node) (*ChoiceDataItem, error) {
	c := &ChoiceDataItem{}
	c.Name = n.As("", "name")
	c.Must = n.Ab("", "must")
	return c, nil
}
