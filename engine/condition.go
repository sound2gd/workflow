package workflow

import (
	"errors"
	// "fmt"
	"github.com/jteeuwen/go-pkg-xmlx"
	"github.com/widuu/gojson"
	"strconv"
	log "xtion.net/mcrm/logger"
)

const (
	Op_Equal uint32 = iota
	Op_Less
	Op_Greater
	Op_LessEq
	Op_GreaterEq
	Op_NotEq
)

const (
	Logic_And uint32 = iota
	Logic_Or
)

//条件
type Condition struct {
	GetNot   bool   //是否取反
	Operator uint32 //比较操作符
	DataKey  string //业务数据取值属性
	Value    string //比较目标值
	Logic    uint32 //逻辑操作符
}

func New_Condition(n *xmlx.Node) (*Condition, error) {
	cd := &Condition{}
	cd.GetNot = n.Ab("", "getnot")
	cd.DataKey = n.As("", "datakey")
	cd.Value = n.As("", "value")

	op := n.As("", "op")
	switch op {
	case "eq":
		cd.Operator = Op_Equal
	case "less":
		cd.Operator = Op_Less
	case "greator":
		cd.Operator = Op_Greater
	case "lesseq":
		cd.Operator = Op_LessEq
	case "greatoreq":
		cd.Operator = Op_GreaterEq
	case "noteq":
		cd.Operator = Op_NotEq
	default:
		return nil, errors.New("Operator not supported")
	}

	lp := n.As("", "lop")
	if lp == "and" || lp == "" {
		cd.Logic = Logic_And
	} else if lp == "or" {
		cd.Logic = Logic_Or
	} else {
		return nil, errors.New("logic op not supported")
	}
	return cd, nil
}

//计算条件结果
func (c *Condition) Eval(appdataJson string) (bool, error) {
	appdata := make(map[string]string)
	//json ---> map[string]string
	log.Debug("condition.Eval.appdata", appdata)
	//todo: 这里目前只支持1级属性的获取, 以后可以扩展成: p1.pp2.ppp3,多级属性获取.
	appdata = gojson.Json(appdataJson).GetDataFirstLevel()

	re := false
	tv, ok := appdata[c.DataKey]
	if !ok {
		log.Error("condition.Eval", "not find in appdata: ", c.DataKey)
		return re, errors.New("not find in appdata: " + c.DataKey)
	}
	//先判断等于和不等于两种情况, 用字符串比较
	if c.Operator == Op_Equal {
		re = (tv == c.Value)
		return re, nil
	}
	if c.Operator == Op_NotEq {
		re = (tv != c.Value)
		return re, nil
	}
	if tv == "" {
		log.Error("condition.Eval", "tv is empty")
		return re, nil
	}
	//其他情况, 用float64比较
	ft, err := strconv.ParseFloat(tv, 64)
	if err != nil {
		return re, err
	}

	fv, err := strconv.ParseFloat(c.Value, 64)
	if err != nil {
		return re, err
	}
	switch c.Operator {
	case Op_Greater:
		re = (ft > fv)
	case Op_GreaterEq:
		re = (ft >= fv)
	case Op_Less:
		re = (ft < fv)
	case Op_LessEq:
		re = (ft <= fv)
	default:
		return false, errors.New("condition Operator not supported")
	}
	if c.GetNot {
		re = !re
	}
	return re, nil
}
