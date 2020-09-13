package tfdiags

// nativeError is a Diagnostic implementation that wraps a normal Go error
type nativeError struct {
	err error
}

var _ Diagnostic = nativeError{}

func (e nativeError) Severity() Severity {
	return Error
}

func (e nativeError) Description() Description {
	return Description{
		Summary: FormatError(e.err),
	}
}

func FromError(err error) Diagnostic {
	return &nativeError{
		err: err,
	}
}
