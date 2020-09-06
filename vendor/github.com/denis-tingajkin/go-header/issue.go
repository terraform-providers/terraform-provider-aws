package goheader

type Issue interface {
	Location() Location
	Message() string
}

type issue struct {
	msg      string
	location Location
}

func (i *issue) Location() Location {
	return i.location
}

func (i *issue) Message() string {
	return i.msg
}

func NewIssueWithLocation(msg string, location Location) Issue {
	return &issue{
		msg:      msg,
		location: location,
	}
}

func NewIssue(msg string) Issue {
	return &issue{
		msg: msg,
	}
}
