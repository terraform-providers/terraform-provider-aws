package fmtverbs

import (
	"regexp"
	"strings"
)

func Escape(b string) string {
	// NOTE: the order of these replacements matter

	// %s
	// figure out why the * doesn't match both later
	b = regexp.MustCompile(`(?m:^%(\.[0-9])?[sdfgtq]$)`).ReplaceAllString(b, `#@@_@@ TFMT:$0:TMFT @@_@@#`)
	b = regexp.MustCompile(`(?m:^[ \t]*%(\.[0-9])?[sdfgtq]$)`).ReplaceAllString(b, `#@@_@@ TFMT:$0:TMFT @@_@@#`)

	// %[n]s
	b = regexp.MustCompile(`(?m:^%(\.[0-9])?\[[\d]+\][sdfgtq]$)`).ReplaceAllString(b, `#@@_@@ TFMT:$0:TMFT @@_@@#`)
	b = regexp.MustCompile(`(?m:^[ \t]*%(\.[0-9])?\[[\d]+\][sdfgtq]$)`).ReplaceAllString(b, `#@@_@@ TFMT:$0:TMFT @@_@@#`)

	// = [%s]
	b = regexp.MustCompile(`(?m:\[%(\.[0-9])?[sdfgtq]\]$)`).ReplaceAllString(b, `["@@_@@ TFMT:$0:TFMT @@_@@"]`)

	// = [%[n]s]
	b = regexp.MustCompile(`(?m:\[%(\.[0-9])?\[[\d]+\][sdfgtq]\]$)`).ReplaceAllString(b, `["@@_@@ TFMT:$0:TFMT @@_@@"]`)

	// = %s
	b = regexp.MustCompile(`(?m:%(\.[0-9])?[sdfgtq]$)`).ReplaceAllString(b, `"@@_@@ TFMT:$0:TFMT @@_@@"`)

	// = %[n]s
	b = regexp.MustCompile(`(?m:%(\.[0-9])?\[[\d]+\][sdfgtq]$)`).ReplaceAllString(b, `"@@_@@ TFMT:$0:TFMT @@_@@"`)

	// base64encode(%s) or md5(%s)
	b = regexp.MustCompile(`\(%`).ReplaceAllString(b, `(TFFMTKTBRACKETPERCENT`)

	//  .12 - something.%s.prop
	b = regexp.MustCompile(`\.%s`).ReplaceAllString(b, `.TFMTKTKTTFMTs`)
	b = regexp.MustCompile(`\.%d`).ReplaceAllString(b, `.TFMTKTKTTFMTd`)
	b = regexp.MustCompile(`\.%f`).ReplaceAllString(b, `.TFMTKTKTTFMTf`)
	b = regexp.MustCompile(`\.%g`).ReplaceAllString(b, `.TFMTKTKTTFMTg`)
	b = regexp.MustCompile(`\.%t`).ReplaceAllString(b, `.TFMTKTKTTFMTt`)
	b = regexp.MustCompile(`\.%q`).ReplaceAllString(b, `.TFMTKTKTTFMTq`)

	return b
}

func Unscape(fb string) string {
	// NOTE: the order of these replacements matter

	//undo replace
	fb = regexp.MustCompile(`[ ]*#@@_@@ TFMT:`).ReplaceAllString(fb, ``)
	fb = strings.ReplaceAll(fb, "#@@_@@ TFMT:", "")
	fb = strings.ReplaceAll(fb, ":TMFT @@_@@#", "")

	fb = strings.ReplaceAll(fb, "[\"@@_@@ TFMT:", "")
	fb = strings.ReplaceAll(fb, ":TFMT @@_@@\"]", "")

	// order here matters, replace the ones with [], then do the ones without
	fb = strings.ReplaceAll(fb, "\"@@_@@ TFMT:", "")
	fb = strings.ReplaceAll(fb, ":TFMT @@_@@\"", "")

	// .12
	fb = strings.ReplaceAll(fb, ".TFMTKTKTTFMT", ".%")

	// function(%
	fb = strings.ReplaceAll(fb, "TFFMTKTBRACKETPERCENT", "%")

	return fb
}
