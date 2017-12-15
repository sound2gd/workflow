package engine

// FlowActorType 流程参与者类型
type FlowActorType uint32

//流程参与者类型枚举
const (
	//创建人
	FlowActorTypeCreator FlowActorType = iota
	//代理人
	FlowActorTypeAgent
	//当前用户
	FlowActorTypeCurrentUser
)

// WorkItemStatus 工作项状态
type WorkItemStatus uint32

//工作项状态枚举
const (
	//已接收
	WorkItemStatusReceived WorkItemStatus = iota
	//已读
	WorkItemStatusReaded
	//已完成
	WorkItemStatusFinished
	//自动完成
	//WorkItemStatusAutoFinished
	//暂停
	//WorkItemStatusPause
)

// TransitOut 环节迁出规则
type TransitOut uint32

//环节迁出规则枚举
const (
	//只有一个条件满足,就可以迁出
	One TransitOut = iota
	//所有条件满足,才可以迁出
	All
)

// ActivityRunMode 环节运行模式
type ActivityRunMode uint32

//环节运行模式枚举
const (
	//手动
	ActivityRunModeManual ActivityRunMode = iota
	//自动
	ActivityRunModeAuto
)

// TransitJoinMode 迁入模式
type TransitJoinMode uint32

//迁入模式枚举
const (
	//与
	TransitJoinModeAND TransitJoinMode = iota
	//异或
	TransitJoinModeXOR
)

// TransitSplitMode 迁出模式
type TransitSplitMode uint32

//迁出模式枚举
const (
	//与
	TransitSplitModeAND TransitSplitMode = iota
	//异或
	TransitSplitModeXOR
)

// WaitMode 并行流程,迁出的模式
type WaitMode uint32

//并行流程,迁出的模式
const (
	//有一个满足, 就ok
	WaitModeOne WaitMode = iota
	//全部都满足,才ok
	WaitModeAll
)

////条件的结果
//type ConditionResult uint32

////条件的结果枚举
//const (
//	ConditionResult_True ConditionResult = iota
//	ConditionResult_false
//	ConditionResult_Unknown
//)
