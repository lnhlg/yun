package yun

//Mode 运行模式
type Mode int

const (
	//DEBUG 调试模式
	DEBUG	Mode = iota
	//TEST 测试模式
	TEST
	//RELEASE 发行模式
	RELEASE
)