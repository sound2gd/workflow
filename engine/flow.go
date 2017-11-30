package workflow

import (
	"github.com/jteeuwen/go-pkg-xmlx"
	log "xtion.net/mcrm/logger"
)

//流程对象的定义
type Flow struct {
	FlowId     string
	FlowName   string
	FlowStatus map[string]*Status
	VersionNo  int32
}

func New_Flow(flowid, flowname, flowxml string, versionno int32) *Flow {
	flow := &Flow{
		FlowId:     flowid,
		FlowName:   flowname,
		FlowStatus: make(map[string]*Status),
		VersionNo:  versionno,
	}
	err := flow.parseFlow(flowxml)
	if err != nil {
		log.Info("workflow.New_Flow", err)
	}
	return flow
}

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
	//fmt.Println(sds)
	//fmt.Println(len(sds))
	for _, sn := range sds {
		st, err := New_Status(sn)
		if err != nil {
			return err
		}
		f.FlowStatus[st.Name] = st
	}
	return nil
}
