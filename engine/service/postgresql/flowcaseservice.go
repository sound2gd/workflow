package workflow

import (
	"encoding/json"
	//"fmt"
	"github.com/jackc/pgx"
	"sort"
	"strconv"
	"strings"
	"time"
)

//--------------------流程定义提供者类----------------------------------------
type FlowPgProvider struct {
	ConnCfg *pgx.ConnConfig //数据库连接串
}

func New_FlowPgProvider(connstr string) (*FlowPgProvider, error) {
	cfg, err := util.GetConnCfg(connstr)
	if err != nil {
		return nil, err
	}
	p := &FlowPgProvider{
		ConnCfg: cfg,
	}
	return p, nil
}

//获得流程的定义keys:flowid, name, descript, flowxml
func (f *FlowPgProvider) GetFlow(flowid string) (*Flow, error) {
	log.Info("flowprovider.GetFlow", flowid)
	//fmt.Println(flowid)
	conn, err := pgx.Connect(*f.ConnCfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	sql := `select name,flowxml,versionno from crm_t_workflow where flowid=$1 limit 1`
	row := conn.QueryRow(sql, flowid)
	var flowname, flowxml string
	var verionno int32
	if err := row.Scan(&flowname, &flowxml, &verionno); err != nil {
		return nil, err
	}
	fl := New_Flow(flowid, flowname, flowxml, verionno)
	return fl, nil
}

func (f *FlowPgProvider) GetFlowByVersionNo(flowid string, versionno int32) (*Flow, error) {
	log.Info("flowprovider.GetFlowByVersionNo.flowid", flowid)
	log.Info("flowprovider.GetFlowByVersionNo.versionno", versionno)
	//fmt.Println(flowid)
	conn, err := pgx.Connect(*f.ConnCfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	sql := `(select name,flowxml,versionno from crm_t_workflow
		where flowid=$1 and versionno = $2 limit 1)
		union
		(select name,flowxml,versionno from crm_t_workflow_version
			where flowid=$1 and versionno = $2 limit 1)`
	row := conn.QueryRow(sql, flowid, versionno)
	var flowname, flowxml string
	var verionno int32
	if err := row.Scan(&flowname, &flowxml, &verionno); err != nil {
		return nil, err
	}
	fl := New_Flow(flowid, flowname, flowxml, verionno)
	return fl, nil
}

//加载一个流程的完整信息
func (f *FlowPgProvider) LoadFlowCase(caseid string) (*FlowCase, error) {
	//fmt.Println("LoadFlowCase")
	conn, err := pgx.Connect(*f.ConnCfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	//todo 抄送人
	sql := `select flowid,
	(select name from crm_t_workflow wf where wf.flowid = c.flowid) as flowname,
    (select itemid from crm_t_flowcaseitem item
			where item.caseid = c.caseid order by itemid desc limit 1) as itemid,
	(select bussinessid_1 from crm_t_appflowcase afc where afc.caseid = c.caseid) as bizid1,
    (select bussinessid_2 from crm_t_appflowcase afc where afc.caseid = c.caseid) as bizid2,
    (select appid from crm_t_appflowcase afc where afc.caseid = c.caseid) as appid,
    (select choiceitems from crm_t_flowcaseitem fc
			where c.caseid = fc.caseid order by  itemid desc limit 1) choiceitems,
	creator, creatorname, step, status, createtime, endtime,
	array_to_string(array(
	SELECT copyuser FROM crm_t_flowcopyuser as c
	where c.caseid = $1
	), ',') AS copyuser,
    '' as appdata,
    (select sendtime from crm_t_flowcaseitem item
			where item.caseid = c.caseid order by itemid desc limit 1) as sendtime,
    serialnumber,
   (select handleuserid from crm_t_flowcaseitem item
		 where item.caseid = c.caseid order by  itemid desc limit 1) handleuserid,
   (select handleusername from crm_t_flowcaseitem item
		 where item.caseid = c.caseid order by  itemid desc limit 1) handleusername,
   (select handletime from crm_t_flowcaseitem item
		 where item.caseid = c.caseid order by  itemid desc limit 1) handletime,
   (select stepstatus from crm_t_flowcaseitem item
		 where item.caseid = c.caseid order by  itemid desc limit 1) stepstatus,
   (select pluginid from crm_t_plugin p where p.workflowid = c.flowid limit 1) as pluginid,
   versionno
	 from crm_t_flowcase c where caseid=$1 limit 1`

	row := conn.QueryRow(sql, caseid)
	var flowid, flowname, creator, creatorname, step, copyuserStr, appid, handleuserid, handleusername string
	var status, itemid, stepstatus, versionno int32
	var createtime, endtime, sendtime, handletime pgx.NullTime
	var appdata, bizid1, bizid2, serialnumber, pluginid pgx.NullString
	var choiceitems pgx.NullString
	if err := row.Scan(&flowid, &flowname, &itemid, &bizid1, &bizid2, &appid, &choiceitems, &creator,
		&creatorname, &step, &status, &createtime, &endtime, &copyuserStr, &appdata, &sendtime,
		&serialnumber, &handleuserid, &handleusername, &handletime, &stepstatus, &pluginid, &versionno); err != nil {
		return nil, err
	}

	var copyuser []int = make([]int, 0)
	if copyuserStr != "" {
		users := strings.Split(copyuserStr, ",")
		if len(users) > 0 {
			for _, user := range users {
				u, _ := strconv.Atoi(user)
				copyuser = append(copyuser, u)

			}
		}
	}

	c := &Case{
		CaseID:         caseid,
		ItemID:         itemid,
		AppID:          appid,
		FlowID:         flowid,
		FlowName:       flowname,
		CreatorID:      creator,
		CreatorName:    creatorname,
		Step:           step,
		Status:         status,
		CreateTime:     createtime.Time,
		CopyUser:       copyuser,
		HandleUserid:   handleuserid,
		HandleUserName: handleusername,
		StepStatus:     stepstatus,
		VersionNo:      versionno,
	}
	if appdata.Valid {
		c.AppData = appdata.String
	}
	if bizid1.Valid {
		c.BizID1 = bizid1.String
	}
	if bizid2.Valid {
		c.BizID2 = bizid2.String
	}
	if endtime.Valid {
		c.EndTime = endtime.Time
	}
	if sendtime.Valid {
		c.SendTime = sendtime.Time.Format(f_datetime)
	}
	if choiceitems.Valid {
		c.ChoiceItems = choiceitems.String
	}
	if serialnumber.Valid {
		c.SerialNumber = serialnumber.String
	}
	if handletime.Valid {
		c.HandleTime = handletime.Time.Format(f_datetime)
	}
	if pluginid.Valid {
		c.PluginID = pluginid.String
	}
	cis, err := f.GetCaseItems(caseid)
	if err != nil {
		return nil, err
	}

	fc := &FlowCase{CaseInfo: c, CaseItems: cis}
	//fmt.Println(fc)
	return fc, nil
}

func (f *FlowPgProvider) GetCaseItems(caseid string) (map[int32]*CaseItem, error) {
	//fmt.Println("GetCaseItems")
	conn, err := pgx.Connect(*f.ConnCfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	//加载items
	sqlc := `SELECT itemid, handleuserid, handleusername,stepname, stepstatus,
	 choice, mark, createtime,
	 handletime,
	 agentuserid, agentusername,
	 sysenterinfo, sysexitinfo,choiceitems
	  from crm_t_flowcaseitem where caseid=$1 order by itemid asc`

	rows, err := conn.Query(sqlc, caseid)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cis := make(map[int32]*CaseItem)
	//fmt.Println("good")

	for rows.Next() {
		var itemid int32
		var stepstatus int32
		var stepname string
		var choice, mark, agentuserid, agentusername, handleuserid, handleusername pgx.NullString
		var sysenterinfo, sysexitinfo pgx.NullString
		var createtime time.Time
		var handletime pgx.NullTime
		var choiceitems pgx.NullString
		if err := rows.Scan(&itemid, &handleuserid, &handleusername, &stepname, &stepstatus,
			&choice, &mark, &createtime,
			&handletime,
			&agentuserid, &agentusername,
			&sysenterinfo, &sysexitinfo, &choiceitems); err != nil {

			return nil, err
		}
		ci := &CaseItem{
			ItemID:     itemid,
			StepName:   stepname,
			CreateTime: createtime,
			StepStatus: stepstatus,
		}
		//fmt.Println(ci)
		if handleuserid.Valid {
			ci.HandleUserid = handleuserid.String
		}
		if handleusername.Valid {
			ci.HandleUserName = handleusername.String
		}
		if choice.Valid {
			ci.Choice = choice.String
		}
		if mark.Valid {
			ci.Mark = mark.String
		}
		if handletime.Valid {
			ci.HandleTime = handletime.Time.Format(f_datetime)
		}
		if agentuserid.Valid {
			ci.AgentUserid = agentuserid.String
		}
		if agentusername.Valid {
			ci.AgentUserName = agentusername.String
		}
		if sysenterinfo.Valid {
			ci.SysEnterInfo = sysenterinfo.String
		}
		if sysexitinfo.Valid {
			ci.SysExitInfo = sysexitinfo.String
		}
		if choiceitems.Valid {
			ci.ChoiceItems = choiceitems.String
		}
		cis[ci.ItemID] = ci
	}
	//fmt.Println(cis)
	return cis, nil
}

//保存一个新的流程实例
func (f *FlowPgProvider) SaveNewCase(fc *FlowCase, versionno int32) (string, error) {
	conn_query, err := pgx.Connect(*f.ConnCfg)
	if err != nil {
		return "", err
	}
	defer conn_query.Close()
	//流水号
	var serialnumberTemp string
	sql_querySN := `
	select * from crm_f_generate_serialnumber($1)`

	row := conn_query.QueryRow(sql_querySN, "crm_t_flowcase")
	row.Scan(&serialnumberTemp)
	log.Debug("flowprovide.SaveNewCase", "流水号"+serialnumberTemp)

	conn, err := pgx.Connect(*f.ConnCfg)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	tx, _ := conn.Begin()
	//插入case
	ca := fc.CaseInfo
	sql_case := `INSERT INTO crm_t_flowcase(
	caseid, flowid, creator, creatorname, step, createtime,serialnumber,versionno)
	VALUES ($1, $2, $3, $4, $5, $6,$7,$8)`
	if _, err := tx.Exec(sql_case, ca.CaseID, ca.FlowID, ca.CreatorID, ca.CreatorName,
		ca.Step, ca.CreateTime, serialnumberTemp, versionno); err != nil {
		tx.Rollback()
		return "", err
	}
	//插入item, 只有一条
	ci := fc.CaseItems[0]
	sql_ci := `INSERT INTO crm_t_flowcaseitem(
	itemid, caseid, handleuserid, handleusername, stepname, stepstatus, createtime,handletime,sendtime)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	if _, err := tx.Exec(sql_ci, ci.ItemID, ca.CaseID, ci.HandleUserid, ci.HandleUserName,
		ci.StepName, ci.StepStatus, ci.CreateTime, time.Now(), time.Now()); err != nil {
		tx.Rollback()
		return "", err
	}
	tx.Commit()
	return serialnumberTemp, nil
}

//在一个事务中提交流程数据
func (f *FlowPgProvider) ComitFlow(ca *Case, ci *CaseItem, ni *CaseItem) error {
	conn, err := pgx.Connect(*f.ConnCfg)
	if err != nil {
		return err
	}
	defer conn.Close()
	tx, _ := conn.Begin()

	if ca.Status == 0 {
		sql_case := `UPDATE crm_t_flowcase
	SET step=$1, status=$2
	WHERE caseid=$3`
		if _, err := tx.Exec(sql_case, ca.Step, ca.Status, ca.CaseID); err != nil {
			tx.Rollback()
			return err
		}
	} else { //只有结束流程的时候才写结束时间
		sql_case := `UPDATE crm_t_flowcase
	SET step=$1, endtime=$2, status=$3
	WHERE caseid=$4`
		if _, err := tx.Exec(sql_case, ca.Step, ca.EndTime, ca.Status, ca.CaseID); err != nil {
			tx.Rollback()
			return err
		}
	}
	sql_ci := `UPDATE crm_t_flowcaseitem
	SET stepstatus=$1, choice=$2, mark=$3, handletime=$4, sysexitinfo=$5
	 WHERE itemid=$6 and caseid=$7`
	if _, err := tx.Exec(sql_ci, ci.StepStatus, ci.Choice, ci.Mark, ci.HandleTime, ci.SysExitInfo,
		ci.ItemID, ca.CaseID); err != nil {
		tx.Rollback()
		return err
	}

	//审核选项写入
	//fmt.Println("审核选项")
	flow, err := f.GetFlowByVersionNo(ca.FlowID, ca.VersionNo)
	if err != nil {
		return err
	}
	choices := ""
	dynamicChoice := make([]*ChoiceResp, 0, 1)
	result := make(map[int]*Choice)

	for _, status := range flow.FlowStatus {
		//fmt.Println(status)
		if ni.StepName == status.Name {
			for _, c := range status.Choices {
				result[c.Index] = c
			}
		}
	}
	log.Info("before sort....ComitFlow.sort.ByIndex", result)
	sort.Sort(ByIndex(result))
	log.Info("after sort....ComitFlow.sort.ByIndex", result)

	for _, c := range result {
		dynamicChoice = append(dynamicChoice, &ChoiceResp{Name: c.Name, DataItems: c.DataItems})
	}

	cho, _ := json.Marshal(dynamicChoice)
	choices = string(cho)

	//fmt.Println(len(choices))
	sql_ni := `INSERT INTO crm_t_flowcaseitem(
	itemid, caseid, handleuserid, handleusername, stepname, stepstatus, createtime, agentuserid,
	 agentusername, sysenterinfo,choiceitems,handletime)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	if _, err := tx.Exec(sql_ni, ni.ItemID, ca.CaseID, ni.HandleUserid, ni.HandleUserName,
		ni.StepName, ni.StepStatus, ni.CreateTime, ni.AgentUserid, ni.AgentUserName, ni.SysEnterInfo,
		choices, time.Now()); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

//找到步骤处理人的代理人
func (f *FlowPgProvider) FindAgent(userid string) (user *FlowUser, find bool) {
	conn, err := pgx.Connect(*f.ConnCfg)
	if err != nil {
		return nil, false
	}
	defer conn.Close()

	sql := `select a.agentid, u.username
		from crm_t_flowagent a
		inner join com_t_userinfo u on cast(u.usernumber as varchar)=a.agentid
		where a.userid=$1`
	row := conn.QueryRow(sql, userid)
	var agentid, agentname string
	if err := row.Scan(&agentid, &agentname); err != nil {
		return nil, false
	}
	user = &FlowUser{
		Userid:   agentid,
		UserName: agentname,
	}
	return user, true
}

//记录步骤进入,退出的消息
func (f *FlowPgProvider) StepHandled(ca *Case, ci *CaseItem, ni *CaseItem) error {
	conn, err := pgx.Connect(*f.ConnCfg)
	if err != nil {
		return err
	}
	defer conn.Close()
	sql_ci := `UPDATE crm_t_flowcaseitem SET sysexitinfo=$1 WHERE itemid=$2 and caseid=$3`
	if _, err := conn.Exec(sql_ci, ci.SysExitInfo, ci.ItemID, ca.CaseID); err != nil {
		return err
	}
	sql_ni := `UPDATE crm_t_flowcaseitem SET sysenterinfo=$1 WHERE itemid=$2 and caseid=$3`
	if _, err := conn.Exec(sql_ni, ni.SysEnterInfo, ni.ItemID, ca.CaseID); err != nil {
		return err
	}
	return nil
}

//回写流程发送时间
func (f *FlowPgProvider) WriteBackSendTime(caseinfo *CaseInfo) error {
	sql := `UPDATE crm_t_flowcaseitem set sendtime = $1 where caseid = $2 and itemid = $3`
	conn, err := pgx.Connect(*f.ConnCfg)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.Exec(sql, time.Now(), caseinfo.CaseID, caseinfo.ItemID); err != nil {
		return err
	}
	return nil
}

// //修改流程的状态
// func (f *FlowPgProvider) ChangeFlowStatus(caseid string, statusname string) error {
// 	conn, err := pgx.Connect(*f.ConnCfg)
// 	if err != nil {
// 		return err
// 	}
// 	defer conn.Close()
// 	sql_case := `UPDATE crm_t_flowcase
// 	SET status=$1, endtime=$2
// 	WHERE caseid=$3`
// 	if _, err := tx.Exec(sql_case, statusname, time.Now(), caseid); err != nil {
// 		return err
// 	}
// }

// //修改工作项的状态
// func (f *FlowPgProvider) ChangeItemStatus(caseid string, itemid int, status WorkItemStatus) error {

// }
