package tftest

import (
	"os"
)

// RunningAsPlugin returns true if it detects the usual Terraform plugin
// detection environment variables, suggesting that the current process is
// being launched as a plugin server.
func RunningAsPlugin() bool {
	const cookieVar = "TF_PLUGIN_MAGIC_COOKIE"
	const cookieVal = "d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2"

	return os.Getenv(cookieVar) == cookieVal
}
