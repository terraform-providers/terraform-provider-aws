package tfdiags

type Diagnostic interface {
	Severity() Severity
	Description() Description
}

type Severity rune

//go:generate go run golang.org/x/tools/cmd/stringer -type=Severity

const (
	Error   Severity = 'E'
	Warning Severity = 'W'
)

type Description struct {
	Summary string
	Detail  string
}
