package format

import (
	"github.com/katbyte/terrafmt/lib/fmtverbs"
)

func FmtVerbBlock(content, path string) (string, error) {
	content = fmtverbs.Escape(content)

	fb, err := Block(content, path)
	if err != nil {
		return fb, err
	}

	fb = fmtverbs.Unscape(fb)

	return fb, nil
}
