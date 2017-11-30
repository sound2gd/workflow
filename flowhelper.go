package workflow

import (
	"errors"
	"fmt"
	"github.com/jackc/pgx"
	"strings"
	//"strings"
	"github.com/nsqio/go-nsq"
	"github.com/satori/go.uuid"
	"strconv"
	"time"
	"xtion.net/mcrm/flowservice/message"
	log "xtion.net/mcrm/logger"
	"xtion.net/mcrm/util"
)

//事务
type CaseInfo struct {
	CaseId         string    `json:"caseid"`         //流程实例id
	ItemId         int32     `json:"itemid"`         //流程当前步骤id
	FlowId         string    `json:"flowid"`         //流程id
	Name           string    `json:"flowname"`       //流程名称
	Creator        string    `json:"creator"`        //流程发起人账号
	Creatorname    string    `json:"creatorname"`    //流程发起人姓名
	Createtime     time.Time `json:"createtime"`     //流程发起时间
	Handleuserid   string    `json:"handleuserid"`   //步骤原处理人(有代理人)
	Handleusername string    `json:"handleusername"` //步骤原处理人姓名
	Handletime     string    `json:"handletime"`     //处理时间
	ChoiceItems    string    `json:"choiceitems"`    //审核选项
	Stepname       string    `json:"stepname"`       //当前步骤名称
	Stepstatus     int32     `json:"stepstatus"`     //当前步骤的状态，0:未处理 1:已读 2:已处理
	Status         int32     `json:"status"`         //状态,0:审批中 1:通过 2:不通过
	Appid          string    `json:"appid"`          //流程关联的业务对象(记录在crm_t_entityreg)
	Bizid1         string    `json:"bizid1"`         //业务主键1
	Bizid2         string    `json:"bizid2"`         //业务主键2
	SendTime       string    `json:"sendtime"`       //发送时间
	SerialNumber   string    `json:"serialnumber"`   //流水号
	Choice         string    `json:"choice"`         //审核
	PluginId       string    `json:"pluginid"`       //插件id
	FlowStatus     int32     `json:"flowstatus"`     //流程状态 1启用0停用
}

//代办事务
type CaseList struct {
	Items      []*CaseInfo
	TotalItems int32
}

//流程列表
type FlowList struct {
	Items      []*FlowInfo
	TotalItems int32
}

//流程的信息
type FlowInfo struct {
	FlowId         string    `json:"flowid"`
	Name           string    `json:"flowname"`
	Descript       string    `json:"descript"`
	FlowXml        string    `json:"flowxml"`
	StepCount      int32     `json:"stepcount"`
	CreateTime     time.Time `json:"createtime"`
	Creator        string    `json:"creator"`
	Status         int32     `json:"status"`
	UpdateTime     string    `json:"updatetime"`
	Updator        string    `json:"updator"`
	FlowType       int32     `json:"flowtype"`
	AppId          string    `json:"appid"`
	EntityType     int32     `json:"entitytype"`     //1系统对象2插件对象
	FlowCategory   int32     `json:"flowcategory"`   //1表示固定流程，0表示自由流程
	PluginStatus   int32     `json:"pluginstatus"`   //插件状态 1在用
	PVersionStatus int32     `json:"pversionstatus"` //插件版本
	PowerControl   int32     `json:"powercontrol"`   //权限控制
}

//------------------------------------------------------------
type FlowHelper struct {
	conCfg   *pgx.ConnConfig
	constr   string
	producer *nsq.Producer
	topic    string
}

//初始化
func New_FLowHelper(constr string, producer *nsq.Producer, topic string) (*FlowHelper, error) {

	cfg, err := util.GetConnCfg(constr)
	if err != nil {
		return nil, err
	}
	fh := &FlowHelper{
		conCfg:   cfg,
		constr:   constr,
		producer: producer,
		topic:    topic,
	}
	return fh, nil
}

