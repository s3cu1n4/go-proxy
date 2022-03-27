package logs

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	lvindex = map[string]int{
		LevelNone:   LvNone,
		LevelPanic:  lvPanic,
		LevelFatal:  LvFatal,
		LevelError:  LvError,
		LevelWarn:   LvWarn,
		LevelNotice: LvNotice,
		LevelInfo:   LvInfo,
		LevelDebug:  LvDebug,
	}

	lvname = map[int]string{
		LvNone:   "None",
		lvPanic:  "Panic",
		LvFatal:  "Fatal",
		LvError:  "Error",
		LvWarn:   "Warn",
		LvNotice: "Notice",
		LvInfo:   "Info",
		LvDebug:  "Debug",
	}

	outname = map[int]string{
		OutTypeAll:  "All",
		OutTypeStd:  "Std",
		OutTypeFile: "File",
	}

	_line_sep byte = '\n'
)

type logger struct {
	opt        *option
	fn         LogFunc
	fn_lv      int
	std        *os.File
	fle        map[string]*fileLog
	lock       *sync.RWMutex
	flush_time int64
}

// 自定义函数
func (this *logger) Func(lv int, fn LogFunc) {
	this.fn_lv = lv
	this.fn = fn
}

// 设置日志选项
func (this *logger) SetOption(opt *option) {
	if opt.LogPath != "" {
		opt.LogPath = strings.TrimRight(opt.LogPath, "\\/")
	}
	if opt.LogFile == "" {
		opt.LogFile = "logs.log"
	}

	this.opt = opt
}

/*
// 刷新/清理文件日志
func (this *logger) FlushLogFile(timeout int64) {
	tmv := time.Now().Unix()
	this.lock.Lock()
	defer this.lock.Unlock()
	for k, h := range this.fle {
		if tmv-h.Time > timeout {
			h.File.Close()
			delete(this.fle, k)
		}
	}
}
*/

// 获取日志选项
func (this *logger) GetOption() *option {
	return this.opt
}

// 关闭/启用 Debug
func (this *logger) EnableDebug(enable bool) {
	this.opt.Debug = enable
}

// 初始化日志文件
func (this *logger) getFileLog(file string) (*fileLog, error) {
	opt := this.opt

	var obj *fileLog = nil
	var info os.FileInfo
	var err error

	if v, ok := this.fle[file]; ok {
		obj = v
	}

	info, err = os.Stat(file)
	if os.IsExist(err) {
		if opt.LogSize > 0 && info.Size() >= opt.LogSize {
			if obj != nil {
				obj.File.Close()
				obj = nil
			}
			if opt.LogCount <= 1 {
				os.Rename(file, file+".bak")
			} else {
				arr := []string{file}
				for i := 0; i < opt.LogCount; i++ {
					fle := file + "." + strconv.Itoa(i)
					_, err = os.Stat(fle)
					if os.IsNotExist(err) {
						continue
					}
					arr = append(arr, fle)
				}
				for i := len(arr) - 1; i >= 0; i-- {
					os.Rename(arr[i], file+"."+strconv.Itoa(i))
				}
			}
		}
	} else {
		path := filepath.Dir(file)
		if _, err = os.Stat(path); os.IsNotExist(err) {
			if err = os.MkdirAll(path, 0777); err != nil {
				if obj != nil {
					obj.File.Close()
					delete(this.fle, file)
				}
				return nil, err
			}
		}
	}

	nowtime := time.Now().Unix()
	if obj == nil {
		fp, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			return nil, err
		}
		obj = &fileLog{
			File: fp,
			Time: nowtime,
		}
		this.fle[file] = obj
	} else {
		obj.Time = nowtime
	}

	var timeout int64 = 60
	if nowtime-this.flush_time > timeout {
		for k, h := range this.fle {
			if nowtime-h.Time > timeout {
				h.File.Close()
				delete(this.fle, k)
			}
		}
		this.flush_time = nowtime
	}

	return obj, nil
}

// 写文件
// file 文件名
// content 日志内容
func (this *logger) logFile(file string, content []byte) error {
	log, err := this.getFileLog(file)
	if err != nil {
		return err
	}
	_, err = log.File.Write(content)
	return err
}

// 标准输出
func (this *logger) logStd(lv int, content []byte) {
	printColorText(this.std, lv, content)
}

