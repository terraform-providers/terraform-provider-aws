// +build windows

// Display color on windows
// refer:
//  golang.org/x/sys/windows
// 	golang.org/x/crypto/ssh/terminal
// 	https://docs.microsoft.com/en-us/windows/console
package color

import (
	"fmt"
	"syscall"
	"unsafe"
)

// color on windows cmd
// you can see on windows by command: COLOR /?
// windows color build by: "Bg + Fg" OR only "Fg"
// Consists of any two of the following:
// the first is the background color, and the second is the foreground color
// 颜色属性由两个十六进制数字指定
//  - 第一个对应于背景，第二个对应于前景。
// 	- 当只传入一个值时，则认为是前景色
// 每个数字可以为以下任何值:
// more see: https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/cmd
const (
	// Foreground colors.
	winFgBlack  uint16 = 0x00 // 0 黑色
	winFgBlue   uint16 = 0x01 // 1 蓝色
	winFgGreen  uint16 = 0x02 // 2 绿色
	winFgAqua   uint16 = 0x03 // 3 浅绿 skyblue
	winFgRed    uint16 = 0x04 // 4 红色
	winFgPink   uint16 = 0x05 // 5 紫色/品红
	winFgYellow uint16 = 0x06 // 6 黄色
	winFgWhite  uint16 = 0x07 // 7 白色
	winFgGray   uint16 = 0x08 // 8 灰色

	winFgLightBlue   uint16 = 0x09 // 9 淡蓝色
	winFgLightGreen  uint16 = 0x0a // 10 淡绿色
	winFgLightAqua   uint16 = 0x0b // 11 淡浅绿色
	winFgLightRed    uint16 = 0x0c // 12 淡红色
	winFgLightPink   uint16 = 0x0d // 13 Purple 淡紫色, Pink 粉红
	winFgLightYellow uint16 = 0x0e // 14 淡黄色
	winFgLightWhite  uint16 = 0x0f // 15 亮白色

	// Background colors.
	winBgBlack  uint16 = 0x00 // 黑色
	winBgBlue   uint16 = 0x10 // 蓝色
	winBgGreen  uint16 = 0x20 // 绿色
	winBgAqua   uint16 = 0x30 // 浅绿 skyblue
	winBgRed    uint16 = 0x40 // 红色
	winBgPink   uint16 = 0x50 // 紫色
	winBgYellow uint16 = 0x60 // 黄色
	winBgWhite  uint16 = 0x70 // 白色
	winBgGray   uint16 = 0x80 // 128 灰色

	winBgLightBlue   uint16 = 0x90 // 淡蓝色
	winBgLightGreen  uint16 = 0xa0 // 淡绿色
	winBgLightAqua   uint16 = 0xb0 // 淡浅绿色
	winBgLightRed    uint16 = 0xc0 // 淡红色
	winBgLightPink   uint16 = 0xd0 // 淡紫色
	winBgLightYellow uint16 = 0xe0 // 淡黄色
	winBgLightWhite  uint16 = 0xf0 // 240 亮白色

	// bg black, fg white
	winDefSetting = winBgBlack | winFgWhite

	// Option settings
	// see https://docs.microsoft.com/en-us/windows/console/char-info-str
	winFgIntensity uint16 = 0x0008 // 8 前景强度
	winBgIntensity uint16 = 0x0080 // 128 背景强度

	WinOpLeading    uint16 = 0x0100 // 前导字节
	WinOpTrailing   uint16 = 0x0200 // 尾随字节
	WinOpHorizontal uint16 = 0x0400 // 顶部水平
	WinOpReverse    uint16 = 0x4000 // 反转前景和背景
	WinOpUnderscore uint16 = 0x8000 // 32768 下划线
)

// color on windows
var winColorsMap map[Color]uint16

var (
	// for cmd.exe
	// echo %ESC%[1;33;40m Yellow on black %ESC%[0m
	escChar = ""
	// isMSys bool
	kernel32 *syscall.LazyDLL

	procGetConsoleMode *syscall.LazyProc
	// procSetConsoleMode *syscall.LazyProc

	procSetTextAttribute           *syscall.LazyProc
	procGetConsoleScreenBufferInfo *syscall.LazyProc

	// console screen buffer info
	// eg {size:{x:215 y:3000} cursorPosition:{x:0 y:893} attributes:7 window:{left:0 top:882 right:214 bottom:893} maximumWindowSize:{x:215 y:170}}
	defScreenInfo consoleScreenBufferInfo
)

