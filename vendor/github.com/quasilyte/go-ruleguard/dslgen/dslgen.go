// +build generate

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func main() {
	// See #23.

	data, err := dirToBytes("../dsl/fluent")
	if err != nil {
		panic(err)
	}

	f, err := os.Create("./dsl_sources.go")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fmt.Fprintf(f, `package dslgen

var Fluent = []byte(%q)
`, string(data))
}

func dirToBytes(dir string) ([]byte, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	for i, f := range files {
		data, err := ioutil.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			return nil, err
		}
		if i != 0 {
			newline := bytes.IndexByte(data, '\n')
			data = data[newline:]
		}
		buf.Write(data)
		buf.WriteByte('\n')
	}
	return buf.Bytes(), nil
}
