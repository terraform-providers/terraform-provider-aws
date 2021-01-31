package tfjson

// Pos represents a position in a config file
type Pos struct {
	Line   int `json:"line"`
	Column int `json:"column"`
	Byte   int `json:"byte"`
}

// Range represents a range of bytes between two positions
type Range struct {
	Filename string `json:"filename"`
	Start    Pos    `json:"start"`
	End      Pos    `json:"end"`
}

// Diagnostic represents information to be presented to a user about an
// error or anomaly in parsing or evaluating configuration
type Diagnostic struct {
	Severity string `json:"severity,omitempty"`
	Summary  string `json:"summary,omitempty"`
	Detail   string `json:"detail,omitempty"`
	Range    *Range `json:"range,omitempty"`
}

// ValidateOutput represents JSON output from terraform validate
// (available from 0.12 onwards)
type ValidateOutput struct {
	Valid        bool         `json:"valid"`
	ErrorCount   int          `json:"error_count"`
	WarningCount int          `json:"warning_count"`
	Diagnostics  []Diagnostic `json:"diagnostics"`
}
