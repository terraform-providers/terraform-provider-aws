package hashcode

import (
	"bytes"
	"fmt"
	"hash/crc32"
)

// String hashes a string to a unique hashcode.
//
// Deprecated: This will be removed in v2 without replacement. If you need
// its functionality, you can copy it, import crc32 directly, or reference the
// v1 package.
//
// crc32 returns a uint32, but for our use we need
// and non negative integer. Here we cast to an integer
// and invert it if the result is negative.
func String(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}

// Strings hashes a list of strings to a unique hashcode.
//
// Deprecated: This will be removed in v2 without replacement. If you need
// its functionality, you can copy it, import crc32 directly, or reference the
// v1 package.
func Strings(strings []string) string {
	var buf bytes.Buffer

	for _, s := range strings {
		buf.WriteString(fmt.Sprintf("%s-", s))
	}

	return fmt.Sprintf("%d", String(buf.String()))
}
