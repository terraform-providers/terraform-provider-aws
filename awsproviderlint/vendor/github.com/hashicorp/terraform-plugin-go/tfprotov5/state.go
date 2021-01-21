package tfprotov5

// RawState is the raw, undecoded state for providers to upgrade. It is
// undecoded as Terraform, for whatever reason, doesn't have the previous
// schema available to it, and so cannot decode the state itself and pushes
// that responsibility off onto providers.
//
// It is safe to assume that Flatmap can be ignored for any state written by
// Terraform 0.12.0 or higher, but it is not safe to assume that all states
// written by 0.12.0 or higher will be in JSON format; future versions may
// switch to an alternate encoding for states.
type RawState struct {
	JSON    []byte
	Flatmap map[string]string
}
