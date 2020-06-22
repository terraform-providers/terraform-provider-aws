package acctest

import (
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/plugin"
	tftest "github.com/hashicorp/terraform-plugin-test"
)

var TestHelper *tftest.Helper

func UseBinaryDriver(name string, providerFunc plugin.ProviderFunc) {
	sourceDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	if tftest.RunningAsPlugin() {
		plugin.Serve(&plugin.ServeOpts{
			ProviderFunc: providerFunc,
		})
		os.Exit(0)
	} else {
		TestHelper = tftest.AutoInitProviderHelper(name, sourceDir)
	}
}
