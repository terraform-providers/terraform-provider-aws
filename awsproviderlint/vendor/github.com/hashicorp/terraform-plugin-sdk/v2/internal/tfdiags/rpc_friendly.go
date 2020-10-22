package tfdiags

import (
	"encoding/gob"
)

type rpcFriendlyDiag struct {
	Severity_ Severity
	Summary_  string
	Detail_   string
}

// rpcFriendlyDiag transforms a given diagnostic so that is more friendly to
// RPC.
//
// In particular, it currently returns an object that can be serialized and
// later re-inflated using gob. This definition may grow to include other
// serializations later.
func makeRPCFriendlyDiag(diag Diagnostic) Diagnostic {
	desc := diag.Description()
	return &rpcFriendlyDiag{
		Severity_: diag.Severity(),
		Summary_:  desc.Summary,
		Detail_:   desc.Detail,
	}
}

func (d *rpcFriendlyDiag) Severity() Severity {
	return d.Severity_
}

func (d *rpcFriendlyDiag) Description() Description {
	return Description{
		Summary: d.Summary_,
		Detail:  d.Detail_,
	}
}

func init() {
	gob.Register((*rpcFriendlyDiag)(nil))
}
