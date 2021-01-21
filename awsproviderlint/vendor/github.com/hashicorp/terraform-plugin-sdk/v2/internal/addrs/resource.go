package addrs

import (
	"fmt"
)

// resource is an address for a resource block within configuration, which
// contains potentially-multiple resource instances if that configuration
// block uses "count" or "for_each".
type resource struct {
	Mode resourceMode
	Type string
	Name string
}

func (r resource) String() string {
	switch r.Mode {
	case ManagedResourceMode:
		return fmt.Sprintf("%s.%s", r.Type, r.Name)
	case DataResourceMode:
		return fmt.Sprintf("data.%s.%s", r.Type, r.Name)
	default:
		// Should never happen, but we'll return a string here rather than
		// crashing just in case it does.
		return fmt.Sprintf("<invalid>.%s.%s", r.Type, r.Name)
	}
}

// resourceInstance is an address for a specific instance of a resource.
// When a resource is defined in configuration with "count" or "for_each" it
// produces zero or more instances, which can be addressed using this type.
type resourceInstance struct {
	Resource resource
	Key      instanceKey
}

func (r resourceInstance) ContainingResource() resource {
	return r.Resource
}

func (r resourceInstance) String() string {
	if r.Key == NoKey {
		return r.Resource.String()
	}
	return r.Resource.String() + r.Key.String()
}

// absResource is an absolute address for a resource under a given module path.
type absResource struct {
	Module   ModuleInstance
	Resource resource
}

// Resource returns the address of a particular resource within the receiver.
func (m ModuleInstance) Resource(mode resourceMode, typeName string, name string) absResource {
	return absResource{
		Module: m,
		Resource: resource{
			Mode: mode,
			Type: typeName,
			Name: name,
		},
	}
}

func (r absResource) String() string {
	if len(r.Module) == 0 {
		return r.Resource.String()
	}
	return fmt.Sprintf("%s.%s", r.Module.String(), r.Resource.String())
}

// absResourceInstance is an absolute address for a resource instance under a
// given module path.
type absResourceInstance struct {
	Module   ModuleInstance
	Resource resourceInstance
}

// ResourceInstance returns the address of a particular resource instance within the receiver.
func (m ModuleInstance) ResourceInstance(mode resourceMode, typeName string, name string, key instanceKey) absResourceInstance {
	return absResourceInstance{
		Module: m,
		Resource: resourceInstance{
			Resource: resource{
				Mode: mode,
				Type: typeName,
				Name: name,
			},
			Key: key,
		},
	}
}

// ContainingResource returns the address of the resource that contains the
// receving resource instance. In other words, it discards the key portion
// of the address to produce an absResource value.
func (r absResourceInstance) ContainingResource() absResource {
	return absResource{
		Module:   r.Module,
		Resource: r.Resource.ContainingResource(),
	}
}

func (r absResourceInstance) String() string {
	if len(r.Module) == 0 {
		return r.Resource.String()
	}
	return fmt.Sprintf("%s.%s", r.Module.String(), r.Resource.String())
}

// resourceMode defines which lifecycle applies to a given resource. Each
// resource lifecycle has a slightly different address format.
type resourceMode rune

//go:generate go run golang.org/x/tools/cmd/stringer -type resourceMode

const (
	// InvalidResourceMode is the zero value of ResourceMode and is not
	// a valid resource mode.
	InvalidResourceMode resourceMode = 0

	// ManagedResourceMode indicates a managed resource, as defined by
	// "resource" blocks in configuration.
	ManagedResourceMode resourceMode = 'M'

	// DataResourceMode indicates a data resource, as defined by
	// "data" blocks in configuration.
	DataResourceMode resourceMode = 'D'
)
