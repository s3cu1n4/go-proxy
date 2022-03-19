package logs

import (
	"os"
	"sync"
	"time"
)

var (
	_log = New()
)

func GetDefault() *logger {
	return _log
}

func New(opt ...*option) *logger {
	obj := new(logger)
	if len(opt) > 0 {
		obj.SetOption(opt[0])
	} else {
		obj.SetOption(NewOption())
	}
	obj.std = os.Stderr
	obj.fle = make(map[string]*fileLog, 0)
	obj.lock = new(sync.RWMutex)

	return obj
}

// 获取日志级别编号
func GetLevelNo(lvName string) int {
	if lv, ok := lvindex[lvName]; ok {
		return lv
	}
	return -1
}

// 获取日志级别名称
func GetLevelName(lv int) string {
	if name, ok := lvname[lv]; ok {
		return name
	}
	return ""
}

// 获取一个配置选项
func NewOption() *option {
	opt := option{
		Debug:    false,
		Level:    LvDebug,
		OutType:  OutTypeStd,
		LogPath:  "log/",
		LogSize:  104857600, // 100M
		LogCount: 5,
	}

	return &opt
}

// 设置选项
func SetOption(opt *option) {
	_log.SetOption(opt)
}

// 恐慌
func Panic(v ...interface{}) {
	_log.Panic(v...)
}
func Panicf(format string, v ...interface{}) {
	_log.Panicf(format, v...)
}

// 致命错误
func Fatal(v ...interface{}) {
	_log.Fatal(v...)
}
func Fatalf(format string, v ...interface{}) {
	_log.Fatalf(format, v...)
}

// 错误日志
func Error(v ...interface{}) {
	_log.Error(v...)
}
func Errorf(format string, v ...interface{}) {
	_log.Errorf(format, v...)
}

// 警告错误
func Warn(v ...interface{}) {
	_log.Warn(v...)
}
func Warnf(format string, v ...interface{}) {
	_log.Warnf(format, v...)
}

// 提示
func Notice(v ...interface{}) {
	_log.Notice(v...)
}
func Noticef(format string, v ...interface{}) {
	_log.Noticef(format, v...)
}

// 信息
func Info(v ...interface{}) {
	_log.Info(v...)
}
func Infof(format string, v ...interface{}) {
	_log.Infof(format, v...)
}

// 调试
func Debug(v ...interface{}) {
	_log.Debug(v...)
}
func Debugf(format string, v ...interface{}) {
	_log.Debugf(format, v...)
}

// 打印信息
func Print(v ...interface{}) {
	_log.Print(v...)
}
func Printf(format string, v ...interface{}) {
	_log.Printf(format, v...)
}

// 标准输出, 带时间
func Msg(v ...interface{}) {
	_log.Msg(v...)
}
func Msgf(format string, v ...interface{}) {
	_log.Msgf(format, v...)
}

// 标准输出, 不带时间
func ShowMsg(v ...interface{}) {
	_log.ShowMsg(v...)
}
func ShowMsgf(format string, v ...interface{}) {
	_log.ShowMsgf(format, v...)
}

// 记录到文件, 带时间
func File(file string, v ...interface{}) {
	_log.File(file, v...)
}
func Filef(file string, format string, v ...interface{}) {
	_log.Filef(file, format, v...)
}

// 记录到文件, 不带时间
func WriteFile(file string, v ...interface{}) {
	_log.WriteFile(file, v...)
}
func WriteFilef(file string, format string, v ...interface{}) {
	_log.WriteFilef(file, format, v...)
}

// 添加头
func appendHeader(buf *[]byte) {
	t := time.Now()
	y, m, d := t.Date()
	hour, min, sec := t.Clock()
	nsec := t.Nanosecond()
	appendNum(buf, y, 4)
	*buf = append(*buf, '/')
	appendNum(buf, int(m), 2)
	*buf = append(*buf, '/')
	appendNum(buf, d, 2)
	*buf = append(*buf, ' ')
	appendNum(buf, hour, 2)
	*buf = append(*buf, ':')
	appendNum(buf, min, 2)
	*buf = append(*buf, ':')
	appendNum(buf, sec, 2)
	*buf = append(*buf, '.')
	appendNum(buf, nsec, 3)
}
func appendNum(buf *[]byte, num, wid int) {
	var b [20]byte
	pos := len(b)
	nwid := wid
	for num > 0 || wid >= 1 {
		pos--
		wid--
		n := num % 10
		b[pos] = byte('0' + n)
		num = num / 10
	}
	*buf = append(*buf, b[pos:pos+nwid]...)
}
