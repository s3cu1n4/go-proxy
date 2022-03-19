// +build !windows

package logs

import (
	"os"
)

var (
	// 颜色代码
	// 前景色 30 黑色, 31 红色, 32 绿色, 33 黄色, 34 蓝色, 35 紫红色, 36 青蓝色, 37 白色
	// 背景色 40 黑色, 41 红色, 42 绿色, 43 黄色, 44 蓝色, 45 紫红色, 46 青蓝色, 47 白色
	lvcolor = map[int][]byte{
		LvFatal: []byte{27, 91, 59, 51, 53, 109},
		LvError: []byte{27, 91, 59, 51, 49, 109},
		LvWarn:  []byte{27, 91, 59, 51, 51, 109},
	}
	ovcolor = []byte{27, 91, 48, 109}
)

func printColorText(out *os.File, lv int, content []byte) {
	if color, ok := lvcolor[lv]; ok {
		out.Write(color)
		out.Write(content)
		out.Write(ovcolor)
	} else {
		out.Write(content)
	}
}
