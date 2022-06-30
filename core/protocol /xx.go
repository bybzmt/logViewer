package protocol

type OP uint16

type VarChar struct {
	Len  uint16
	Data []uint8
}
type CharArray struct {
	Len  uint16
	Data []VarChar
}

const (
	OP_EXIT OP = iota
	OP_PING
	MSG_PONG
	//列出文件列表
	OP_LIST
	//列表响应文件状态
	MSG_STATE
	//删除文件
	OP_RM
	MSG_OK
	//行起始确定
	OP_LINE_MATCH
	OP_LINE_REGEX
	//打开文件
	OP_OPEN
	RESP_OPEN
	//启动操作
	OP_START
	//响应结束
	MSG_END
	//取消动作
	OP_CANCEL
	//时间段 start(int64) + end(int64)
	OP_TIME
	//seek uint64
	OP_SEEK
	//过滤字符串
	OP_MATCH
	OP_MATCH_OR
	OP_REGEXP
	//进度报告
	OP_PROGRESS
	MSG_PROGRESS
	//查找方向
	OP_REVERSE
	//找查数量
	OP_LIMIT
	//输出速度
	OP_SPEED
	//跟随文件变化
	OP_TAILF
	OP_TOFILE
	//安静模式
	OP_QUIET
	MSG_LINE
	//统计器
	OP_COUNT_MATCH
	OP_COUNT_REGEX
	OP_COUNT
	MSG_COUNT
)
