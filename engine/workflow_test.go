package main

import (
	"testing"

	"xtion.net/mcrm/workflow"
)

const (
	conn_str = "jdbc:postgresql://211.155.27.215:5432/xw_dl_1007807;userid=crm;password=crm"
)

func Test_NewFlow3(t *testing.T) {
	userid := "675608"
	username := "明亮9"
	wfid := "954a0d8b-554b-4a7e-9423-dee0dece4c25"
	wf, err := workflow.New_Workflow(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	appdata := make(map[string]string)
	appdata["amount"] = "15001"
	appdata["product"] = "1001"
	caseid, errc := wf.CreateWorkflow(wfid, appdata, userid, username)
	if errc != nil {
		t.Error(errc)
	}
	t.Log("caseid: ", caseid)
	//instep := ""
	//---------------------------------------------------
	nsif0, err0 := wf.PreRun(0, "")
	if err0 != nil {
		t.Error(err0)
		return
	}
	showNsif(nsif0, t)
	user0 := &workflow.FlowUser{Userid: "647749", UserName: "熊利祥0"}
	if instep, err := wf.Run(0, "", "mark0", user0); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("0:" + instep + " ---------------------------------------------")
	}
	nsif1, err1 := wf.PreRun(1, "同意")
	if err1 != nil {
		t.Error(err1)
		return
	}
	showNsif(nsif1, t)
	user1 := &workflow.FlowUser{Userid: "647748", UserName: "熊利祥1"}
	if instep, err := wf.Run(1, "同意", "mark1", user1); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("1: " + instep + " ---------------------------------------------")
	}
	nsif2, err2 := wf.PreRun(2, "同意")
	if err2 != nil {
		t.Error(err2)
		return
	}
	showNsif(nsif2, t)
	user2 := &workflow.FlowUser{Userid: "647748", UserName: "熊利祥2"}
	if instep, err := wf.Run(2, "同意", "mark1", user2); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("2: " + instep + " ---------------------------------------------")
	}
	nsif3, err3 := wf.PreRun(3, "不同意")
	if err3 != nil {
		t.Error(err3)
		return
	}
	showNsif(nsif3, t)
	user3 := &workflow.FlowUser{Userid: "647748", UserName: "熊利祥3"}
	if instep, err := wf.Run(3, "不同意", "mark1", user3); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("3: " + instep + " ---------------------------------------------")
	}
	nsif4, err4 := wf.PreRun(4, "")
	if err4 != nil {
		t.Error(err4)
		return
	}
	showNsif(nsif4, t)
	user4 := &workflow.FlowUser{Userid: "647748", UserName: "熊利祥3"}
	if instep, err := wf.Run(4, "", "mark1", user4); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("4: " + instep + " ---------------------------------------------")
	}
	nsif5, err5 := wf.PreRun(5, "同意")
	if err5 != nil {
		t.Error(err5)
		return
	}
	showNsif(nsif5, t)
	if instep, err := wf.Run(5, "同意", "mark1", nil); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("5: " + instep + " ---------------------------------------------")
	}
}

func Test_NewFlow22(t *testing.T) {
	userid := "675608"
	username := "明亮9"
	wfid := "10ca18b9-7278-4637-b6ed-76a26a106997"
	wf, err := workflow.New_Workflow(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	appdata := make(map[string]string)
	appdata["amount"] = "15001"
	appdata["product"] = "1001"
	caseid, errc := wf.CreateWorkflow(wfid, appdata, userid, username)
	if errc != nil {
		t.Error(errc)
	}
	t.Log("caseid: ", caseid)
	//instep := ""
	//---------------------------------------------------
	nsif0, err0 := wf.PreRun(0, "")
	if err0 != nil {
		t.Error(err0)
		return
	}
	showNsif(nsif0, t)
	if instep, err := wf.Run(0, "", "mark0", nsif0.Users[0]); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("0: " + instep + "---------------------------------------------")
	}
	//------------------------------------------------
	nsif1, err1 := wf.PreRun(1, "退回")
	if err1 != nil {
		t.Error(err1)
		return
	}
	showNsif(nsif1, t)
	if instep, err := wf.Run(1, "退回", "mark1", nsif1.Users[0]); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("1: " + instep + "---------------------------------------------")
	}
	nsif2, err2 := wf.PreRun(2, "")
	if err2 != nil {
		t.Error(err2)
		return
	}
	showNsif(nsif2, t)
	if instep, err := wf.Run(2, "", "mark2", nsif2.Users[0]); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("2: " + instep + "---------------------------------------------")
	}
	nsif3, err3 := wf.PreRun(3, "同意")
	if err3 != nil {
		t.Error(err3)
		return
	}
	showNsif(nsif3, t)
	if instep, err := wf.Run(3, "同意", "mark3", nsif3.Users[0]); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("3: " + instep + "---------------------------------------------")
	}
	nsif4, err4 := wf.PreRun(4, "不同意")
	if err4 != nil {
		t.Error(err4)
		return
	}
	showNsif(nsif4, t)
	if instep, err := wf.Run(4, "不同意", "mark3", nil); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("4: " + instep + "---------------------------------------------")
	}
}

