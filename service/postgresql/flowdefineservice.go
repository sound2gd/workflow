package workflow

import (
	"errors"
	"time"

	"github.com/jackc/pgx"
)

//-----------------------------------------流程定义的方法-----------------------------------------
type FlowHelper struct{}

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