func init() {
	// if at linux, mac, or windows's ConEmu, Cmder, putty
	if isSupportColor {
		return
	}

	// init some info
	isLikeInCmd = true
	initWinColorsMap()

	// isMSys = utils.IsMSys()
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	// https://docs.microsoft.com/en-us/windows/console/setconsolemode
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	// procSetConsoleMode = kernel32.NewProc("SetConsoleMode")

	procSetTextAttribute = kernel32.NewProc("SetConsoleTextAttribute")
	// https://docs.microsoft.com/en-us/windows/console/getconsolescreenbufferinfo
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")

	// fetch console screen buffer info
	// err := getConsoleScreenBufferInfo(uintptr(syscall.Stdout), &defScreenInfo)
}

// initWinColorsMap init colors to win-colors mapping
func initWinColorsMap() {
	// init map
	winColorsMap = map[Color]uint16{
		// Foreground colors
		FgBlack:   winFgBlack,
		FgRed:     winFgRed,
		FgGreen:   winFgGreen,
		FgYellow:  winFgYellow,
		FgBlue:    winFgBlue,
		FgMagenta: winFgPink, // diff
		FgCyan:    winFgAqua, // diff
		FgWhite:   winFgWhite,
		FgDefault: winFgWhite,

		// Extra Foreground colors
		FgDarkGray:     winFgGray,
		FgLightRed:     winFgLightBlue,
		FgLightGreen:   winFgLightGreen,
		FgLightYellow:  winFgLightYellow,
		FgLightBlue:    winFgLightRed,
		FgLightMagenta: winFgLightPink,
		FgLightCyan:    winFgLightAqua,
		FgLightWhite:   winFgLightWhite,

		// Background colors
		BgBlack:   winBgBlack,
		BgRed:     winBgRed,
		BgGreen:   winBgGreen,
		BgYellow:  winBgYellow,
		BgBlue:    winBgBlue,
		BgMagenta: winBgPink, // diff
		BgCyan:    winBgAqua, // diff
		BgWhite:   winBgWhite,
		BgDefault: winBgBlack,

		// Extra Background colors
		BgDarkGray:     winBgGray,
		BgLightRed:     winBgLightBlue,
		BgLightGreen:   winBgLightGreen,
		BgLightYellow:  winBgLightYellow,
		BgLightBlue:    winBgLightRed,
		BgLightMagenta: winBgLightPink,
		BgLightCyan:    winBgLightAqua,
		BgLightWhite:   winBgLightWhite,

		// Option settings(注释掉的，将在win cmd中忽略掉)
		// OpReset: winDefSetting,  // 重置所有设置
		OpBold: winFgIntensity, // 加粗 ->
		// OpFuzzy:                    // 模糊(不是所有的终端仿真器都支持)
		// OpItalic                    // 斜体(不是所有的终端仿真器都支持)
		OpUnderscore: WinOpUnderscore, // 下划线
		// OpBlink                      // 闪烁
		// OpFastBlink                  // 快速闪烁(未广泛支持)
		// OpReverse: WinOpReverse      // 颠倒的 交换背景色与前景色
		// OpConcealed                  // 隐匿的
		// OpStrikethrough              // 删除的，删除线(未广泛支持)
	}
}

// winPrint
func winPrint(str string, colors ...Color) {
	_, _ = winInternalPrint(str, convertColorsToWinAttr(colors), false)
}

// winPrintln
func winPrintln(str string, colors ...Color) {
	_, _ = winInternalPrint(str, convertColorsToWinAttr(colors), true)
}

