package tfprotov5

const (
	// StringKindPlain indicates a string is plaintext, and should be
	// interpreted as having no formatting information.
	StringKindPlain StringKind = 0

	// StringKindMarkdown indicates a string is markdown-formatted, and
	// should be rendered using a Markdown renderer to correctly display
	// its formatting.
	StringKindMarkdown StringKind = 1
)

// StringKind indicates a formatting or encoding scheme for a string.
type StringKind int32

func (s StringKind) String() string {
	switch s {
	case 0:
		return "PLAIN"
	case 1:
		return "MARKDOWN"
	}
	return "UNKNOWN"
}
