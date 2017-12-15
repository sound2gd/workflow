package engine

import (
	"errors"
	"github.com/jteeuwen/go-pkg-xmlx"
	"github.com/widuu/gojson"
	"strconv"
)

//比较符常量
const (
	OpEqual     uint32 = iota //等于
	OpLess                    // <
	OpGreater                 // >
	OpLessEq                  // <=
	OpGreaterEq               // >=
	OpNotEq                   // !=
)

//逻辑操作符
const (
	LogicAnd uint32 = iota // and
	LogicOr                // or
)

// Condition 条件
type Condition struct {
	GetNot   bool   //是否取反
	Operator uint32 //比较操作符
	DataKey  string //业务数据取值属性
	Value    string //比较目标值
	Logic    uint32 //逻辑操作符
}

// NewCondition 根据xml描述创建一个新的节点
func NewCondition(n *xmlx.Node) (*Condition, error) {
	cd := &Condition{}
	cd.GetNot = n.Ab("", "getnot")
	cd.DataKey = n.As("", "datakey")
	cd.Value = n.As("", "value")

	op := n.As("", "op")
	switch op {
	case "eq":
		cd.Operator = OpEqual
	case "less":
		cd.Operator = OpLess
	case "greator":
		cd.Operator = OpGreater
	case "lesseq":
		cd.Operator = OpLessEq
	case "greatoreq":
		cd.Operator = OpGreaterEq
	case "noteq":
		cd.Operator = OpNotEq
	default:
		return nil, errors.New("Operator not supported")
	}

	lp := n.As("", "lop")
	if lp == "and" || lp == "" {
		cd.Logic = LogicAnd
	} else if lp == "or" {
		cd.Logic = LogicOr
	} else {
		return nil, errors.New("logic op not supported")
	}
	return cd, nil
}

// Eval 计算条件结果
// appdataJSON: appdata 的json字符串
// 返回bool的结果
func (c *Condition) Eval(appdataJSON string) (bool, error) {
	appdata := gojson.Json(appdataJSON)
	if appdata.IsValid() == false {
		return false, errors.New("appdata parse to json failed. ")
	}

	re := false
	//todo: 这里目前只支持1级属性的获取, 以后可以扩展成: p1.pp2.ppp3,多级属性获取.
	dataValue := appdata.Get(c.DataKey)
	if dataValue.IsValid() == false {
		return false, errors.New("key not found in appdata ")
	}
	dv := dataValue.Tostring()
	//先判断等于和不等于两种情况, 用字符串格式比较
	if c.Operator == OpEqual {
		re = (dv == c.Value)
		return re, nil
	}
	if c.Operator == OpNotEq {
		re = (dv != c.Value)
		return re, nil
	}
	if dv != "" && c.Value != "" {
		//其他情况, 用float64比较, 只能比较大小
		ft, err := strconv.ParseFloat(dv, 64)
		if err != nil {
			return re, err
		}
		fv, err := strconv.ParseFloat(c.Value, 64)
		if err != nil {
			return re, err
		}
		switch c.Operator {
		case OpGreater:
			re = (ft > fv)
		case OpGreaterEq:
			re = (ft >= fv)
		case OpLess:
			re = (ft < fv)
		case OpLessEq:
			re = (ft <= fv)
		default:
			return false, errors.New("condition Operator not supported")
		}
	}
	//取反操作
	if c.GetNot {
		re = !re
	}
	return re, nil
}