func Test_NewFlow2(t *testing.T) {
	userid := "675608"
	username := "明亮9"
	wfid := "10ca18b9-7278-4637-b6ed-76a26a106997"
	wf, err := workflow.New_Workflow(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	appdata := make(map[string]string)
	appdata["amount"] = "15001"
	appdata["product"] = "1001"
	caseid, errc := wf.CreateWorkflow(wfid, appdata, userid, username)
	if errc != nil {
		t.Error(errc)
	}
	t.Log("caseid: ", caseid)
	//instep := ""
	//---------------------------------------------------
	nsif0, err0 := wf.PreRun(0, "")
	if err0 != nil {
		t.Error(err0)
		return
	}
	showNsif(nsif0, t)
	if instep, err := wf.Run(0, "", "mark0", nsif0.Users[0]); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("0: " + instep + "---------------------------------------------")
	}
	//------------------------------------------------
	nsif1, err1 := wf.PreRun(1, "同意")
	if err1 != nil {
		t.Error(err1)
		return
	}
	showNsif(nsif1, t)
	if instep, err := wf.Run(1, "同意", "mark1", nsif1.Users[0]); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("1: " + instep + "---------------------------------------------")
	}
	nsif2, err2 := wf.PreRun(2, "同意")
	if err2 != nil {
		t.Error(err2)
		return
	}
	showNsif(nsif2, t)
	if instep, err := wf.Run(2, "同意", "mark2", nil); err != nil {
		t.Error(err)
		return
	} else {
		t.Log("2: " + instep + "---------------------------------------------")
	}
}

func showNsif(nsif *workflow.NextStatuInfo, t *testing.T) {
	t.Log("next step is: ", nsif.StepName)
	t.Log("is free: ", nsif.IsFree)
	if !nsif.IsFree {
		t.Log("user num: ", len(nsif.Users))
		for i := 0; i < len(nsif.Users); i++ {
			t.Log("user ", i, "is: ", nsif.Users[i])
		}
	}
}

