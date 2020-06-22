package main

import (
	"os"

	c "github.com/gookit/color"
	"github.com/katbyte/terrafmt/cli"
	"github.com/katbyte/terrafmt/lib/common"
)

func main() {
	if err := cli.Make().Execute(); err != nil {
		common.Log.Errorf(c.Sprintf("<red>terrafmt:</> %v", err))
		os.Exit(1)
	}

	os.Exit(0)
}
