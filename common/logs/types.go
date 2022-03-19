package logs

import (
	"os"
)

// 日志级别
const (
	lvOut  = -1
	LvNone = iota
	lvPanic
	LvFatal
	LvError
	LvWarn
	LvNotice
	LvInfo
	LvDebug

	LevelNone   = "none"
	LevelPanic  = "panic"
	LevelFatal  = "fatal"
	LevelError  = "error"
	LevelWarn   = "warn"
	LevelNotice = "notice"
	LevelInfo   = "info"
	LevelDebug  = "debug"
)

// 输出方式
const (
	OutTypeAll = iota
	OutTypeStd
	OutTypeFile
)

// 输出时间
const (
	printTimeNone = 0
	printTimeShow = 1
)

// 日志配置选项
type option struct {
	Debug     bool           // 启用Debug
	Level     int            // 日志级别
	OutType   int            // 输出方式
	LogPath   string         // 日志保存路径
	LogSize   int64          // 日志文件大小
	LogCount  int            // 保留日志文件数
	LogFile   string         // 默认日志文件名
	LvLogFile map[int]string // 日志级别文件
}

// 文件日志
type fileLog struct {
	File *os.File
	Time int64
}

// 自定日志函数
// out 输入方式
// lvN 日志级别
// lvName 级别名称
// v ... 日志内容
type LogFunc func(out, lvN int, lvName string, v ...interface{})