func Test_NewFlow(t *testing.T) {
	userid := "675608"
	username := "明亮9"
	wfid := "aa9cadc9-35d0-4888-9852-b9395b55ee55"

	wf, err := workflow.New_Workflow(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	appdata := make(map[string]string)
	appdata["amount"] = "15001"
	appdata["product"] = "1001"
	caseid, errc := wf.CreateWorkflow(wfid, appdata, userid, username)
	if errc != nil {
		t.Error(errc)
	}
	t.Log("caseid: ", caseid)
	//instep := ""
	//---------------------------------------------------
	nsif0, err0 := wf.PreRun(0, "")
	if err0 != nil {
		t.Error(err0)
		return
	}
	t.Log("next step 1: ", nsif0.StepName)
	t.Log("user num: ", len(nsif0.Users))
	t.Log("next user: ", nsif0.Users[0])

	if instep, err := wf.Run(0, "", "mark0", nsif0.Users[0]); err != nil {
		t.Error(err)
		return
	} else {
		//t.Logf("%+v \n", wf.Fcase.CaseInfo)
		//for _, ci := range wf.Fcase.CaseItems {
		//	t.Logf("%#v \n", ci)
		//}
		t.Log(instep + "---------------------------------------------")
	}
	//---------------------------------------------------
	nsif1, err1 := wf.PreRun(1, "同意")
	if err1 != nil {
		t.Error(err1)
		return
	}
	t.Log("next step 2: ", nsif1.StepName)
	t.Log("user num: ", len(nsif1.Users))
	t.Log("next user: ", nsif1.Users[0])

	if instep, err := wf.Run(1, "同意", "mark1", nsif1.Users[0]); err != nil {
		t.Error(err)
		return
	} else {
		//t.Logf("%+v \n", wf.Fcase.CaseInfo)
		//for _, ci := range wf.Fcase.CaseItems {
		//	t.Logf("%#v \n", ci)
		//}
		t.Log(instep + "---------------------------------------------")
	}
	//---------------------------------------------------
	nsif2, err2 := wf.PreRun(2, "同意")
	if err2 != nil {
		t.Error(err2)
		return
	}
	t.Log("next step 3: ", nsif2.StepName)
	t.Log("user num: ", len(nsif2.Users))
	t.Log("next user: ", nsif2.Users[0])

	if instep, err := wf.Run(2, "同意", "mark2", nsif2.Users[0]); err != nil {
		t.Error(err)
		return
	} else {
		//t.Logf("%+v \n", wf.Fcase.CaseInfo)
		//for _, ci := range wf.Fcase.CaseItems {
		//	t.Logf("%#v \n", ci)
		//}
		t.Log(instep + "---------------------------------------------")
	}
	//---------------------------------------------------
	nsif3, err3 := wf.PreRun(3, "同意")
	if err3 != nil {
		t.Error(err3)
		return
	}
	t.Log("next step 4: ", nsif3.StepName)
	t.Log("user num: ", len(nsif3.Users))
	t.Log("next user: ", nsif3.Users[0])

	if instep, err := wf.Run(3, "同意", "mark3", nsif3.Users[0]); err != nil {
		t.Error(err)
		return
	} else {
		//t.Logf("%+v \n", wf.Fcase.CaseInfo)
		//for _, ci := range wf.Fcase.CaseItems {
		//	t.Logf("%#v \n", ci)
		//}
		t.Log(instep + "---------------------------------------------")
	}
	//---------------------------------------------------
	nsif4, err4 := wf.PreRun(4, "同意")
	if err4 != nil {
		t.Error(err4)
		return
	}
	t.Log("next step 5: ", nsif4.StepName)
	t.Log("user num: ", len(nsif4.Users))
	t.Log("next user: ", nsif4.Users[0])

	if instep, err := wf.Run(4, "同意", "mark4", nsif4.Users[0]); err != nil {
		t.Error(err)
		return
	} else {
		//t.Logf("%+v \n", wf.Fcase.CaseInfo)
		//for _, ci := range wf.Fcase.CaseItems {
		//	t.Logf("%#v \n", ci)
		//}
		t.Log(instep + "---------------------------------------------")
	}
	//---------------------------------------------------
	nsif5, err5 := wf.PreRun(5, "同意")
	if err5 != nil {
		t.Error(err5)
		return
	}
	t.Log("next step 6: ", nsif5.StepName)
	t.Log("user num: ", len(nsif5.Users))
	if len(nsif5.Users) > 0 {
		t.Log("next user: ", nsif5.Users[0])
	}

	if instep, err := wf.Run(5, "同意", "mark5", nil); err != nil {
		t.Error(err)
		return
	} else {
		t.Logf("%+v \n", wf.Fcase.CaseInfo)
		for _, ci := range wf.Fcase.CaseItems {
			t.Logf("%#v \n", ci)
		}
		t.Log(instep + "---------------------------------------------")
	}
	//---------------------------------------------------
}

func Test_GetTodoCases1(t *testing.T) {
	wh, err := workflow.New_FLowHelper(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	//查询条件
	flowname := ""
	userid := "647749"
	cl, err := wh.GetTodoCases(flowname, userid, 1, 10)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%#v", cl)
}

func Test_GetTodoCases2(t *testing.T) {
	wh, err := workflow.New_FLowHelper(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	//查询条件
	flowname := "测试"
	userid := "647749"
	cl, err := wh.GetTodoCases(flowname, userid, 1, 10)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%#v", cl)
}

func Test_GetWorkFlows1(t *testing.T) {
	wh, err := workflow.New_FLowHelper(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	fl, err := wh.GetWorkFlows("a,b`c'd;e--f", 1, 10)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(fl)
}

func Test_GetWorkFlows2(t *testing.T) {
	wh, err := workflow.New_FLowHelper(conn_str)
	if err != nil {
		t.Error(err)
	}
	fl, err := wh.GetWorkFlows("", 1, 10)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(fl)
}

func Test_GetCaseDetail1(t *testing.T) {
	wh, err := workflow.New_FLowHelper(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	caseid := "a7fd2c1b-c966-4493-a257-73fab468511f"
	fc, err := wh.GetCaseDetail(caseid)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%#v", fc)
}

func Test_GetCaseDetail2(t *testing.T) {
	wh, err := workflow.New_FLowHelper(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	caseid := "a57979a1-e63a-4f3c-87cb-2b5620f3fc17"
	fc, err := wh.GetCaseDetail(caseid)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%#v", fc)
}

func Test_AddCase1(t *testing.T) {
	userid := "675608"
	username := "明亮9"
	wfid := "aa9cadc9-35d0-4888-9852-b9395b55ee55"
	wh, err := workflow.New_FLowHelper(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	biz_1 := "12323232"
	biz_2 := "2222222222"
	appdata := make(map[string]string)
	appdata["amount"] = "15001"
	appdata["product"] = "1001"
	cs, err := wh.AddCase(wfid, username, userid, biz_1, biz_2, appdata)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%#v", cs)
}

func Test_AddCase2(t *testing.T) {
	userid := "675608"
	username := "明亮9"
	wfid := "aa9cadc9-35d0-4888-9852-b9395b55ee55"
	wh, err := workflow.New_FLowHelper(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	biz_1 := "12323232"
	biz_2 := ""
	appdata := make(map[string]string)
	appdata["amount"] = "15001"
	appdata["product"] = "1001"
	cs, err := wh.AddCase(wfid, username, userid, biz_1, biz_2, appdata)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%#v", cs)
}

func Test_CommitCase1(t *testing.T) {
	caseid := "0ed936f1-560f-45a1-8269-db4b5d540c37"
	choice := "同意"
	remark := "...dsfdsf..df."
	userid := "675608"
	username := "明亮9"
	itemid := int32(0)
	appdata := make(map[string]string)
	appdata["amount"] = "15001"
	appdata["product"] = "1001"
	flowuser := &workflow.FlowUser{
		Userid:   userid,
		UserName: username,
	}
	wh, err := workflow.New_FLowHelper(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	instep, err2 := wh.CommitCase(caseid, choice, remark, itemid, flowuser, appdata)
	if err2 != nil {
		t.Error(err2)
		return
	}
	t.Log(instep)
}

func Test_PreCommitCase1(t *testing.T) {
	caseid := "0ed936f1-560f-45a1-8269-db4b5d540c37"
	choice := "同意"
	itemid := int32(1)
	appdata := make(map[string]string)
	appdata["amount"] = "15001"
	appdata["product"] = "1001"
	wh, err := workflow.New_FLowHelper(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	nsif, err := wh.PreCommitCase(caseid, choice, itemid, appdata)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%#v", nsif)
}

func Test_PreCommitCase2(t *testing.T) {
	caseid := "0ed936f1-560f-45a1-8269-db4b5d540c37"
	choice := "不同意"
	itemid := int32(1)
	appdata := make(map[string]string)
	appdata["amount"] = "15001"
	appdata["product"] = "1001"
	wh, err := workflow.New_FLowHelper(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	nsif, err := wh.PreCommitCase(caseid, choice, itemid, appdata)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("%#v", nsif)
}

func Test_AbandonCase(t *testing.T) {
	caseid := "6c553ac7-fe92-4237-9a42-e925a63ab208"
	choice := "同意"
	remark := "我要作废"
	itemid := int32(1)
	appdata := make(map[string]string)
	appdata["amount"] = "15001"
	appdata["product"] = "1001"
	wh, err := workflow.New_FLowHelper(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	err2 := wh.AbandonCase(caseid, choice, remark, itemid, appdata)
	if err2 != nil {
		t.Error(err2)
		return
	}
}

func Test_FinishCase(t *testing.T) {
	caseid := "0ed936f1-560f-45a1-8269-db4b5d540c37"
	choice := "同意"
	remark := "我要作废"
	itemid := int32(1)
	appdata := make(map[string]string)
	appdata["amount"] = "15001"
	appdata["product"] = "1001"
	wh, err := workflow.New_FLowHelper(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	err2 := wh.FinishCase(caseid, choice, remark, itemid, appdata)
	if err2 != nil {
		t.Error(err2)
		return
	}
}

func Test_SendbackCase(t *testing.T) {
	choice := "同意"
	remark := "我要作废"
	userid := "675608"
	username := "明亮9"
	appdata := make(map[string]string)
	appdata["amount"] = "15001"
	appdata["product"] = "1001"
	wfid := "aa9cadc9-35d0-4888-9852-b9395b55ee55"
	wh, err := workflow.New_FLowHelper(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	biz_1 := "12323232"
	biz_2 := ""
	caseid, err := wh.AddCase(wfid, username, userid, biz_1, biz_2, appdata)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("caseid: %#v", caseid)
	//----------------------
	itemid := int32(0)
	flowuser := &workflow.FlowUser{
		Userid:   userid,
		UserName: username,
	}
	instep, err2 := wh.CommitCase(caseid, choice, remark, itemid, flowuser, appdata)
	if err2 != nil {
		t.Error(err2)
		return
	}
	t.Log(instep)
	//----------------------
	itemid = int32(1)
	err3 := wh.SendbackCase(caseid, choice, remark, itemid, appdata)
	if err3 != nil {
		t.Error(err3)
		return
	}
}

func Test_FallbackCase(t *testing.T) {
	choice := "同意"
	remark := "我要作废"
	userid := "675608"
	username := "明亮9"
	appdata := make(map[string]string)
	appdata["amount"] = "15001"
	appdata["product"] = "1001"
	wfid := "aa9cadc9-35d0-4888-9852-b9395b55ee55"
	wh, err := workflow.New_FLowHelper(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	biz_1 := "12323232"
	biz_2 := ""
	caseid, err := wh.AddCase(wfid, username, userid, biz_1, biz_2, appdata)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("caseid: %#v", caseid)
	//----------------------
	itemid := int32(0)
	flowuser := &workflow.FlowUser{
		Userid:   userid,
		UserName: username,
	}
	instep, err2 := wh.CommitCase(caseid, choice, remark, itemid, flowuser, appdata)
	if err2 != nil {
		t.Error(err2)
		return
	}
	t.Log(instep)
	itemid = int32(1)
	instep, err3 := wh.CommitCase(caseid, choice, remark, itemid, flowuser, appdata)
	if err3 != nil {
		t.Error(err3)
		return
	}
	t.Log(instep)
	//----------------------
	itemid = int32(2)
	err4 := wh.FallbackCase(caseid, choice, remark, itemid, appdata)
	if err4 != nil {
		t.Error(err4)
		return
	}
}

func Test_ReadedCase(t *testing.T) {
	caseid := "12cbec16-44f8-4f33-9aaf-78f85fd3b3aa"
	itemid := int32(3)
	userid := "675608"
	wh, err := workflow.New_FLowHelper(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	err2 := wh.Readed(itemid, caseid, userid)
	if err2 != nil {
		t.Error(err2)
		return
	}
}

func Test_SetAgent1(t *testing.T) {
	userid := "675608"
	agentid := "647749"
	wh, err := workflow.New_FLowHelper(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	err2 := wh.SetAgent(userid, agentid)
	if err2 != nil {
		t.Error(err2)
		return
	}
}

func Test_SetAgent2(t *testing.T) {
	userid := "661639"
	agentid := "675608"
	wh, err := workflow.New_FLowHelper(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	err2 := wh.SetAgent(userid, agentid)
	if err2 != nil {
		t.Error(err2)
		return
	}
}

func Test_UnsetAgent(t *testing.T) {
	userid := "661639"
	wh, err := workflow.New_FLowHelper(conn_str)
	if err != nil {
		t.Error(err)
		return
	}
	err2 := wh.UnsetAgent(userid)
	if err2 != nil {
		t.Error(err2)
		return
	}
}
