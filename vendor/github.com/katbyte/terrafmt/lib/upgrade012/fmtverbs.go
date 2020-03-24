package upgrade012

import (
	"github.com/katbyte/terrafmt/lib/fmtverbs"
)

func Upgrade12VerbBlock(b string) (string, error) {
	b = fmtverbs.Escape(b)

	fb, err := Block(b)
	if err != nil {
		return fb, err
	}

	fb = fmtverbs.Unscape(fb)

	return fb, nil
}
