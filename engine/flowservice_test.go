package engine_test

import (
	"fmt"
	"testing"
)

func TestAddFlow(t *testing.T) {
	flowhelper := New_FLowHelper()

	workflow1 := &FlowInfo{uuid.New(), "busytest", "hgcflow", "<xml>hgc test</xml>", 0}
	_, err1 := flowhelper.AddFlow(workflow1, "1007807")
	if err1 != nil {
		t.Error("add flow failed! detail info :", err1)
	}
}

func TestUpdateFlow(t *testing.T) {
	flowhelper := New_FLowHelper()

	workflow1 := &FlowInfo{"0c20d144-6b38-4001-adcb-940cb76ebcaa", "busytestchanged", "hgcflowchanged", "<xml>hgc test changed</xml>", 0}
	_, err1 := flowhelper.UpdateFlow(workflow1, "1007807")
	if err1 != nil {
		t.Error("update flow failed! detail info :", err1)
	}
}

func TestDeleteFlow(t *testing.T) {
	flowhelper := New_FLowHelper()

	_, err1 := flowhelper.DeleteFlow("0c20d144-6b38-4001-adcb-940cb76ebcaa", "1007807")
	if err1 != nil {
		t.Error("update flow failed! detail info :", err1)
	}
}

func TestGetMyAgent(t *testing.T) {
	flowhelper := New_FLowHelper()

	myagent, err1 := flowhelper.GetMyAgent("合同", "611987", "1007807", 1, 10)
	if err1 != nil {
		t.Error("get myagent failed! detail info :", err1)
	}
	fmt.Println(myagent)
}

func TestGetMyAffair(t *testing.T) {
	flowhelper := New_FLowHelper()

	myaffair, err1 := flowhelper.GetMyAffair("1", "1", "611987", "1007807", 1, 10)
	if err1 != nil {
		t.Error("get myaffair failed! detail info :", err1)
	}
	fmt.Println(myaffair)
}

func TestGetWorkFlow(t *testing.T) {
	flowhelper := New_FLowHelper()

	workflowlist, err1 := flowhelper.GetWorkFlowList("", "", "1007807", 1, 10)
	if err1 != nil {
		t.Error("get workflowinfo failed! detail info :", err1)
	}
	fmt.Println(workflowlist)
}

func TestGetWorkFlow1(t *testing.T) {
	flowhelper := New_FLowHelper()
	wf, _, err := flowhelper.GetWorkflowForTest("jdbc:postgresql://211.155.27.215:5432/xw_dl_1007807;userid=crm;password=crm")
	if err != nil {
		t.Error("get workflow failed! , detail info :", err)
	}

	flow, err1 := flowhelper.GetWorkFlowDetail("2413f773-4255-4f6a-b514-c69bfc4e4ca6", wf)
	if err1 != nil {
		t.Error("get workflow detail failed! detail info :", err1)
	}
	fmt.Println(flow)
}

func TestGetWorkFlow2(t *testing.T) {
	flowhelper := New_FLowHelper()
	wf, _, err := flowhelper.GetWorkflowForTest("jdbc:postgresql://211.155.27.215:5432/xw_dl_1007807;userid=crm;password=crm")
	if err != nil {
		t.Error("get workflow failed! , detail info :", err)
	}

	flow, err1 := flowhelper.GetWorkFlowDetail("", wf)
	if err1 != nil {
		t.Error("get workflow detail failed! detail info :", err1)
	}
	fmt.Println(flow)
}

func TestAddAffair(t *testing.T) {
	flowhelper := New_FLowHelper()
	wf, _, err := flowhelper.GetWorkflowForTest("jdbc:postgresql://211.155.27.215:5432/xw_dl_1007807;userid=crm;password=crm")
	if err != nil {
		t.Error("get workflow failed! , detail info :", err)
	}

	appdata := make(map[string]string)
	appdata["days"] = "three days"
	result, err1 := flowhelper.AddAffair("130fea06-b14b-4f57-bbc3-a7bb094f699f", "黄国臣", "644361", appdata, wf)
	if err1 != nil {
		t.Error("add affair failed! detail info :", err1)
	}
}
