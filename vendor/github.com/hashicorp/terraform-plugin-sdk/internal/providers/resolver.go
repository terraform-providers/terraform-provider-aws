package providers

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/internal/plugin/discovery"
)

// Resolver is an interface implemented by objects that are able to resolve
// a given set of resource provider version constraints into Factory
// callbacks.
type Resolver interface {
	// Given a constraint map, return a Factory for each requested provider.
	// If some or all of the constraints cannot be satisfied, return a non-nil
	// slice of errors describing the problems.
	ResolveProviders(reqd discovery.PluginRequirements) (map[string]Factory, []error)
}

// ResolverFunc wraps a callback function and turns it into a Resolver
// implementation, for convenience in situations where a function and its
// associated closure are sufficient as a resolver implementation.
type ResolverFunc func(reqd discovery.PluginRequirements) (map[string]Factory, []error)

// ResolveProviders implements Resolver by calling the
// wrapped function.
func (f ResolverFunc) ResolveProviders(reqd discovery.PluginRequirements) (map[string]Factory, []error) {
	return f(reqd)
}

// ResolverFixed returns a Resolver that has a fixed set of provider factories
// provided by the caller. The returned resolver ignores version constraints
// entirely and just returns the given factory for each requested provider
// name.
//
// This function is primarily used in tests, to provide mock providers or
// in-process providers under test.
func ResolverFixed(factories map[string]Factory) Resolver {
	return ResolverFunc(func(reqd discovery.PluginRequirements) (map[string]Factory, []error) {
		ret := make(map[string]Factory, len(reqd))
		var errs []error
		for name := range reqd {
			if factory, exists := factories[name]; exists {
				ret[name] = factory
			} else {
				errs = append(errs, fmt.Errorf("provider %q is not available", name))
			}
		}
		return ret, errs
	})
}

// Factory is a function type that creates a new instance of a resource
// provider, or returns an error if that is impossible.
type Factory func() (Interface, error)

// FactoryFixed is a helper that creates a Factory that just returns some given
// single provider.
//
// Unlike usual factories, the exact same instance is returned for each call
// to the factory and so this must be used in only specialized situations where
// the caller can take care to either not mutate the given provider at all
// or to mutate it in ways that will not cause unexpected behavior for others
// holding the same reference.
func FactoryFixed(p Interface) Factory {
	return func() (Interface, error) {
		return p, nil
	}
}
