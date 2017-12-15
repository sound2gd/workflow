package workflow

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx"
)

//-----------------------------------------------------
type OrgPgProvider struct {
	ConnCfg *pgx.ConnConfig //数据库连接串
}

func New_OrgPgProvider(connstr string) (*OrgPgProvider, error) {
	cfg, err := util.GetConnCfg(connstr)
	if err != nil {
		return nil, err
	}

	p := &OrgPgProvider{
		ConnCfg: cfg,
	}
	return p, nil
}

func (o *OrgPgProvider) FindUser(role, departid string) ([]*FlowUser, error) {
	log.Debug("workflow.orgprovider.FindUser", "role:%s,deptid:%s", role, departid)
	conn, err := pgx.Connect(*o.ConnCfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	sql := `select u.usernumber,u.username,u.logourl as headphoto from com_t_userinfo u
      inner join com_t_department d on u.departmentid = d.departmentid
      inner join com_t_userrole ur on u.usernumber = ur.usernumber
      inner join com_t_role r on ur.roleid = r.roleid
      where r.rolename = $1 and d.departmentid = $2
      and u.status = 1 `

	rows, err := conn.Query(sql, role, departid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	//一般不会超过10个
	fus := make([]*FlowUser, 0, 10)
	for rows.Next() {
		var una string
		var uid int32
		var headphoto pgx.NullString
		if err := rows.Scan(&uid, &una, &headphoto); err != nil {
			return nil, err
		}
		userid := fmt.Sprintf("%v", uid)
		fu := &FlowUser{Userid: userid, UserName: una}
		if headphoto.Valid {
			fu.HeadPhoto = headphoto.String
		}
		fus = append(fus, fu)
	}
	return fus, nil
}

func (o *OrgPgProvider) FindUserDept(userid string) (string, error) {
	conn, err := pgx.Connect(*o.ConnCfg)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	sql := `select departmentid from com_t_userinfo u where usernumber = $1 and u.status = 1 limit 1`
	uid, _ := strconv.Atoi(userid)
	row := conn.QueryRow(sql, uid)
	var dpid string
	if err := row.Scan(&dpid); err != nil {
		return "", err
	}
	return dpid, nil
}

func (o *OrgPgProvider) FindUserParentDept(userid string) (string, error) {
	conn, err := pgx.Connect(*o.ConnCfg)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	sql := `select d.pdepartmentid from com_t_userinfo u
        inner join com_t_department d on u.departmentid = d.departmentid
        where u.usernumber = $1 and u.status = 1 `
	uid, _ := strconv.Atoi(userid)
	row := conn.QueryRow(sql, uid)
	var pdid string
	if err := row.Scan(&pdid); err != nil {
		return "", err
	}
	return pdid, nil
}

func (o *OrgPgProvider) GetUser(userid string) (us []*FlowUser, err error) {
	//fmt.Println("userid: ", userid)
	conn, err := pgx.Connect(*o.ConnCfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	sql := `select usernumber,username,logourl as headphoto from com_t_userinfo u
	where usernumber::text = any(string_to_array($1,',')) and u.status = 1`
	//fmt.Println("sql: ", sql)
	rows, err := conn.Query(sql, userid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	count := len(strings.Split(userid, ","))
	fus := make([]*FlowUser, 0, count)
	for rows.Next() {
		var uid int32
		var una string
		var headphoto pgx.NullString
		if err := rows.Scan(&uid, &una, &headphoto); err != nil {
			return nil, err
		}

		fu := &FlowUser{Userid: strconv.Itoa(int(uid)), UserName: una}
		if headphoto.Valid {
			fu.HeadPhoto = headphoto.String
		}
		fus = append(fus, fu)
	}

	return fus, nil
}

func (o *OrgPgProvider) FindUserByDept(departid string) (us []*FlowUser, err error) {
	conn, err := pgx.Connect(*o.ConnCfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	sql := `select u.usernumber,u.username,u.logourl from com_t_userinfo u
      inner join com_t_department d on u.departmentid = d.departmentid
      where d.departmentid = $1 and u.status = 1`
	rows, err := conn.Query(sql, departid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	//一般不会超过30个
	fus := make([]*FlowUser, 0, 30)
	for rows.Next() {
		var una string
		var uid int32
		var headphoto pgx.NullString
		if err := rows.Scan(&uid, &una, &headphoto); err != nil {
			return nil, err
		}
		userid := fmt.Sprintf("%v", uid)
		fu := &FlowUser{Userid: userid, UserName: una}
		if headphoto.Valid {
			fu.HeadPhoto = headphoto.String
		}
		fus = append(fus, fu)
	}
	return fus, nil
}

func (o *OrgPgProvider) FindUserByRole(role string) (us []*FlowUser, err error) {
	conn, err := pgx.Connect(*o.ConnCfg)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	sql := `select u.usernumber,u.username,u.logourl as headphoto from com_t_userinfo u
      inner join com_t_userrole ur on u.usernumber = ur.usernumber
      inner join com_t_role r on ur.roleid = r.roleid
      where r.rolename = $1 and u.status = 1`
	//fmt.Println("FindUserByRole")
	//fmt.Println(sql)
	rows, err := conn.Query(sql, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	//一般不会超过30个
	fus := make([]*FlowUser, 0, 30)
	for rows.Next() {
		var una string
		var uid int32
		var headphoto pgx.NullString
		if err := rows.Scan(&uid, &una, &headphoto); err != nil {
			return nil, err
		}
		userid := fmt.Sprintf("%v", uid)
		fu := &FlowUser{Userid: userid, UserName: una}
		if headphoto.Valid {
			fu.HeadPhoto = headphoto.String
		}
		fus = append(fus, fu)
	}
	return fus, nil
}
