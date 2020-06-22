/*
Package color is Command line color library.
Support rich color rendering output, universal API method, compatible with Windows system

Source code and other details for the project are available at GitHub:

	https://github.com/gookit/color

More usage please see README and tests.
*/
package color

import (
	"fmt"
	"regexp"
)

// console color mode
// const (
// 	ModeNormal = iota
// 	Mode256    // 8 bite
// 	ModeRGB    // 24 bite
// 	ModeGrayscale
// )

// color render templates
// ESC 操作的表示:
// 	"\033"(Octal 8进制) = "\x1b"(Hexadecimal 16进制) = 27 (10进制)
const (
	SettingTpl   = "\x1b[%sm"
	FullColorTpl = "\x1b[%sm%s\x1b[0m"
)

// ResetSet 重置/正常 关闭所有属性。
const ResetSet = "\x1b[0m"

// CodeExpr regex to clear color codes eg "\033[1;36mText\x1b[0m"
const CodeExpr = `\033\[[\d;?]+m`

// Enable switch color display
var Enable = true

var (
	// mark current env, It's like in `cmd.exe`
	// if not in windows, is's always is False.
	isLikeInCmd bool
	// match color codes
	codeRegex = regexp.MustCompile(CodeExpr)
	// mark current env is support color.
	// Always: isLikeInCmd != isSupportColor
	isSupportColor = IsSupportColor()
)

/*************************************************************
 * global settings
 *************************************************************/

// Set set console color attributes
func Set(colors ...Color) (int, error) {
	if !Enable { // not enable
		return 0, nil
	}

	// on windows cmd.exe
	if isLikeInCmd {
		return winSet(colors...)
	}

	return fmt.Printf(SettingTpl, colors2code(colors...))
}

// Reset reset console color attributes
func Reset() (int, error) {
	if !Enable { // not enable
		return 0, nil
	}

	// on windows cmd.exe
	if isLikeInCmd {
		return winReset()
	}

	return fmt.Print(ResetSet)
}

// Disable disable color output
func Disable() {
	Enable = false
}

/*************************************************************
 * render color code
 *************************************************************/

// RenderCode render message by color code.
// Usage:
// 	msg := RenderCode("3;32;45", "some", "message")
func RenderCode(code string, args ...interface{}) string {
	message := fmt.Sprint(args...)
	if len(code) == 0 {
		return message
	}

	// disabled OR not support color
	if !Enable || !isSupportColor {
		return ClearCode(message)
	}

	return fmt.Sprintf(FullColorTpl, code, message)
}

// RenderString render a string with color code.
// Usage:
// 	msg := RenderString("3;32;45", "a message")
func RenderString(code string, str string) string {
	// some check
	if len(code) == 0 || str == "" {
		return str
	}

	// disabled OR not support color
	if !Enable || !isSupportColor {
		return ClearCode(str)
	}

	return fmt.Sprintf(FullColorTpl, code, str)
}

// ClearCode clear color codes.
// eg:
// 		"\033[36;1mText\x1b[0m" -> "Text"
func ClearCode(str string) string {
	return codeRegex.ReplaceAllString(str, "")
}

/*************************************************************
 * colored message Printer
 *************************************************************/

// PrinterFace interface
type PrinterFace interface {
	fmt.Stringer
	Sprint(a ...interface{}) string
	Sprintf(format string, a ...interface{}) string
	Print(a ...interface{})
	Printf(format string, a ...interface{})
	Println(a ...interface{})
}

// Printer a generic color message printer.
// Usage:
// 	p := &Printer{"32;45;3"}
// 	p.Print("message")
type Printer struct {
	// ColorCode color code string. eg "32;45;3"
	ColorCode string
}

// NewPrinter instance
func NewPrinter(colorCode string) *Printer {
	return &Printer{colorCode}
}

// String returns color code string. eg: "32;45;3"
func (p *Printer) String() string {
	// panic("implement me")
	return p.ColorCode
}

// Sprint returns rendering colored messages
func (p *Printer) Sprint(a ...interface{}) string {
	return RenderCode(p.String(), a...)
}

// Sprintf returns format and rendering colored messages
func (p *Printer) Sprintf(format string, a ...interface{}) string {
	return RenderString(p.String(), fmt.Sprintf(format, a...))
}

// Print rendering colored messages
func (p *Printer) Print(a ...interface{}) {
	fmt.Print(RenderCode(p.String(), a...))
}

// Printf format and rendering colored messages
func (p *Printer) Printf(format string, a ...interface{}) {
	fmt.Print(RenderString(p.String(), fmt.Sprintf(format, a...)))
}

// Println rendering colored messages with newline
func (p *Printer) Println(a ...interface{}) {
	fmt.Println(RenderCode(p.String(), a...))
}

// IsEmpty color code
func (p *Printer) IsEmpty() bool {
	return p.ColorCode == ""
}
