package goheader

import "fmt"

type Location struct {
	Line     int
	Position int
}

func (l Location) String() string {
	return fmt.Sprintf("%v:%v", l.Line+1, l.Position)
}
