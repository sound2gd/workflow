package workflow

import (
	"errors"
	"github.com/jteeuwen/go-pkg-xmlx"
	"github.com/widuu/gojson"
)

type FlowHandler interface {
	GetName() string
	//GetParams() map[string]string
	Execute(appdata string, flowcase *FlowCase, itemid int32) (string, error)
}

//----------------------------------

type ApiHandler struct {
	Url    string   //api地址
	Params []*Param //参数
}

func New_ApiHandler(n *xmlx.Node) (*ApiHandler, error) {
	h := &ApiHandler{}
	h.Url = n.As("", "name")
	psn := n.SelectNodes("", "param")
	h.Params = make([]*Param, 0, len(psn))
	for _, pn := range psn {
		if p, err := New_Param(pn); err != nil {
			return nil, err
		} else {
			h.Params = append(h.Params, p)
		}
	}
	return h, nil
}

func (a *ApiHandler) GetName() string {
	return a.Url
}

//func (a *ApiHandler) GetParams() map[string]string {
//	return a.Params
//}

func (a *ApiHandler) Execute(appdata string, fc *FlowCase, itemid int32) (string, error) {
	data := make(map[string]string)
	data = gojson.Json(appdata).GetDataFirstLevel()

	//收集参数
	ps := make(map[string]string)
	for _, p := range a.Params {
		switch p.Resource {
		case Resource_Appdata:
			//todo:
			ps[p.Key] = data[p.Key]
		case Resource_Local:
			ps[p.Key] = p.Value
		case Resource_Flow:
			ps[p.Key] = ""
		}
	}
	//执行调用
	return "执行成功", nil
}

//----------------------------------
//参数取值的来源
const (
	Resource_Appdata uint32 = iota //从appdata取
	Resource_Local                 //从xml文件
	Resource_Flow                  //从流程数据
)

type Param struct {
	Resource uint32 //参数来源
	Key      string
	Value    string
}

//解析流程中配置的参数
func New_Param(n *xmlx.Node) (*Param, error) {
	p := &Param{}
	pt := n.As("", "resource")
	switch pt {
	case "app":
		p.Resource = Resource_Appdata
	case "local":
		p.Resource = Resource_Local
	case "flow":
		p.Resource = Resource_Flow
	default:
		return nil, errors.New("not supported resource")
	}
	p.Key = n.As("", "key")
	p.Value = n.As("", "value")
	return p, nil
}