// 打印日志
// showtime 显示时间
// lv 日志级别
// tp 打印方式
// out_file 输出到指定文件
// new_line 是否新行
// v... 日志内容
func (this *logger) print(showtime, lv, tp int, out_file string, new_line bool, v ...interface{}) {
	opt := this.opt
	if opt.Debug == false && lv > opt.Level {
		return
	}

	this.lock.Lock()
	defer this.lock.Unlock()

	lv_name := lvname[lv]
	if this.fn != nil && lv <= this.fn_lv {
		this.fn(tp, lv, lv_name, v...)
		return
	}

	buf := make([]byte, 0)
	if showtime == printTimeShow {
		appendHeader(&buf)
	}
	if lv_name != "" {
		buf = append(buf, " ["+lv_name+"] "...)
	} else if len(buf) > 0 {
		buf = append(buf, ' ')
	}

	args := fmt.Sprint(v...)
	buf = append(buf, args...)
	if buf[len(buf)-1] != _line_sep {
		buf = append(buf, _line_sep)
	}

	if tp == OutTypeFile || tp == OutTypeAll {
		if out_file == "" {
			if fle, ok := opt.LvLogFile[lv]; ok && fle != "" {
				out_file = fle
			} else {
				out_file = opt.LogFile
			}
		}

		out_file = opt.LogPath + string(os.PathSeparator) + out_file
		err := this.logFile(out_file, buf)
		if err != nil {
			tmpa := make([]byte, 0)
			tmpa = append(tmpa, "Write log error("...)
			tmpa = append(tmpa, lv_name...)
			tmpa = append(tmpa, "): "...)
			tmpa = append(tmpa, err.Error()...)
			tmpa = append(tmpa, _line_sep)
			tmpa = append(tmpa, "      --------- Log Data ----------"...)
			tmpa = append(tmpa, _line_sep, _line_sep)

			tmpa = append(tmpa, buf...)
			this.logStd(lv, tmpa)
			return
		}
	}

	if tp == OutTypeStd || tp == OutTypeAll {
		this.logStd(lv, buf)
	}
	if lv == lvPanic {
		panic(args)
	}
}

// 恐慌
func (this *logger) Panic(v ...interface{}) {
	this.print(printTimeShow, lvPanic, this.opt.OutType, "", true, v...)
}
func (this *logger) Panicf(format string, v ...interface{}) {
	this.print(printTimeShow, lvPanic, this.opt.OutType, "", true, fmt.Sprintf(format, v...))
}

// 致命错误
func (this *logger) Fatal(v ...interface{}) {
	this.print(printTimeShow, LvFatal, this.opt.OutType, "", true, v...)
}
func (this *logger) Fatalf(format string, v ...interface{}) {
	this.print(printTimeShow, LvFatal, this.opt.OutType, "", true, fmt.Sprintf(format, v...))
}

// 错误日志
func (this *logger) Error(v ...interface{}) {
	this.print(printTimeShow, LvError, this.opt.OutType, "", true, v...)
}
func (this *logger) Errorf(format string, v ...interface{}) {
	this.print(printTimeShow, LvError, this.opt.OutType, "", true, fmt.Sprintf(format, v...))
}

// 警告错误
func (this *logger) Warn(v ...interface{}) {
	this.print(printTimeShow, LvWarn, this.opt.OutType, "", true, v...)
}
func (this *logger) Warnf(format string, v ...interface{}) {
	this.print(printTimeShow, LvWarn, this.opt.OutType, "", true, fmt.Sprintf(format, v...))
}

// 提示
func (this *logger) Notice(v ...interface{}) {
	this.print(printTimeShow, LvNotice, this.opt.OutType, "", true, v...)
}
func (this *logger) Noticef(format string, v ...interface{}) {
	this.print(printTimeShow, LvNotice, this.opt.OutType, "", true, fmt.Sprintf(format, v...))
}

// 信息
func (this *logger) Info(v ...interface{}) {
	this.print(printTimeShow, LvInfo, this.opt.OutType, "", true, v...)
}
func (this *logger) Infof(format string, v ...interface{}) {
	this.print(printTimeShow, LvInfo, this.opt.OutType, "", true, fmt.Sprintf(format, v...))
}

// 调试
func (this *logger) Debug(v ...interface{}) {
	this.print(printTimeShow, LvDebug, this.opt.OutType, "", true, v...)
}
func (this *logger) Debugf(format string, v ...interface{}) {
	this.print(printTimeShow, LvDebug, this.opt.OutType, "", true, fmt.Sprintf(format, v...))
}

// 打印信息
func (this *logger) Print(v ...interface{}) {
	this.print(printTimeShow, lvOut, this.opt.OutType, "", true, v...)
}
func (this *logger) Printf(format string, v ...interface{}) {
	this.print(printTimeShow, lvOut, this.opt.OutType, "", true, fmt.Sprintf(format, v...))
}

// 标准输出, 带时间
func (this *logger) Msg(v ...interface{}) {
	this.print(printTimeShow, lvOut, OutTypeStd, "", true, v...)
}
func (this *logger) Msgf(format string, v ...interface{}) {
	this.print(printTimeShow, lvOut, OutTypeStd, "", true, fmt.Sprintf(format, v...))
}

// 标准输出, 不带时间
func (this *logger) ShowMsg(v ...interface{}) {
	this.print(printTimeNone, lvOut, OutTypeStd, "", true, v...)
}
func (this *logger) ShowMsgf(format string, v ...interface{}) {
	this.print(printTimeNone, lvOut, OutTypeStd, "", true, fmt.Sprintf(format, v...))
}

// 记录到文件, 带时间
func (this *logger) File(file string, v ...interface{}) {
	this.print(printTimeShow, lvOut, OutTypeFile, file, true, v...)
}
func (this *logger) Filef(file string, format string, v ...interface{}) {
	this.print(printTimeShow, lvOut, OutTypeFile, file, true, fmt.Sprintf(format, v...))
}

// 记录到文件, 不带时间
func (this *logger) WriteFile(file string, v ...interface{}) {
	this.print(printTimeNone, lvOut, OutTypeFile, file, true, v...)
}
func (this *logger) WriteFilef(file string, format string, v ...interface{}) {
	this.print(printTimeNone, lvOut, OutTypeFile, file, true, fmt.Sprintf(format, v...))
}
