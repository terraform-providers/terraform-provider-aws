package tfdiags

// diagnosticBase can be embedded in other diagnostic structs to get
// default implementations of Severity and Description. This type also
// has default implementations of Source that return no source
// location or expression-related information, so embedders should generally
// override those method to return more useful results where possible.
type diagnosticBase struct {
	severity Severity
	summary  string
	detail   string
}

func (d diagnosticBase) Severity() Severity {
	return d.severity
}

func (d diagnosticBase) Description() Description {
	return Description{
		Summary: d.summary,
		Detail:  d.detail,
	}
}

func Diag(sev Severity, summary, detail string) Diagnostic {
	return &diagnosticBase{
		severity: sev,
		summary:  summary,
		detail:   detail,
	}
}
