// The method in the file has no effect
// Only for compatibility with non-Windows systems

// +build !windows

package color

func winSet(_ ...Color) (n int, err error) {
	return
}

func winReset() (n int, err error) {
	return
}

func winPrint(_ string, _ ...Color)   {}
func winPrintln(_ string, _ ...Color) {}
