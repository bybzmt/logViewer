package protocol

type CMD uint8

type VARCHAR struct {
	Len  uint16
	Data []uint8
}

const (
	CMD_EXIT CMD = iota
	CMD_PING
	CMD_PONG
	//列出文件列表
	CMD_LIST
	//打开文件 VARCHAR
	CMD_OPEN
	//启动操作
	CMD_START
	//操作结束
	CMD_END
	//取消所有动作
	CMD_CANCEL
	//时间段 start(int64) + end(int64)
	CMD_TIME
	//seek uint64
	CMD_SEEK
	//过滤字符串 bind(uint16) + VARCHAR
	CMD_GREP
	CMD_GREP_OR
	CMD_REGEXP
	//进度报告 bind(uint16)
	CMD_PROGRESS
	//查找方向
	CMD_REVERSE
	//找查数量
	CMD_LIMIT
	//查换速度
	CMD_SPEED
	//跟随文件变化
	CMD_TAILF
	//删除文件
	CMD_RM
	CMD_APPEND
	CMD_TOFILE
)
