package terraform

//go:generate go run golang.org/x/tools/cmd/stringer -type=instanceType instancetype.go

// instanceType is an enum of the various types of instances store in the State
type instanceType int

const (
	typeInvalid instanceType = iota
	typePrimary
	typeTainted
	typeDeposed
)
