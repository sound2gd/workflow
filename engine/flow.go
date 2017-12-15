package engine

import (
	"errors"
	"github.com/jteeuwen/go-pkg-xmlx"
)

// Flow 流程对象的定义
type Flow struct {
	FlowID     string
	FlowName   string
	FlowStatus map[string]*Status
	VersionNo  int32
}

// NewFlow 根据xml创建一个新的流程对象
func NewFlow(flowid, flowname, flowxml string, versionno int32) (*Flow, error) {
	flow := &Flow{
		FlowID:     flowid,
		FlowName:   flowname,
		FlowStatus: make(map[string]*Status),
		VersionNo:  versionno,
	}
	err := flow.parseFlow(flowxml)
	if err != nil {
		return nil, errors.New("new flow parse xml error ")
	}
	return flow, nil
}

// StatuCount 状态是数量?
func (f *Flow) StatuCount() int {
	return len(f.FlowStatus)
}

//解析xml数据
func (f *Flow) parseFlow(flowxml string) error {
	doc := xmlx.New()
	if err := doc.LoadString(flowxml, nil); err != nil {
		return err
	}
	wfn := doc.SelectNode("", "workflow")
	fsd := wfn.SelectNode("", "flowstatus")
	sds := fsd.SelectNodes("", "status")
	for _, sn := range sds {
		st, err := NewStatus(sn)
		if err != nil {
			return err
		}
		f.FlowStatus[st.Name] = st
	}
	return nil
}
