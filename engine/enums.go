package workflow

//-----------------------------------------------
// //流程状态, 没有这个东西
// type FlowStatus uint32

// //流程状态枚举
// const (
// 	//开始
// 	FlowStatus_Begin FlowStatus = iota
// 	//运行
// 	FlowStatus_Running
// 	//暂停
// 	FlowStatus_Pause
// 	//意外中止
// 	FlowStatus_Unexpected
// 	//结束
// 	FlowStatus_Finished
// )

//-----------------------------------------------
//流程参与者类型
type FlowActorType uint32

//流程参与者类型枚举
const (
	//创建人
	FlowActorType_Creator FlowActorType = iota
	//代理人
	FlowActorType_Agent
	//当前用户
	FlowActorType_CurrentUser
)

//-----------------------------------------------
//判断操作符
type Operator uint32

//判断操作符枚举
const (
	//相等
	Operator_OpEqual Operator = iota + 1
	//小于
	Operator_OpLess
	//大于
	Operator_OpLarge
	//小于等于
	Operator_OpLessEq
	//大于等于
	Operator_OpLargeEq
	//不等于
	Operator_OpNotEq
)

//-----------------------------------------------
//工作项状态
type WorkItemStatus uint32

//工作项状态枚举
const (
	//已接收
	WorkItemStatus_Received WorkItemStatus = iota
	//已读
	WorkItemStatus_Readed
	//已完成
	WorkItemStatus_Finished
	//自动完成
	//WorkItemStatus_AutoFinished
	//暂停
	//WorkItemStatus_Pause
)

//-----------------------------------------------
//环节迁出规则
type TransitOut uint32

//环节迁出规则枚举
const (
	//只有一个条件满足,就可以迁出
	One TransitOut = iota
	//所有条件满足,才可以迁出
	All
)

//-----------------------------------------------
//事件处理方式
type EventHandleType uint32

//事件处理方式枚举
const (
	//邮件
	EventHandleType_Email EventHandleType = iota
	//短信
	EventHandleType_SMS
	//消息推送
	EventHandleType_Message
	//数据接口
	EventHandleType_Datasource
	//工作项转移
	EventHandleType_transfer
	//任务计划
	EventHandleType_Schedule
	//产生新流程
	EventHandleType_Workflow
	//执行工作项
	EventHandleType_WorkItem
)

//-----------------------------------------------
//环节运行模式
type ActivityRunMode uint32

//环节运行模式枚举
const (
	//手动
	ActivityRunMode_Manual ActivityRunMode = iota
	//自动
	ActivityRunMode_Auto
)

//-----------------------------------------------
//迁入模式
type TransitJoinMode uint32

//迁入模式枚举
const (
	//与
	TransitJoinMode_AND TransitJoinMode = iota
	//异或
	TransitJoinMode_XOR
)

//-----------------------------------------------
//迁出模式
type TransitSplitMode uint32

//迁出模式枚举
const (
	//与
	TransitSplitMode_AND TransitSplitMode = iota
	//异或
	TransitSplitMode_XOR
)

//-----------------------------------------------
//并行流程,迁出的模式
type WaitMode uint32

//并行流程,迁出的模式
const (
	//有一个满足, 就ok
	WaitMode_One WaitMode = iota
	//全部都满足,才ok
	WaitMode_All
)

//-----------------------------------------------
////条件的结果
//type ConditionResult uint32

////条件的结果枚举
//const (
//	ConditionResult_True ConditionResult = iota
//	ConditionResult_false
//	ConditionResult_Unknown
//)
//-----------------------------------------------
