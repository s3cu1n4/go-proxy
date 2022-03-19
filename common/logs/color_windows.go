// +build windows

package logs

import (
	"os"
	"syscall"
)

var (
	kernel32    *syscall.LazyDLL  = syscall.NewLazyDLL("kernel32.dll")
	proc        *syscall.LazyProc = kernel32.NewProc("SetConsoleTextAttribute")
	closeHandle *syscall.LazyProc = kernel32.NewProc("CloseHandle")

	// 颜色编号
	// 前景色 0 黑色, 1 蓝色, 2 绿色, 3 青色, 4 红色, 5 紫色, 6 黄色, 7 淡灰色(系统默认值), 8 灰色, 9 亮蓝色, 10 亮绿色, 11 亮青色, 12 亮红色, 13 亮紫色, 14 亮黄色, 15 白色
	// 背景色 16 蓝色, 32 绿色, 64 红色 ... 低4位前景色,高4位背景色
	lvcolor = map[int]int{
		LvFatal: 12,
		LvError: 4,
		LvWarn:  6,
	}
)

func printColorText(out *os.File, lv int, content []byte) {
	if color, ok := lvcolor[lv]; ok {
		handle, _, _ := proc.Call(out.Fd(), uintptr(color))
		out.Write(content)
		handle, _, _ = proc.Call(out.Fd(), uintptr(7))
		closeHandle.Call(handle)
	} else {
		out.Write(content)
	}
}
