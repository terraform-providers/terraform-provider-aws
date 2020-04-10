package provisioners

// Factory is a function type that creates a new instance of a resource
// provisioner, or returns an error if that is impossible.
type Factory func() (Interface, error)
