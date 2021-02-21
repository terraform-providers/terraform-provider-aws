// +build !go1.13

package tfexec

import (
	"os/exec"
	"strings"
)

// cmdString handles go 1.12 as stringer was only added to exec.Cmd in 1.13
func cmdString(c *exec.Cmd) string {
	b := new(strings.Builder)
	b.WriteString(c.Path)
	for _, a := range c.Args[1:] {
		b.WriteByte(' ')
		b.WriteString(a)
	}
	return b.String()
}