//获取用户的代办列表---flowname查询条件
func (f *FlowHelper) GetTodoCases(flowname, usernumber string,
	pageindex, pagesize int32) (*CaseList, error) {

	todolist := &CaseList{make([]*CaseInfo, 0, pagesize), 0}
	sql := `select i.itemid, i.caseid, i.handleuserid, i.handleusername, i.stepname, i.handletime,
            i.createtime, c.creator, c.creatorname, c.status, f.name, f.flowid,
            af.appid, afc.bussinessid_1, afc.bussinessid_2
            from crm_t_flowcaseitem i
            inner join crm_t_flowcase c on c.caseid = i.caseid
            inner join crm_t_workflow f on c.flowid = f.flowid
            inner join crm_t_appflow af on af.flowid = c.flowid
			inner join crm_t_appflowcase afc on afc.caseid = c.caseid
            where (i.handleuserid=$1 or i.agentuserid=$1)
            and i.stepname=c.step and i.stepstatus=0`
	if flowname != "" {
		sql += " and f.name like $xwflag$%" + flowname + "%$xwflag$"
	}
	sqllimit := fmt.Sprintf(" limit %v offset %v", pagesize, (pageindex-1)*pagesize)
	conn, err := pgx.Connect(*f.conCfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	rows, err := conn.Query(sql+sqllimit, usernumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		af := &CaseInfo{}
		var biz2 pgx.NullString
		var ht pgx.NullTime
		err := rows.Scan(&(af.ItemId), &(af.CaseId), &(af.Handleuserid), &(af.Handleusername),
			&(af.Stepname), &ht, &(af.Createtime), &(af.Creator), &(af.Creatorname), &(af.Status),
			&(af.Name), &(af.FlowId), &(af.Appid), &(af.Bizid1), &biz2)
		if err != nil {
			return nil, err
		}
		if biz2.Valid {
			af.Bizid2 = biz2.String
		}
		if ht.Valid {
			af.Handletime = ht.Time.Format(f_datetime)
		}

		todolist.Items = append(todolist.Items, af)
		todolist.TotalItems++
	}
	return todolist, nil
}

//获取用户事务列表
func (f *FlowHelper) GetMyCases(usernumber string, finishstate, filter, pageindex, pagesize int32,
	flowid, keyword, begintime, endtime, createtime, handletime, sorttype string) (*CaseList, error) {
	calist := &CaseList{make([]*CaseInfo, 0, pagesize), 0}
	sql := `with copy as(
			select copy.caseid,copy.copyuser from crm_t_flowcopyuser copy
		   ),
		   tmp as(
           select c.itemid,c.caseid,c.handleuserid,c.handleusername,c.stepname,c.stepstatus,c.handletime,
					 c.choiceitems,c.sendtime,c.choice
           from crm_t_flowcaseitem c where 1=1 %v
           ),
		   tmpmax as(
           select max(c.itemid) as itemid,c.caseid
           from crm_t_flowcaseitem c where 1=1 and exists(select 1 from tmp where caseid = c.caseid)
           group by c.caseid
           )

           select
           (select itemid from crm_t_flowcaseitem tmp
						 where tmp.caseid = fc.caseid order by  itemid desc limit 1) itemid,
           fc.caseid,fc.createtime,fc.creator,fc.creatorname,wf.flowid,wf.name,
           case when fc.step ='通过' or fc.step ='不通过' then
					 (select handleuserid from crm_t_flowcaseitem tmp
						 inner join tmpmax on  tmp.caseid = tmpmax.caseid
						where tmp.caseid = fc.caseid  and  tmp.itemid = (tmpmax.itemid-1) limit 1)
	       else
           (select handleuserid from crm_t_flowcaseitem tmp
						 where tmp.caseid = fc.caseid order by  itemid desc limit 1)
           end as handleuserid,

           case when fc.step ='通过' or fc.step='不通过' then
					 (select handleusername from crm_t_flowcaseitem tmp
						 inner join tmpmax on  tmp.caseid = tmpmax.caseid
						 where tmp.caseid = fc.caseid  and  tmp.itemid = (tmpmax.itemid-1) limit 1)
	       else
           (select handleusername from crm_t_flowcaseitem tmp
						 where tmp.caseid = fc.caseid order by  itemid desc limit 1)
           end as handleusername,

           case when fc.step ='通过' or fc.step='不通过' then
					 (select choice from crm_t_flowcaseitem tmp
						 inner join tmpmax on  tmp.caseid = tmpmax.caseid
						 where tmp.caseid = fc.caseid  and  tmp.itemid = (tmpmax.itemid-1) limit 1)
	       else
           (select choice from crm_t_flowcaseitem tmp
						 where tmp.caseid = fc.caseid order by  itemid desc limit 1)
           end as choice,

           (select stepname from crm_t_flowcaseitem tmp
						 where tmp.caseid = fc.caseid order by  itemid desc limit 1) stepname,
           (select stepstatus from crm_t_flowcaseitem tmp
						 where tmp.caseid = fc.caseid order by  itemid desc limit 1) stepstatus,
           fc.status,
           wf.status as flowstatus,
           (select handletime from crm_t_flowcaseitem tmp
						 where tmp.caseid = fc.caseid order by  itemid desc limit 1) handletime,
           (select choiceitems from crm_t_flowcaseitem tmp
						 where tmp.caseid = fc.caseid order by  itemid desc limit 1) choiceitems,
           (select sendtime from crm_t_flowcaseitem tmp
						 where tmp.caseid = fc.caseid order by  itemid desc limit 1) sendtime,
           af.appid, afc.bussinessid_1, afc.bussinessid_2,serialnumber,
           (select pluginid from crm_t_plugin where workflowid = wf.flowid limit 1) as pluginid
           from crm_t_flowcase fc
           inner join crm_t_workflow wf on wf.flowid = fc.flowid
           inner join crm_t_appflow af on af.flowid = wf.flowid
           inner join crm_t_appflowcase afc on afc.caseid = fc.caseid
           where 1=1 and (exists(select 1 from tmp where fc.caseid = tmp.caseid) %v)`
	// status integer NOT NULL DEFAULT 0, -- 0:审批中 1:通过 2:不通过
	// stepstatus integer NOT NULL, -- 0:未处理; 1:已读; 2:已处理
	tmp_sqlwhere := ""
	v1_sqlwhere := "" // %v1
	v2_sqlwhere := "" // %v2

	//检查参数
	if filter < 1 || filter > 4 ||
		(finishstate < 0 || finishstate > 5) {
		return nil, errors.New("not supported params")
	}

	var sqlwhere string

	//filter: 1我申请的 2审批人是我的 3抄送我的 4包括2和3
	//finishstate: 1已通过 0审批中 2已中止（未通过）3未审批
	//4已审批 5全部

	//new
	if filter == 1 {
		//我申请的 已通过1 审批中0 已中止2 不限5
		sqlwhere += ` and creator = $1`

		if finishstate == 1 {
			sqlwhere += ` and exists(select 1 from tmp where fc.caseid = tmp.caseid and fc.status=1)`
		} else if finishstate == 0 {
			sqlwhere += ` and exists(select 1 from tmp where fc.caseid = tmp.caseid and fc.status=0)`
		} else if finishstate == 2 {
			sqlwhere += ` and exists(select 1 from tmp where fc.caseid = tmp.caseid and fc.status=2)`
		}
	} else if filter == 2 {
		//审批人是我的 未审批3 已审批4 不限5
		if finishstate == 3 {
			tmp_sqlwhere += ` and handleuserid = $1 `
			tmp_sqlwhere += ` and (stepstatus = 0 or stepstatus = 1)`
		} else if finishstate == 4 {
			v1_sqlwhere += ` and handleuserid = $1 `
			v1_sqlwhere += ` and (stepstatus = 2) and itemid <>0 `
		} else if finishstate == 5 {
			v1_sqlwhere += ` and handleuserid = $1 `
			v1_sqlwhere += ` and itemid <>0 `
		}

	} else if filter == 3 {
		//抄送我的 未审批3 已审批4 不限5
		sqlwhere += ` and exists(select 1 from copy where fc.caseid = copy.caseid and copy.copyuser::text= $1)`
		if finishstate == 3 {
			tmp_sqlwhere += ` and handleuserid = $1 `
			tmp_sqlwhere += ` and (stepstatus = 0 or stepstatus = 1)`
		} else if finishstate == 4 {
			tmp_sqlwhere += ` and handleuserid = $1 `
			v1_sqlwhere += ` and (stepstatus = 2) and itemid <>0 `
		} else if finishstate == 5 {
			v1_sqlwhere += ` and itemid <>0 `
		}
	} else if filter == 4 {
		//审批人或者抄送我的 未审批3 已审批4 不限5
		if finishstate == 3 {
			tmp_sqlwhere += ` and handleuserid = $1 `
			tmp_sqlwhere += ` and (stepstatus = 0 or stepstatus = 1)`
		} else if finishstate == 4 {
			v1_sqlwhere += ` and handleuserid = $1 `
			v1_sqlwhere += ` and (stepstatus = 2) and itemid <>0 `
		} else if finishstate == 5 {
			v1_sqlwhere += ` and handleuserid = $1 `
			v1_sqlwhere += ` and itemid <>0 `
			v2_sqlwhere += ` or exists(select 1 from copy where fc.caseid = copy.caseid and copy.copyuser::text = $1)`
		}
	}

	sql = fmt.Sprintf(sql, v1_sqlwhere, v2_sqlwhere)

	if flowid != "" {
		sqlwhere += " and wf.flowid = '" + flowid + "'"
	}

	if keyword != "" {
		sqlwhere += " and (wf.name like $xwflag$%" + keyword + "%$xwflag$ or fc.creatorname like $xwflag$%" +
			keyword + "%$xwflag$)"
	}

	sort := ""
	condition := ""
	sql = "select * from (" + sql + sqlwhere + ") as t  where 1=1 "

	sql += tmp_sqlwhere

	if begintime != "" && endtime != "" {
		condition += " and createtime::date between '" + begintime + "' and '" + endtime + "'"
	} else {
		if begintime != "" {
			condition += " and createtime::date = '" + begintime + "'"
		}

		if endtime != "" {
			condition += " and createtime::date = '" + endtime + "'"
		}
	}

	if handletime != "" {
		condition += " and handletime is not null and handletime < '" + handletime + "'"
	}
	if createtime != "" {
		condition += " and createtime is not null and createtime < '" + createtime + "'"
	}

	//排序
	if sorttype == "" || sorttype == "1" {
		sort = " handletime desc,"
	} else if sorttype == "2" {
		sort = " createtime desc,"
	}
	sort = fmt.Sprintf(" order by %v itemid asc", sort)

	count_sql := "select cast(count(1) as integer) as count from (" + sql + condition + ") as t"
	sql = sql + condition + sort

	sqllimit := fmt.Sprintf(" limit %v offset %v", pagesize, (pageindex-1)*pagesize)
	conn, err := pgx.Connect(*f.conCfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	//todo: 没有$1参数, 多个一个usernumber会报错, 参数个数不匹配.
	//todo: agent 没有处理
	//log.Info("work.flowhelper.GetMyCase", sql+sqllimit)
	rows, err := conn.Query(sql+sqllimit, usernumber)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		af := &CaseInfo{}

		var ht, sendtime pgx.NullTime
		var biz_2, handleuserid, handleusername, choice pgx.NullString
		var choiceitems, serialnumber, pluginid pgx.NullString

		err := rows.Scan(&af.ItemId, &af.CaseId, &af.Createtime, &af.Creator, &af.Creatorname, &af.FlowId, &af.Name,
			&handleuserid, &handleusername, &choice, &af.Stepname, &af.Stepstatus, &af.Status, &af.FlowStatus,
			&ht, &choiceitems, &sendtime, &af.Appid, &af.Bizid1, &biz_2, &serialnumber, &pluginid)
		if err != nil {
			log.Debug("workflow.flowhelper.GetMyCases", err.Error())
		}
		if ht.Valid {
			af.Handletime = ht.Time.Format(f_datetime)
		}
		if biz_2.Valid {
			af.Bizid2 = biz_2.String
		}
		if choiceitems.Valid {
			af.ChoiceItems = choiceitems.String
		}
		if sendtime.Valid {
			af.SendTime = sendtime.Time.Format(f_datetime)
		}
		if serialnumber.Valid {
			af.SerialNumber = serialnumber.String
		}
		if handleuserid.Valid {
			af.Handleuserid = handleuserid.String
		}
		if handleusername.Valid {
			af.Handleusername = handleusername.String
		}
		if choice.Valid {
			af.Choice = choice.String
		}
		if pluginid.Valid {
			af.PluginId = pluginid.String
		}
		calist.Items = append(calist.Items, af)
		//calist.TotalItems++
	}

	//总条数
	row := conn.QueryRow(count_sql, usernumber)
	var count int32
	err = row.Scan(&count)
	if err != nil {
		return nil, err
	}
	calist.TotalItems = count

	return calist, nil
}

func (f *FlowHelper) GetWorkFlows(status, flowname string, pageindex, pagesize int32) (*FlowList, error) {
	return f.GetWorkFlowsForMobile(status, flowname, pageindex, pagesize)
}

func (f *FlowHelper) GetWorkFlowsForMobile(status, flowname string,
	pageindex, pagesize int32) (*FlowList, error) {

	var dynamic_sql string
	//dynamic_sql = " and pstatus = 1"
	dynamic_sql = ""
	return f.WorkFlows(status, flowname, pageindex, pagesize, dynamic_sql)
}

func (f *FlowHelper) GetWorkFlowsForWeb(status, flowname string,
	pageindex, pagesize int32) (*FlowList, error) {

	return f.WorkFlows(status, flowname, pageindex, pagesize, "")
}

//获取流程定义列表
func (f *FlowHelper) WorkFlows(status, flowname string, pageindex, pagesize int32,
	dynamic_sql string) (*FlowList, error) {

	flowlist := &FlowList{make([]*FlowInfo, 0, pagesize), 0}
	sql := `select * from(with t1 as(
  select flow.flowid, flow.name, flow.descript, flow.stepcount, flow.createtime,
	flow.createusername,flow.status,flow.updatetime,flow.updateusername,flow.flowtype,
	(select appid from crm_t_appflow app where app.flowid = flow.flowid limit 1) appid,
	(select xwentitytype from crm_t_appflow app
		inner join crm_t_entityreg reg on reg.xwentityregid = app.appid
		where app.flowid = flow.flowid limit 1) entitytype,flowcategory,p.status as pstatus,
		(select releasestatus from crm_t_plugin_version version
		 where version.originpluginid = p.pluginid order by pluginversion desc limit 1) as pversionstatus
	 from crm_t_workflow flow
	 left join crm_t_plugin p on flow.flowid = p.workflowid where 1=1
	 )
	 select * from t1 where t1.entitytype = 2 %v
	 union
	 select * from t1 where t1.entitytype = 1) t where 1=1`

	// if flowid != "" {
	// 	sql += " and flowid = '" + flowid + "'"
	sql = fmt.Sprintf(sql, dynamic_sql)

	if status != "" {
		sql += " and status =" + status
	}
	if flowname != "" {
		sql += " and name like $xwflag$%" + flowname + "%$xwflag$"
	}

	// if dynamic_sql != "" {
	// 	sql += dynamic_sql
	// }

	count_sql := "select cast(count(1) as integer) as count from (" + sql + ") as t"
	sqllimit := fmt.Sprintf(" limit %v offset %v", pagesize, (pageindex-1)*pagesize)

	sortby := " order by updatetime desc,createtime desc"
	conn, err := pgx.Connect(*f.conCfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Query(sql + sortby + sqllimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var updatetime pgx.NullTime
	var updateusername, appid pgx.NullString
	var entitytype, flowcategory, pstatus, pversionstatus pgx.NullInt32
	for rows.Next() {
		data := &FlowInfo{}
		err := rows.Scan(&data.FlowId, &data.Name, &data.Descript, &data.StepCount, &data.CreateTime,
			&data.Creator, &data.Status, &updatetime, &updateusername, &data.FlowType, &appid,
			&entitytype, &flowcategory, &pstatus, &pversionstatus)
		if err != nil {
			return nil, err
		}
		if updatetime.Valid {
			data.UpdateTime = updatetime.Time.Format(f_datetime)
		}
		if updateusername.Valid {
			data.Updator = updateusername.String
		}
		if appid.Valid {
			data.AppId = appid.String
		}
		if entitytype.Valid {
			data.EntityType = entitytype.Int32
		}
		if flowcategory.Valid {
			data.FlowCategory = flowcategory.Int32
		}

		//系统对象发布状态默认为1已发布
		if data.EntityType == 2 {
			if pstatus.Valid {
				data.PluginStatus = pstatus.Int32
			}
		} else {
			data.PluginStatus = 1
		}

		if pversionstatus.Valid {
			data.PVersionStatus = pversionstatus.Int32
		}
		flowlist.Items = append(flowlist.Items, data)
	}

	//总条数
	row := conn.QueryRow(count_sql)
	var count int32
	err = row.Scan(&count)
	if err != nil {
		return nil, err
	}
	flowlist.TotalItems = count

	return flowlist, nil
}

func (f *FlowHelper) GetWorkFlowDetail(flowid string) (*FlowInfo, error) {

	sql := `select flowid, name, descript,flowxml, stepcount, createtime,
	createusername,status,updatetime,updateusername,flowtype,
	(select appid from crm_t_appflow app where app.flowid = flow.flowid limit 1) appid,
	(select xwentitytype from crm_t_appflow app
		inner join crm_t_entityreg reg on reg.xwentityregid = app.appid
		where app.flowid = flow.flowid limit 1) entitytype,
	(select plugin.powercontrol from crm_t_appflow app
    	left join crm_t_plugin plugin on app.appid = plugin.pluginid
		where app.flowid = flow.flowid limit 1) powercontrol,
	flowcategory
	from crm_t_workflow flow where 1=1 `
	if flowid != "" {
		sql += " and flowid = '" + flowid + "'"
	} else {
		sql += " and 1=2"
	}
	conn, err := pgx.Connect(*f.conCfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	row := conn.QueryRow(sql)

	var updatetime pgx.NullTime
	var updateusername, appid pgx.NullString
	var entitytype, flowcategory, powercontrol pgx.NullInt32

	data := &FlowInfo{}
	err = row.Scan(&data.FlowId, &data.Name, &data.Descript, &data.FlowXml, &data.StepCount, &data.CreateTime,
		&data.Creator, &data.Status, &updatetime, &updateusername, &data.FlowType, &appid, &entitytype, &powercontrol, &flowcategory)

	if err != nil {
		return nil, err
	}
	if updatetime.Valid {
		data.UpdateTime = updatetime.Time.Format(f_datetime)
	}
	if updateusername.Valid {
		data.Updator = updateusername.String
	}
	if appid.Valid {
		data.AppId = appid.String
	}
	if entitytype.Valid {
		data.EntityType = entitytype.Int32
	}
	if flowcategory.Valid {
		data.FlowCategory = flowcategory.Int32
	}
	if powercontrol.Valid {
		data.PowerControl = powercontrol.Int32
	} else {
		data.PowerControl = 0
	}

	return data, nil
}

//流程实例详情
func (f *FlowHelper) GetCaseDetail(caseid string) (*FlowCaseList, error) {
	wf, err := New_Workflow(f.constr)
	if err != nil {
		return nil, err
	}
	flowcase, err := wf.FlowProvid.LoadFlowCase(caseid)
	if err != nil {
		return nil, err
	}

	caseitems := make([]*CaseItem, 0, 1)
	for _, item := range flowcase.CaseItems {
		caseitems = append(caseitems, item)
	}
	res := &FlowCaseList{
		CaseInfo:  flowcase.CaseInfo,
		CaseItems: caseitems,
	}
	return res, nil
}

//新发起一个流程, 返回caseid
func (f *FlowHelper) AddCase(enterprise, caseid, flowid, flowname, usernumber, username, biz_1, biz_2,
	appid, handeruserid, handerusername string, copyuser []int,
	appdata, remark string) (string, string, error) {
	wf, err := New_Workflow(f.constr)
	if err != nil {
		return "", "", err
	}
	serialnumber := ""
	serialnumber, caseid, err = wf.CreateWorkflow(caseid, flowid, appdata, usernumber, username)
	if err != nil {
		return "", "", err
	}

	//run
	flowuser := &FlowUser{
		Userid:   handeruserid,
		UserName: handerusername,
	}
	wf.Run(0, "", remark, flowuser)

	//写入crm_t_appflowcase表
	conn, err := pgx.Connect(*f.conCfg)
	if err != nil {
		return "", "", err
	}
	defer conn.Close()
	//记录流程实例与业务数据的关联关系
	sqlinsert := `insert into crm_t_appflowcase(caseid, bussinessid_1, bussinessid_2,appid)
	values($1, $2, $3, $4)`
	if _, err = conn.Exec(sqlinsert, caseid, biz_1, biz_2, appid); err != nil {
		return "", "", err
	}

	//抄送人
	for _, user := range copyuser {
		sqlinsertCopyUser := `INSERT INTO crm_t_flowcopyuser(caseid, copyuser)VALUES ($1, $2);`
		if _, err = conn.Exec(sqlinsertCopyUser, caseid, user); err != nil {
			//跳过把？？？
			log.Error("workflow.flowhelper.AddCase", err.Error())
		}
	}

	users := make([]int, 0, 0)

	//流程详情
	err = wf.LoadWorkflow(caseid, appdata)
	flowname = wf.FlowDef.FlowName

	recordid1 := uuid.NewV4().String()
	recordid2 := uuid.NewV4().String()

	//处理人消息内容
	contentHandler := fmt.Sprintf("提交了%v【%v】申请，等待您的审批", flowname, serialnumber)
	//抄送人消息内容
	contentCopyer := fmt.Sprintf("抄送给您的%v【%v】申请，等待%v处理", flowname, serialnumber, handerusername)

	usernumberInt, _ := strconv.Atoi(usernumber)
	sqlInsertMessage := "select * from crm_f_m_w_taskjob_message_insert($1, $2, $3, $4, $5, $6, $7, $8);"
	sqlUpdateMessage := "select crm_f_m_w_message_updatestatus($1, '00000000-0000-0000-0000-000000000000', $2, 1);"
	parameters := biz_1 + ",1"

	// 处理人
	if _, err = conn.Exec(sqlInsertMessage, recordid1, caseid, contentHandler, parameters, 9, usernumberInt, time.Now().Format(f_datetime), handeruserid); err != nil {
		//跳过把？？？
		log.Error("workflow.flowhelper.PushMessage.Handler", err.Error())
	}
	if _, err = conn.Exec(sqlUpdateMessage, usernumber, recordid1); err != nil {
		//跳过把？？？
		log.Error("workflow.flowhelper.PushMessage.Handler", err.Error())
	}
	// 抄送人
	copyuserString := arrayToString(copyuser)
	// log.Info("workflow.flowhelper.PushMessage.Copyer", copyuserString)
	if copyuserString != "" {
		if _, err = conn.Exec(sqlInsertMessage, recordid2, caseid, contentCopyer, parameters, 11, usernumberInt, time.Now().Format(f_datetime), copyuserString); err != nil {
			//跳过把？？？
			log.Error("workflow.flowhelper.PushMessage.Copyer", err.Error())
		}
		if _, err = conn.Exec(sqlUpdateMessage, usernumber, recordid2); err != nil {
			//跳过把？？？
			log.Error("workflow.flowhelper.PushMessage.Copyer", err.Error())
		}
	}

	msg_code := 1100

	//处理人是自己不就给自己再推送消息
	if handeruserid != usernumber {
		//消息编码1100,消息id,流程实例id,流程类型7
		push_contentHandler := fmt.Sprintf("%v,%v,%v,%v,%v", strconv.Itoa(msg_code), recordid1, caseid, "9", parameters)
		msgModelHandler := &util.WfMessage{
			EnterNumber:  enterprise,
			FlowId:       caseid,
			Msg_key:      recordid1,
			Msg_title:    "您收到了一条审批消息",
			Content:      contentHandler,
			Push_content: push_contentHandler,
			Messagetype:  11,
			Usernumber:   usernumberInt,
			MessageCode:  msg_code,
			Sendtime:     time.Now().Format(f_datetime),
			Handler:      handeruserid,
			Agent:        "", //代理人e号
			Users:        users,
			MUsers:       users,
			SystemCode:   "SYS00011", // todo: 这里写死了, 应该从外部传入
		}

		log.Info("flowhelper--addcase", msgModelHandler)
		push.PublishMsg(msgModelHandler, f.topic, f.producer)
	}

	//消息编码1100,消息id,流程实例id,流程类型7
	copyuser = removeFromArray(copyuser, usernumberInt)
	push_contentCopyer := fmt.Sprintf("%v,%v,%v,%v,%v", strconv.Itoa(msg_code), recordid2, caseid, "11", parameters)
	msgModelCopyer := &util.WfMessage{
		EnterNumber:  enterprise,
		FlowId:       caseid,
		Msg_key:      recordid2,
		Msg_title:    "您收到了一条审批消息",
		Content:      contentCopyer,
		Push_content: push_contentCopyer,
		Messagetype:  11,
		Usernumber:   usernumberInt,
		MessageCode:  msg_code,
		Sendtime:     time.Now().Format(f_datetime),
		Handler:      "",
		Agent:        "", //代理人e号
		Users:        users,
		MUsers:       copyuser,
		SystemCode:   "SYS00011", // todo: 这里写死了, 应该从外部传入
	}

	log.Info("flowhelper--addcase", msgModelCopyer)
	push.PublishMsg(msgModelCopyer, f.topic, f.producer)

	return serialnumber, caseid, nil
}

//预新发起一个流程, 返回步骤和人
func (f *FlowHelper) PreAddCase(flowid, usernumber, username, appdata string) (*NextStatuInfo, error) {

	wf, err := New_Workflow(f.constr)
	if err != nil {
		return nil, err
	}
	_, err = wf.PreCreateWorkflow(flowid, appdata, usernumber, username)
	if err != nil {
		return nil, err
	}
	//prerun
	res, err := wf.PreRun(0, "")

	return res, err
}

//预提交, 选择审批选项, 返回下一步去到的步骤和可选审批人
func (f *FlowHelper) PreCommitCase(caseid, choice string, itemid int32,
	appdata string) (nsif *NextStatuInfo, err error) {

	wf, err := New_Workflow(f.constr)
	if err != nil {
		return nil, err
	}
	err = wf.LoadWorkflow(caseid, appdata)
	if err != nil {
		return nil, err
	}
	res, err := wf.PreRun(itemid, choice)

	return res, err
}

//处理待办项, 返回进入的状态名称
func (f *FlowHelper) CommitCase(enterprise, usernumber, caseid, choice, remark string, itemid int32,
	flowuser *FlowUser,
	appdata string) (string, error) {

	wf, err := New_Workflow(f.constr)
	if err != nil {
		return "", err
	}
	err = wf.LoadWorkflow(caseid, appdata)
	if err != nil {
		return "", err
	}
	stepname, err := wf.Run(itemid, choice, remark, flowuser)
	if err == nil {
		//流程详情
		wf, err := New_Workflow(f.constr)
		if err != nil {
			return stepname, err
		}
		err = wf.LoadWorkflow(caseid, appdata)
		if err != nil {
			return stepname, err
		}
		//flowid := wf.FlowDef.FlowId
		flowname := wf.FlowDef.FlowName
		serialnumber := wf.Fcase.CaseInfo.SerialNumber

		//发起人
		applyer := wf.Fcase.CaseInfo.CreatorId
		biz_1 := wf.Fcase.CaseInfo.BizId1

		//审批状态 0:审批中 1:通过 2:不通过
		//判断审批是否完成
		status := wf.Fcase.CaseInfo.Status
		statusName := ""
		if status == 1 {
			statusName = "通过"
		}
		if status == 2 {
			statusName = "结束"
		}

		creatorName := wf.Fcase.CaseInfo.CreatorName
		//previousHandlerName := wf.Fcase.CaseItems[itemid-1].HandleUserName

		//下一步处理人
		handeruserid := ""
		handlerName := ""
		if flowuser != nil {
			handeruserid = flowuser.Userid
			handlerName = flowuser.UserName
		}

		//审批过程处理过的人
		users := make([]int, 0, 0)
		// for _, item := range wf.Fcase.CaseItems {
		// 	if item.HandleUserid != handeruserid && item.HandleUserid != applyer {
		// 		hInt, _ := strconv.Atoi(item.HandleUserid)
		// 		users = append(users, hInt)
		// 	}
		// }
		//抄送人
		copyuser := wf.Fcase.CaseInfo.CopyUser

		conn, err := pgx.Connect(*f.conCfg)
		if err != nil {
			return "", err
		}
		defer conn.Close()

		recordid1 := uuid.NewV4().String()
		recordid2 := uuid.NewV4().String()
		recordid3 := uuid.NewV4().String()

		var contentHandler, contentCopyer, contentApplyer string

		contentHandler = fmt.Sprintf("%v了%v的%v【%v】申请，已转至由您的审批", choice, creatorName, flowname, serialnumber)
		if status == 0 {
			contentCopyer = fmt.Sprintf("%v了%v抄送给您的%v【%v】申请，等待%v处理", choice, creatorName, flowname, serialnumber, handlerName)
			contentApplyer = fmt.Sprintf("%v了您的%v【%v】申请，等待%v处理", choice, flowname, serialnumber, handlerName)
		} else {
			contentCopyer = fmt.Sprintf("%v了%v抄送给您的%v【%v】申请，已经%v", choice, creatorName, flowname, serialnumber, statusName)
			contentApplyer = fmt.Sprintf("%v了您的%v【%v】申请，已经%v", choice, flowname, serialnumber, statusName)
		}

		usernumberInt, _ := strconv.Atoi(usernumber)
		sqlInsertMessage := "select * from crm_f_m_w_taskjob_message_insert($1, $2, $3, $4, $5, $6, $7, $8);"
		sqlUpdateMessage := "select crm_f_m_w_message_updatestatus($1, '00000000-0000-0000-0000-000000000000', $2, 1);"
		itemidString := strconv.Itoa(int(itemid + 1)) //审批要往后推一步
		parameters := biz_1 + "," + itemidString

		// 处理人
		if _, err = conn.Exec(sqlInsertMessage, recordid1, caseid, contentHandler, parameters, 9, usernumberInt, time.Now().Format(f_datetime), handeruserid); err != nil {
			//跳过把？？？
			log.Error("workflow.flowhelper.PushMessage.Handler", err.Error())
		}
		if _, err = conn.Exec(sqlUpdateMessage, usernumber, recordid1); err != nil {
			//跳过把？？？
			log.Error("workflow.flowhelper.PushMessage.Handler", err.Error())
		}
		// 抄送人
		copyuserString := arrayToString(copyuser)
		// log.Info("workflow.flowhelper.PushMessage.Copyer", copyuserString)
		if copyuserString != "" {
			if _, err = conn.Exec(sqlInsertMessage, recordid2, caseid, contentCopyer, parameters, 11, usernumberInt, time.Now().Format(f_datetime), copyuserString); err != nil {
				//跳过把？？？
				log.Error("workflow.flowhelper.PushMessage.Copyer", err.Error())
			}
			if _, err = conn.Exec(sqlUpdateMessage, usernumber, recordid2); err != nil {
				//跳过把？？？
				log.Error("workflow.flowhelper.PushMessage.Copyer", err.Error())
			}
		}
		// 发起人
		if _, err = conn.Exec(sqlInsertMessage, recordid3, caseid, contentApplyer, parameters, 10, usernumberInt, time.Now().Format(f_datetime), applyer); err != nil {
			//跳过把？？？
			log.Error("workflow.flowhelper.PushMessage.Applyer", err.Error())
		}
		if _, err = conn.Exec(sqlUpdateMessage, usernumber, recordid3); err != nil {
			//跳过把？？？
			log.Error("workflow.flowhelper.PushMessage.Applyer", err.Error())
		}

		msg_code := 1101

		if handeruserid != usernumber {
			//消息编码1101,消息id,流程实例id,流程类型11
			push_contentHandler := fmt.Sprintf("%v,%v,%v,%v,%v", strconv.Itoa(msg_code), recordid1, caseid, "9", parameters)
			msgModelHandler := &util.WfMessage{
				EnterNumber:  enterprise,
				FlowId:       caseid,
				Msg_key:      recordid1,
				Msg_title:    "您收到了一条审批消息",
				Content:      contentHandler,
				Push_content: push_contentHandler,
				Usernumber:   usernumberInt,
				Messagetype:  11,
				// Isstore:      0,
				MessageCode: msg_code,
				Sendtime:    time.Now().Format(f_datetime),
				Applyer:     "",
				Handler:     handeruserid,
				Agent:       "", //代理人e号
				Users:       users,
				MUsers:      users,
				SystemCode:  "SYS00011", // todo: 这里写死了, 应该从外部传入
			}
			log.Info("flowhelper--commitcase", msgModelHandler)
			push.PublishMsg(msgModelHandler, f.topic, f.producer)
		}

		//消息编码1101,消息id,流程实例id,流程类型11
		push_contentCopyer := fmt.Sprintf("%v,%v,%v,%v,%v", strconv.Itoa(msg_code), recordid2, caseid, "11", parameters)
		copyuser = removeFromArray(copyuser, usernumberInt)
		msgModelCopyer := &util.WfMessage{
			EnterNumber:  enterprise,
			FlowId:       caseid,
			Msg_key:      recordid2,
			Msg_title:    "您收到了一条审批消息",
			Content:      contentCopyer,
			Push_content: push_contentCopyer,
			Usernumber:   usernumberInt,
			Messagetype:  11,
			// Isstore:      0,
			MessageCode: msg_code,
			Sendtime:    time.Now().Format(f_datetime),
			Applyer:     "",
			Handler:     "",
			Agent:       "", //代理人e号
			Users:       users,
			MUsers:      copyuser,
			SystemCode:  "SYS00011", // todo: 这里写死了, 应该从外部传入
		}
		log.Info("flowhelper--commitcase", msgModelCopyer)
		push.PublishMsg(msgModelCopyer, f.topic, f.producer)

		//消息编码1101,消息id,流程实例id,流程类型11
		if applyer != usernumber {
			push_contentApplyer := fmt.Sprintf("%v,%v,%v,%v,%v", strconv.Itoa(msg_code), recordid3, caseid, "10", parameters)
			msgModelApplyer := &util.WfMessage{
				EnterNumber:  enterprise,
				FlowId:       caseid,
				Msg_key:      recordid3,
				Msg_title:    "您收到了一条审批消息",
				Content:      contentApplyer,
				Push_content: push_contentApplyer,
				Usernumber:   usernumberInt,
				Messagetype:  11,
				// Isstore:      0,
				MessageCode: msg_code,
				Sendtime:    time.Now().Format(f_datetime),
				Applyer:     applyer,
				Handler:     "",
				Agent:       "", //代理人e号
				Users:       users,
				MUsers:      users,
				SystemCode:  "SYS00011", // todo: 这里写死了, 应该从外部传入
			}
			log.Info("flowhelper--commitcase", msgModelApplyer)
			push.PublishMsg(msgModelApplyer, f.topic, f.producer)
		}

	}

	return stepname, err
}

//作废流程实列
func (f *FlowHelper) AbandonCase(enterprise, usernumber, caseid, choice, remark string, itemid int32,
	appdata string) error {
	wf, err := New_Workflow(f.constr)
	if err != nil {
		return err
	}
	err = wf.LoadWorkflow(caseid, appdata)
	if err != nil {
		return err
	}
	user := &FlowUser{}

	//发起人
	applyer := wf.Fcase.CaseItems[itemid-1].HandleUserid
	//发起人姓名
	creatorName := wf.Fcase.CaseInfo.CreatorName
	//抄送人
	copyuser := wf.Fcase.CaseInfo.CopyUser
	//业务实体Id
	biz_1 := wf.Fcase.CaseInfo.BizId1
	//当前步骤处理人
	//handlerName := wf.Fcase.CaseInfo.HandleUserName
	//流程名
	flowname := wf.FlowDef.FlowName
	//流程序号
	serialnumber := wf.Fcase.CaseInfo.SerialNumber

	usernumberInt, _ := strconv.Atoi(usernumber)
	//审批过程处理过的人
	users := make([]int, 0, 0)

	recordid2 := uuid.NewV4().String()
	recordid3 := uuid.NewV4().String()

	contentCopyer := fmt.Sprintf("%v中止了抄送给您的%v【%v】申请，已经结束。中止原因：%v", creatorName, flowname, serialnumber, remark)
	contentApplyer := fmt.Sprintf("%v中止了%v【%v】申请，已经结束。中止原因：%v", creatorName, flowname, serialnumber, remark)

	// 抄送人
	copyuserString := arrayToString(copyuser)
	sqlInsertMessage := "select * from crm_f_m_w_taskjob_message_insert($1, $2, $3, $4, $5, $6, $7, $8);"
	sqlUpdateMessage := "select crm_f_m_w_message_updatestatus($1, '00000000-0000-0000-0000-000000000000', $2, 1);"
	itemidString := strconv.Itoa(int(itemid))
	parameters := biz_1 + "," + itemidString

	conn, err := pgx.Connect(*f.conCfg)
	if err != nil {
		return err
	}
	defer conn.Close()

	if copyuserString != "" {
		if _, err = conn.Exec(sqlInsertMessage, recordid2, caseid, contentCopyer, parameters, 11, usernumberInt, time.Now().Format(f_datetime), copyuserString); err != nil {
			//跳过把？？？
			log.Error("workflow.flowhelper.PushMessage.Copyer", err.Error())
		}
		if _, err = conn.Exec(sqlUpdateMessage, usernumber, recordid2); err != nil {
			//跳过把？？？
			log.Error("workflow.flowhelper.PushMessage.Copyer", err.Error())
		}
	}
	// 发起人
	if _, err = conn.Exec(sqlInsertMessage, recordid3, caseid, contentApplyer, parameters, 10, usernumberInt, time.Now().Format(f_datetime), applyer); err != nil {
		//跳过把？？？
		log.Error("workflow.flowhelper.PushMessage.Applyer", err.Error())
	}
	if _, err = conn.Exec(sqlUpdateMessage, usernumber, recordid3); err != nil {
		//跳过把？？？
		log.Error("workflow.flowhelper.PushMessage.Applyer", err.Error())
	}

	msg_code := 1101

	//消息编码1101,消息id,流程实例id,流程类型11
	push_contentCopyer := fmt.Sprintf("%v,%v,%v,%v,%v", strconv.Itoa(msg_code), recordid2, caseid, "11", parameters)
	copyuser = removeFromArray(copyuser, usernumberInt)
	msgModelCopyer := &util.WfMessage{
		EnterNumber:  enterprise,
		FlowId:       caseid,
		Msg_key:      recordid2,
		Msg_title:    "您收到了一条审批消息",
		Content:      contentCopyer,
		Push_content: push_contentCopyer,
		Usernumber:   usernumberInt,
		Messagetype:  11,
		// Isstore:      0,
		MessageCode: msg_code,
		Sendtime:    time.Now().Format(f_datetime),
		Applyer:     "",
		Handler:     "",
		Agent:       "", //代理人e号
		Users:       users,
		MUsers:      copyuser,
		SystemCode:  "SYS00011", // todo: 这里写死了, 应该从外部传入
	}
	log.Info("flowhelper--commitcase", msgModelCopyer)
	push.PublishMsg(msgModelCopyer, f.topic, f.producer)

	//消息编码1101,消息id,流程实例id,流程类型11
	if applyer != usernumber {
		push_contentApplyer := fmt.Sprintf("%v,%v,%v,%v,%v", strconv.Itoa(msg_code), recordid3, caseid, "10", parameters)
		msgModelApplyer := &util.WfMessage{
			EnterNumber:  enterprise,
			FlowId:       caseid,
			Msg_key:      recordid3,
			Msg_title:    "您收到了一条审批消息",
			Content:      contentApplyer,
			Push_content: push_contentApplyer,
			Usernumber:   usernumberInt,
			Messagetype:  11,
			// Isstore:      0,
			MessageCode: msg_code,
			Sendtime:    time.Now().Format(f_datetime),
			Applyer:     applyer,
			Handler:     "",
			Agent:       "", //代理人e号
			Users:       users,
			MUsers:      users,
			SystemCode:  "SYS00011", // todo: 这里写死了, 应该从外部传入
		}
		log.Info("flowhelper--commitcase", msgModelApplyer)
		push.PublishMsg(msgModelApplyer, f.topic, f.producer)
	}

	return wf.JumpToStep(itemid, Status_Abandon, choice, remark, user)
}

//结束流程实列
func (f *FlowHelper) FinishCase(caseid, choice, remark string, itemid int32,
	appdata string) error {
	wf, err := New_Workflow(f.constr)
	if err != nil {
		return err
	}
	err = wf.LoadWorkflow(caseid, appdata)
	if err != nil {
		return err
	}
	user := &FlowUser{}
	return wf.JumpToStep(itemid, Status_Finished, choice, remark, user)
}

//流程实列, 退回到发起人
func (f *FlowHelper) SendbackCase(enterprise, usernumber, caseid, choice, remark string, itemid int32,
	appdata string) error {
	wf, err := New_Workflow(f.constr)
	if err != nil {
		return err
	}
	err = wf.LoadWorkflow(caseid, appdata)
	if err != nil {
		return err
	}
	//用流程的发起人, 步骤上从0开始的
	ni := wf.Fcase.CaseItems[0]
	user := &FlowUser{Userid: ni.HandleUserid, UserName: ni.HandleUserName}

	//发起人
	applyer := wf.Fcase.CaseInfo.CreatorId
	//发起人姓名
	creatorName := wf.Fcase.CaseInfo.CreatorName
	//抄送人
	copyuser := wf.Fcase.CaseInfo.CopyUser
	//业务实体Id
	biz_1 := wf.Fcase.CaseInfo.BizId1
	//当前步骤处理人
	handlerName := wf.Fcase.CaseInfo.HandleUserName
	//流程名
	flowname := wf.FlowDef.FlowName
	//流程序号
	serialnumber := wf.Fcase.CaseInfo.SerialNumber

	usernumberInt, _ := strconv.Atoi(usernumber)
	//审批过程处理过的人
	users := make([]int, 0, 0)

	recordid2 := uuid.NewV4().String()
	recordid3 := uuid.NewV4().String()

	contentCopyer := fmt.Sprintf("%v退回了%v的%v【%v】申请，退回原因：%v", handlerName, creatorName, flowname, serialnumber, remark)
	contentApplyer := fmt.Sprintf("%v退回了您的%v【%v】申请，退回原因：%v", handlerName, flowname, serialnumber, remark)

	// 抄送人
	copyuserString := arrayToString(copyuser)
	sqlInsertMessage := "select * from crm_f_m_w_taskjob_message_insert($1, $2, $3, $4, $5, $6, $7, $8);"
	sqlUpdateMessage := "select crm_f_m_w_message_updatestatus($1, '00000000-0000-0000-0000-000000000000', $2, 1);"
	itemidString := strconv.Itoa(int(itemid + 1)) //往后推一步（退回是审批的一种特殊情况）
	parameters := biz_1 + "," + itemidString

	conn, err := pgx.Connect(*f.conCfg)
	if err != nil {
		return err
	}
	defer conn.Close()

	if copyuserString != "" {
		if _, err = conn.Exec(sqlInsertMessage, recordid2, caseid, contentCopyer, parameters, 11, usernumberInt, time.Now().Format(f_datetime), copyuserString); err != nil {
			//跳过把？？？
			log.Error("workflow.flowhelper.PushMessage.Copyer", err.Error())
		}
		if _, err = conn.Exec(sqlUpdateMessage, usernumber, recordid2); err != nil {
			//跳过把？？？
			log.Error("workflow.flowhelper.PushMessage.Copyer", err.Error())
		}
	}
	// 发起人
	if _, err = conn.Exec(sqlInsertMessage, recordid3, caseid, contentApplyer, parameters, 9, usernumberInt, time.Now().Format(f_datetime), applyer); err != nil {
		//跳过把？？？
		log.Error("workflow.flowhelper.PushMessage.Applyer", err.Error())
	}
	if _, err = conn.Exec(sqlUpdateMessage, usernumber, recordid3); err != nil {
		//跳过把？？？
		log.Error("workflow.flowhelper.PushMessage.Applyer", err.Error())
	}

	msg_code := 1101

	//消息编码1101,消息id,流程实例id,流程类型11
	push_contentCopyer := fmt.Sprintf("%v,%v,%v,%v,%v", strconv.Itoa(msg_code), recordid2, caseid, "11", parameters)
	copyuser = removeFromArray(copyuser, usernumberInt)
	msgModelCopyer := &util.WfMessage{
		EnterNumber:  enterprise,
		FlowId:       caseid,
		Msg_key:      recordid2,
		Msg_title:    "您收到了一条审批消息",
		Content:      contentCopyer,
		Push_content: push_contentCopyer,
		Usernumber:   usernumberInt,
		Messagetype:  11,
		// Isstore:      0,
		MessageCode: msg_code,
		Sendtime:    time.Now().Format(f_datetime),
		Applyer:     "",
		Handler:     "",
		Agent:       "", //代理人e号
		Users:       users,
		MUsers:      copyuser,
		SystemCode:  "SYS00011", // todo: 这里写死了, 应该从外部传入
	}
	log.Info("flowhelper--commitcase", msgModelCopyer)
	push.PublishMsg(msgModelCopyer, f.topic, f.producer)

	//消息编码1101,消息id,流程实例id,流程类型11
	if applyer != usernumber {
		push_contentApplyer := fmt.Sprintf("%v,%v,%v,%v,%v", strconv.Itoa(msg_code), recordid3, caseid, "10", parameters)
		msgModelApplyer := &util.WfMessage{
			EnterNumber:  enterprise,
			FlowId:       caseid,
			Msg_key:      recordid3,
			Msg_title:    "您收到了一条审批消息",
			Content:      contentApplyer,
			Push_content: push_contentApplyer,
			Usernumber:   usernumberInt,
			Messagetype:  11,
			// Isstore:      0,
			MessageCode: msg_code,
			Sendtime:    time.Now().Format(f_datetime),
			Applyer:     applyer,
			Handler:     "",
			Agent:       "", //代理人e号
			Users:       users,
			MUsers:      users,
			SystemCode:  "SYS00011", // todo: 这里写死了, 应该从外部传入
		}
		log.Info("flowhelper--commitcase", msgModelApplyer)
		push.PublishMsg(msgModelApplyer, f.topic, f.producer)
	}

	return wf.JumpToStep(itemid, Status_Start, choice, remark, user)
}

//流程实列, 退回给上一个步骤
func (f *FlowHelper) FallbackCase(caseid, choice, remark string, itemid int32,
	appdata string) error {
	wf, err := New_Workflow(f.constr)
	if err != nil {
		return err
	}
	err = wf.LoadWorkflow(caseid, appdata)
	if err != nil {
		return err
	}
	//用流程上一步骤的处理人, 步骤上从0开始的
	ilen := len(wf.Fcase.CaseItems)
	ni := wf.Fcase.CaseItems[int32(ilen-2)]
	user := &FlowUser{Userid: ni.HandleUserid, UserName: ni.HandleUserName}
	return wf.JumpToStep(itemid, ni.StepName, choice, remark, user)
}

//标记流程步骤为已读
func (f *FlowHelper) Readed(itemid int32, caseid, usernumber string) error {
	conn, err := pgx.Connect(*f.conCfg)
	if err != nil {
		return err
	}
	defer conn.Close()
	sql := `update crm_t_flowcaseitem set stepstatus=1
	where caseid=$1 and itemid=$2 and (handleuserid=$3 or agentuserid=$3)`
	tag, err := conn.Exec(sql, caseid, itemid, usernumber)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no item readed, caseid:%s, itemid:%v, user:%s",
			caseid, itemid, usernumber)
	}
	return nil
}

//设置代理人
func (f *FlowHelper) SetAgent(userid, agentid string) error {
	conn, err := pgx.Connect(*f.conCfg)
	if err != nil {
		return err
	}
	defer conn.Close()
	sql1 := `delete from crm_t_flowagent where userid=$1`
	_, err = conn.Exec(sql1, userid)
	if err != nil {
		return err
	}
	sql2 := `insert into crm_t_flowagent(userid, agentid) values($1, $2)`
	_, err = conn.Exec(sql2, userid, agentid)
	if err != nil {
		return err
	}
	return nil
}

//取消代理人
func (f *FlowHelper) UnsetAgent(userid string) error {
	conn, err := pgx.Connect(*f.conCfg)
	if err != nil {
		return err
	}
	defer conn.Close()
	sql1 := `delete from crm_t_flowagent where userid=$1`
	_, err = conn.Exec(sql1, userid)
	if err != nil {
		return err
	}
	return nil
}

//动态获取审批选项
func (f *FlowHelper) GetDynamicSel(flowid, stepname string) ([]*Choice, error) {
	wf, err := New_Workflow(f.constr)
	if err != nil {
		return nil, err
	}
	flow, err := wf.FlowProvid.GetFlow(flowid)
	if err != nil {
		return nil, err
	}
	choice := make([]*Choice, 0, 1)
	for _, status := range flow.FlowStatus {
		if stepname == status.Name {
			for _, cho := range status.Choices {
				choice = append(choice, cho)
			}
		}
	}
	return choice, nil
}

//-----------------------------------------流程定义的方法-----------------------------------------

//流程定义详情
func (f *FlowHelper) GetFlow(flowid string) (*FlowInfo, error) {
	if flowid == "" {
		return nil, errors.New("flowid is empty")
	}
	conn, err := pgx.Connect(*f.conCfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	sql := `select name, descript, flowxml, stepcount, createtime, createusername,status
	from crm_t_workflow where flowid=$1 limit 1`
	row := conn.QueryRow(sql, flowid)
	var name, desc, flowxml, creator string
	var createtime time.Time
	var stepcount, status int32
	if err := row.Scan(&name, &desc, &flowxml, &stepcount, &createtime, &creator, &status); err != nil {
		return nil, err
	}
	fi := &FlowInfo{
		FlowId:     flowid,
		Name:       name,
		Descript:   desc,
		FlowXml:    flowxml,
		StepCount:  stepcount,
		CreateTime: createtime,
		Creator:    creator,
		Status:     status,
	}

	return fi, nil
}

//保存一个新的流程定义
func (f *FlowHelper) AddFlow(flow *FlowInfo, appid string) error {
	//判断流程名称重复
	validate_sql := `select cast(count(1) as integer) from crm_t_workflow where name = $1`
	validate_conn, _ := pgx.Connect(*f.conCfg)
	defer validate_conn.Close()

	row := validate_conn.QueryRow(validate_sql, flow.Name)
	var count int32
	validate_err := row.Scan(&count)
	if validate_err != nil {
		return validate_err
	}

	if count > 0 {
		return errors.New("流程名称已存在")
	}
	sql := `INSERT INTO crm_t_workflow(flowid, name, descript, flowxml, stepcount, createtime, createusername,
		flowcategory,updatetime,updateusername)
	  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);`
	//写入流程信息
	conn, _ := pgx.Connect(*f.conCfg)
	defer conn.Close()
	tx, err1 := conn.Begin()
	if err1 != nil {
		log.Error("flowhelper.AddFlow", err1.Error())
		return err1
	}
	defer tx.Rollback()

	_, err := conn.Exec(sql, flow.FlowId, flow.Name, flow.Descript, flow.FlowXml,
		flow.StepCount, flow.CreateTime, flow.Creator, flow.FlowCategory, flow.CreateTime, flow.Creator)
	if err != nil {
		return err
	}
	//写入表单对象
	sql_app := `insert into crm_t_appflow(appid, flowid) values($1, $2)`
	if _, err := conn.Exec(sql_app, appid, flow.FlowId); err != nil {
		return err
	}

	err2 := tx.Commit()
	if err2 != nil {
		return err2
	}
	return nil
}

//删除一个流程定义
func (f *FlowHelper) DeleteFlow(flowid string) error {
	conn, _ := pgx.Connect(*f.conCfg)
	defer conn.Close()
	tx, err1 := conn.Begin()
	if err1 != nil {
		log.Error("flowhelper.DeleteFlow", err1.Error())
		return err1
	}
	defer tx.Rollback()
	//判断流程实例是否已走完
	sql1 := "select cast(count(1) as integer) from crm_t_flowcase where flowid= $1  limit 1"
	row := conn.QueryRow(sql1, flowid)
	var result int32
	if err := row.Scan(&result); err != nil {
		return err
	}
	if result != 0 {
		return errors.New("该流程已存在流程的实例，不能删除")
	}
	//删除流程表单关系
	// sql3 := "delete from crm_t_appflow where flowid=$1"
	// if _, err := conn.Exec(sql3, flowid); err != nil {
	// 	return err
	// }
	//删除流程版本
	// sql4 := "delete from crm_t_workflow_version where flowid=$1"
	// if _, err := conn.Exec(sql4, flowid); err != nil {
	// 	return err
	// }

	//删除外键关联表todo

	//删除流程表
	sql2 := "delete from crm_t_workflow where flowid=$1"
	if _, err := conn.Exec(sql2, flowid); err != nil {
		return err
	}

	err2 := tx.Commit()
	if err2 != nil {
		return err2
	}
	return nil
}

//修改一个流程定义
func (f *FlowHelper) UpdateFlow(flow *FlowInfo) error {
	//判断流程名称重复
	validate_sql := `select cast(count(1) as integer) from crm_t_workflow where name = $1 and flowid <> $2`
	validate_conn, _ := pgx.Connect(*f.conCfg)
	defer validate_conn.Close()

	row := validate_conn.QueryRow(validate_sql, flow.Name, flow.FlowId)
	var count int32
	validate_err := row.Scan(&count)
	if validate_err != nil {
		return validate_err
	}

	if count > 0 {
		return errors.New("流程名称已存在")
	}

	conn, err := pgx.Connect(*f.conCfg)
	if err != nil {
		return err
	}
	defer conn.Close()
	tx, err1 := conn.Begin()
	if err1 != nil {
		log.Error("flowhelper.UpdateFlow", err1.Error())
		return err1
	}
	old_flow, err := f.GetFlow(flow.FlowId)
	if err != nil {
		return err
	}
	//版本历史
	sql_version := `insert into crm_t_workflow_version
	(versionid, flowid, name, flowxml, updatetime, updateusername, versionno)
	values($1,$2,$3,$4,$5,$6,(select versionno from crm_t_workflow where flowid = $7 limit 1))`
	if _, err := conn.Exec(sql_version, uuid.NewV4().String(), old_flow.FlowId, old_flow.Name, old_flow.FlowXml,
		flow.UpdateTime, flow.Updator, old_flow.FlowId); err != nil {
		return err
	}
	sql := `UPDATE crm_t_workflow set name = $1, descript = $2, flowxml = $3, stepcount = $4,
	updatetime = $5,updateusername = $6,versionno = versionno + 1
	where flowid = $7`

	if _, err := conn.Exec(sql, flow.Name, flow.Descript, flow.FlowXml, flow.StepCount, flow.UpdateTime,
		flow.Updator, flow.FlowId); err != nil {
		return err
	}

	err2 := tx.Commit()
	if err2 != nil {
		return err2
	}
	return nil
}

//启用流程
func (f *FlowHelper) EnableFlow(flow *FlowInfo) error {
	status := 1
	sql := `UPDATE crm_t_workflow set status = $1
	where flowid::text = any(string_to_array($2,','))`
	conn, err := pgx.Connect(*f.conCfg)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.Exec(sql, status, flow.FlowId); err != nil {
		return err
	}
	return nil
}

//停用流程
func (f *FlowHelper) DisableFlow(flow *FlowInfo) error {
	status := 0
	sql := `UPDATE crm_t_workflow set status = $1
	where flowid::text = any(string_to_array($2,','))`
	conn, err := pgx.Connect(*f.conCfg)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.Exec(sql, status, flow.FlowId); err != nil {
		return err
	}

	return nil
}

func (f *FlowHelper) WBStepStatus(itemid int32, caseid, usernumber string) error {
	log.Info("WBStepStatus.itemid", itemid)
	log.Info("WBStepStatus.caseid", caseid)
	log.Info("WBStepStatus.usernumber", usernumber)

	wf, err := New_Workflow(f.constr)
	if err != nil {
		return err
	}
	flowcase, err := wf.FlowProvid.LoadFlowCase(caseid)
	if err != nil {
		return err
	}

	if flowcase.CaseInfo.ItemId == itemid &&
		flowcase.CaseInfo.HandleUserid != usernumber {
		log.Debug("fh.WBStepStatus", "have no power to writeback")
		return nil
	}

	sql := `UPDATE crm_t_flowcaseitem set stepstatus = $1
	where caseid = $2 and itemid = $3 and itemid <> 0 and stepstatus = 0`
	conn, err := pgx.Connect(*f.conCfg)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.Exec(sql, 1, caseid, itemid); err != nil {
		return err
	}
	return nil
}

func arrayToString(array []int) string {
	str := ""
	for index, item := range array {
		if index == 0 {
			str = strconv.Itoa(item)
		} else {
			str = str + "," + strconv.Itoa(item)
		}
	}

	str = strings.Trim(str, ",")
	return str
}

func removeFromArray(array []int, elem int) []int {
	i := 0
	found := false
	for index, item := range array {
		if item == elem {
			i = index
			found = true
			break
		}
	}
	if found {
		array = append(array[:i], array[i+1:]...)
	}
	return array
}

func (f *FlowHelper) GetWorkFlowStatus(appid, bizids string, wfstatus int32) (*PluginWorkFlowInfoList, error) {
	// fmt.Println("GetWorkFlowStatus")
	// fmt.Println(appid)
	// fmt.Println(bizids)

	todolist := &PluginWorkFlowInfoList{make([]*PluginWorkFlowInfo, 0, 10), 0}

	sql := `select lower(af.bussinessid_1),fc.status from crm_t_appflowcase af 
left join crm_t_flowcase fc on af.caseid = fc.caseid
where appid = $1 and lower(bussinessid_1) = any(string_to_array(lower($2),','))`

	conn, err := pgx.Connect(*f.conCfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	rows, err := conn.Query(sql, appid, bizids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		af := &PluginWorkFlowInfo{}
		var bizid pgx.NullString
		var status pgx.NullInt32
		err := rows.Scan(&bizid, &status)
		if err != nil {
			return nil, err
		}
		if bizid.Valid {
			af.BizId = bizid.String
		}
		if status.Valid == true {
			switch int(status.Int32) {
			case 0:
				af.WorkFlowStatus = "审批中"
				break
			case 1:
				af.WorkFlowStatus = "已通过"
				break
			case 2:
				af.WorkFlowStatus = "已中止"
				break
			default:
				af.WorkFlowStatus = ""
			}

			if wfstatus != 5 && wfstatus != status.Int32 {
				continue
			}
		}

		todolist.Items = append(todolist.Items, af)
		todolist.TotalItems++

		//fmt.Println(todolist.TotalItems)
	}
	return todolist, nil
}

//插件流程对象
type PluginWorkFlowInfo struct {
	BizId          string `json:"bizid"`          //业务id
	WorkFlowStatus string `json:"workflowstatus"` //流程审批状态 0审批中1通过2不通过
}

type PluginWorkFlowInfoList struct {
	Items      []*PluginWorkFlowInfo
	TotalItems int32
}