// winInternalPrint
// winInternalPrint("hello [OK];", 2|8, true) //亮绿色
func winInternalPrint(str string, attribute uint16, newline bool) (int, error) {
	if !Enable { // not enable
		if newline {
			return fmt.Println(str)
		}

		return fmt.Print(str)
	}

	// fmt.Print("attribute val: ", attribute, "\n")
	_, _ = setConsoleTextAttr(uintptr(syscall.Stdout), attribute)

	if newline {
		fmt.Println(str)
	} else {
		fmt.Print(str)
	}

	// handle, _, _ = procSetTextAttribute.Call(uintptr(syscall.Stdout), winDefSetting)
	// closeHandle := kernel32.NewProc("CloseHandle")
	// closeHandle.Call(handle)

	return winReset()
}

// func winRender(str string, colors ...Color) string {
// 	setConsoleTextAttr(uintptr(syscall.Stdout), convertColorsToWinAttr(colors))
//
// 	return str
// }

// winSet set console color attributes
func winSet(colors ...Color) (int, error) {
	if !Enable { // not enable
		return 0, nil
	}

	return setConsoleTextAttr(uintptr(syscall.Stdout), convertColorsToWinAttr(colors))
}

// winReset reset color settings to default
func winReset() (int, error) {
	// not enable
	if !Enable {
		return 0, nil
	}

	return setConsoleTextAttr(uintptr(syscall.Stdout), winDefSetting)
}

// convertColorsToWinAttr convert generic colors to win-colors attribute
func convertColorsToWinAttr(colors []Color) uint16 {
	var setting uint16
	for _, c := range colors {
		// check exists
		if wc, ok := winColorsMap[c]; ok {
			setting |= wc
		}
	}

	return setting
}

// getWinColor convert Color to win-color value
func getWinColor(color Color) uint16 {
	if wc, ok := winColorsMap[color]; ok {
		return wc
	}

	return 0
}

// setConsoleTextAttr
// ret != 0 is OK.
func setConsoleTextAttr(consoleOutput uintptr, winAttr uint16) (n int, err error) {
	// err is type of syscall.Errno
	ret, _, err := procSetTextAttribute.Call(consoleOutput, uintptr(winAttr))

	// if success, err.Error() is equals "The operation completed successfully."
	if err != nil && err.Error() == "The operation completed successfully." {
		err = nil // set as nil
	}

	return int(ret), err
}

// IsTty returns true if the given file descriptor is a terminal.
func IsTty(fd uintptr) bool {
	var st uint32
	r, _, e := syscall.Syscall(procGetConsoleMode.Addr(), 2, fd, uintptr(unsafe.Pointer(&st)), 0)
	return r != 0 && e == 0
}

// IsTerminal returns true if the given file descriptor is a terminal.
// Usage:
// 	fd := os.Stdout.Fd()
// 	fd := uintptr(syscall.Stdout) // for windows
// 	IsTerminal(fd)
func IsTerminal(fd int) bool {
	var st uint32
	r, _, e := syscall.Syscall(procGetConsoleMode.Addr(), 2, uintptr(fd), uintptr(unsafe.Pointer(&st)), 0)
	return r != 0 && e == 0
}

// from package: golang.org/x/sys/windows
type (
	short int16
	word uint16

	// coord cursor position coordinates
	coord struct {
		x short
		y short
	}

	smallRect struct {
		left   short
		top    short
		right  short
		bottom short
	}

	// Used with GetConsoleScreenBuffer to retrieve information about a console
	// screen buffer. See
	// https://docs.microsoft.com/en-us/windows/console/console-screen-buffer-info-str
	// for details.
	consoleScreenBufferInfo struct {
		size              coord
		cursorPosition    coord
		attributes        word // is windows color setting
		window            smallRect
		maximumWindowSize coord
	}
)

// GetSize returns the dimensions of the given terminal.
func getSize(fd int) (width, height int, err error) {
	var info consoleScreenBufferInfo
	if err := getConsoleScreenBufferInfo(uintptr(fd), &info); err != nil {
		return 0, 0, err
	}

	return int(info.size.x), int(info.size.y), nil
}

// from package: golang.org/x/sys/windows
func getConsoleScreenBufferInfo(consoleOutput uintptr, info *consoleScreenBufferInfo) (err error) {
	r1, _, e1 := syscall.Syscall(procGetConsoleScreenBufferInfo.Addr(), 2, consoleOutput, uintptr(unsafe.Pointer(info)), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = e1
		} else {
			err = syscall.EINVAL
		}
	}

	return
}
