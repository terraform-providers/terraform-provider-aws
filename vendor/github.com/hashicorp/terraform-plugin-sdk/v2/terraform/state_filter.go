package terraform

import (
	"fmt"
	"sort"
)

// stateFilter is responsible for filtering and searching a state.
//
// This is a separate struct from State rather than a method on State
// because StateFilter might create sidecar data structures to optimize
// filtering on the state.
//
// If you change the State, the filter created is invalid and either
// Reset should be called or a new one should be allocated. StateFilter
// will not watch State for changes and do this for you. If you filter after
// changing the State without calling Reset, the behavior is not defined.
type stateFilter struct {
	State *State
}

// Filter takes the addresses specified by fs and finds all the matches.
// The values of fs are resource addressing syntax that can be parsed by
// parseResourceAddress.
func (f *stateFilter) filter(fs ...string) ([]*stateFilterResult, error) {
	// Parse all the addresses
	as := make([]*resourceAddress, len(fs))
	for i, v := range fs {
		a, err := parseResourceAddress(v)
		if err != nil {
			return nil, fmt.Errorf("Error parsing address '%s': %s", v, err)
		}

		as[i] = a
	}

	// If we weren't given any filters, then we list all
	if len(fs) == 0 {
		as = append(as, &resourceAddress{Index: -1})
	}

	// Filter each of the address. We keep track of this in a map to
	// strip duplicates.
	resultSet := make(map[string]*stateFilterResult)
	for _, a := range as {
		for _, r := range f.filterSingle(a) {
			resultSet[r.String()] = r
		}
	}

	// Make the result list
	results := make([]*stateFilterResult, 0, len(resultSet))
	for _, v := range resultSet {
		results = append(results, v)
	}

	// Sort them and return
	sort.Sort(stateFilterResultSlice(results))
	return results, nil
}

func (f *stateFilter) filterSingle(a *resourceAddress) []*stateFilterResult {
	// The slice to keep track of results
	var results []*stateFilterResult

	// Go through modules first.
	modules := make([]*ModuleState, 0, len(f.State.Modules))
	for _, m := range f.State.Modules {
		if f.relevant(a, m) {
			modules = append(modules, m)

			// Only add the module to the results if we haven't specified a type.
			// We also ignore the root module.
			if a.Type == "" && len(m.Path) > 1 {
				results = append(results, &stateFilterResult{
					Path:    m.Path[1:],
					Address: (&resourceAddress{Path: m.Path[1:]}).String(),
					Value:   m,
				})
			}
		}
	}

	// With the modules set, go through all the resources within
	// the modules to find relevant resources.
	for _, m := range modules {
		for n, r := range m.Resources {
			// The name in the state contains valuable information. Parse.
			key, err := parseResourceStateKey(n)
			if err != nil {
				// If we get an error parsing, then just ignore it
				// out of the state.
				continue
			}

			// Older states and test fixtures often don't contain the
			// type directly on the ResourceState. We add this so StateFilter
			// is a bit more robust.
			if r.Type == "" {
				r.Type = key.Type
			}

			if f.relevant(a, r) {
				if a.Name != "" && a.Name != key.Name {
					// Name doesn't match
					continue
				}

				if a.Index >= 0 && key.Index != a.Index {
					// Index doesn't match
					continue
				}

				if a.Name != "" && a.Name != key.Name {
					continue
				}

				// Build the address for this resource
				addr := &resourceAddress{
					Path:  m.Path[1:],
					Name:  key.Name,
					Type:  key.Type,
					Index: key.Index,
				}

				// Add the resource level result
				resourceResult := &stateFilterResult{
					Path:    addr.Path,
					Address: addr.String(),
					Value:   r,
				}
				if !a.InstanceTypeSet {
					results = append(results, resourceResult)
				}

				// Add the instances
				if r.Primary != nil {
					addr.InstanceType = typePrimary
					addr.InstanceTypeSet = false
					results = append(results, &stateFilterResult{
						Path:    addr.Path,
						Address: addr.String(),
						Parent:  resourceResult,
						Value:   r.Primary,
					})
				}

				for _, instance := range r.Deposed {
					if f.relevant(a, instance) {
						addr.InstanceType = typeDeposed
						addr.InstanceTypeSet = true
						results = append(results, &stateFilterResult{
							Path:    addr.Path,
							Address: addr.String(),
							Parent:  resourceResult,
							Value:   instance,
						})
					}
				}
			}
		}
	}

	return results
}

// relevant checks for relevance of this address against the given value.
func (f *stateFilter) relevant(addr *resourceAddress, raw interface{}) bool {
	switch v := raw.(type) {
	case *ModuleState:
		path := v.Path[1:]

		if len(addr.Path) > len(path) {
			// Longer path in address means there is no way we match.
			return false
		}

		// Check for a prefix match
		for i, p := range addr.Path {
			if path[i] != p {
				// Any mismatches don't match.
				return false
			}
		}

		return true
	case *ResourceState:
		if addr.Type == "" {
			// If we have no resource type, then we're interested in all!
			return true
		}

		// If the type doesn't match we fail immediately
		if v.Type != addr.Type {
			return false
		}

		return true
	default:
		// If we don't know about it, let's just say no
		return false
	}
}

// stateFilterResult is a single result from a filter operation. Filter
// can match multiple things within a state (module, resource, instance, etc.)
// and this unifies that.
type stateFilterResult struct {
	// Module path of the result
	Path []string

	// Address is the address that can be used to reference this exact result.
	Address string

	// Parent, if non-nil, is a parent of this result. For instances, the
	// parent would be a resource. For resources, the parent would be
	// a module. For modules, this is currently nil.
	Parent *stateFilterResult

	// Value is the actual value. This must be type switched on. It can be
	// any data structures that `State` can hold: `ModuleState`,
	// `ResourceState`, `InstanceState`.
	Value interface{}
}

func (r *stateFilterResult) String() string {
	return fmt.Sprintf("%T: %s", r.Value, r.Address)
}

func (r *stateFilterResult) sortedType() int {
	switch r.Value.(type) {
	case *ModuleState:
		return 0
	case *ResourceState:
		return 1
	case *InstanceState:
		return 2
	default:
		return 50
	}
}

// stateFilterResultSlice is a slice of results that implements
// sort.Interface. The sorting goal is what is most appealing to
// human output.
type stateFilterResultSlice []*stateFilterResult

func (s stateFilterResultSlice) Len() int      { return len(s) }
func (s stateFilterResultSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s stateFilterResultSlice) Less(i, j int) bool {
	a, b := s[i], s[j]

	// if these address contain an index, we want to sort by index rather than name
	addrA, errA := parseResourceAddress(a.Address)
	addrB, errB := parseResourceAddress(b.Address)
	if errA == nil && errB == nil && addrA.Name == addrB.Name && addrA.Index != addrB.Index {
		return addrA.Index < addrB.Index
	}

	// If the addresses are different it is just lexographic sorting
	if a.Address != b.Address {
		return a.Address < b.Address
	}

	// Addresses are the same, which means it matters on the type
	return a.sortedType() < b.sortedType()
}
