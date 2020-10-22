package terraform

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// resourceAddress is a way of identifying an individual resource (or,
// eventually, a subset of resources) within the state. It is used for Targets.
type resourceAddress struct {
	// Addresses a resource falling somewhere in the module path
	// When specified alone, addresses all resources within a module path
	Path []string

	// Addresses a specific resource that occurs in a list
	Index int

	InstanceType    instanceType
	InstanceTypeSet bool
	Name            string
	Type            string
	Mode            ResourceMode // significant only if InstanceTypeSet
}

// String outputs the address that parses into this address.
func (r *resourceAddress) String() string {
	var result []string
	for _, p := range r.Path {
		result = append(result, "module", p)
	}

	switch r.Mode {
	case ManagedResourceMode:
		// nothing to do
	case DataResourceMode:
		result = append(result, "data")
	default:
		panic(fmt.Errorf("unsupported resource mode %s", r.Mode))
	}

	if r.Type != "" {
		result = append(result, r.Type)
	}

	if r.Name != "" {
		name := r.Name
		if r.InstanceTypeSet {
			switch r.InstanceType {
			case typePrimary:
				name += ".primary"
			case typeDeposed:
				name += ".deposed"
			case typeTainted:
				name += ".tainted"
			}
		}

		if r.Index >= 0 {
			name += fmt.Sprintf("[%d]", r.Index)
		}
		result = append(result, name)
	}

	return strings.Join(result, ".")
}

func parseResourceAddress(s string) (*resourceAddress, error) {
	matches, err := tokenizeResourceAddress(s)
	if err != nil {
		return nil, err
	}
	mode := ManagedResourceMode
	if matches["data_prefix"] != "" {
		mode = DataResourceMode
	}
	resourceIndex, err := parseResourceIndex(matches["index"])
	if err != nil {
		return nil, err
	}
	instanceType, err := parseInstanceType(matches["instance_type"])
	if err != nil {
		return nil, err
	}
	path := parseResourcePath(matches["path"])

	// not allowed to say "data." without a type following
	if mode == DataResourceMode && matches["type"] == "" {
		return nil, fmt.Errorf(
			"invalid resource address %q: must target specific data instance",
			s,
		)
	}

	return &resourceAddress{
		Path:            path,
		Index:           resourceIndex,
		InstanceType:    instanceType,
		InstanceTypeSet: matches["instance_type"] != "",
		Name:            matches["name"],
		Type:            matches["type"],
		Mode:            mode,
	}, nil
}

// Less returns true if and only if the receiver should be sorted before
// the given address when presenting a list of resource addresses to
// an end-user.
//
// This sort uses lexicographic sorting for most components, but uses
// numeric sort for indices, thus causing index 10 to sort after
// index 9, rather than after index 1.
func (addr *resourceAddress) Less(other *resourceAddress) bool {

	switch {

	case len(addr.Path) != len(other.Path):
		return len(addr.Path) < len(other.Path)

	case !reflect.DeepEqual(addr.Path, other.Path):
		// If the two paths are the same length but don't match, we'll just
		// cheat and compare the string forms since it's easier than
		// comparing all of the path segments in turn, and lexicographic
		// comparison is correct for the module path portion.
		addrStr := addr.String()
		otherStr := other.String()
		return addrStr < otherStr

	case addr.Mode != other.Mode:
		return addr.Mode == DataResourceMode

	case addr.Type != other.Type:
		return addr.Type < other.Type

	case addr.Name != other.Name:
		return addr.Name < other.Name

	case addr.Index != other.Index:
		// Since "Index" is -1 for an un-indexed address, this also conveniently
		// sorts unindexed addresses before indexed ones, should they both
		// appear for some reason.
		return addr.Index < other.Index

	case addr.InstanceTypeSet != other.InstanceTypeSet:
		return !addr.InstanceTypeSet

	case addr.InstanceType != other.InstanceType:
		// InstanceType is actually an enum, so this is just an arbitrary
		// sort based on the enum numeric values, and thus not particularly
		// meaningful.
		return addr.InstanceType < other.InstanceType

	default:
		return false

	}
}

func parseResourceIndex(s string) (int, error) {
	if s == "" {
		return -1, nil
	}
	return strconv.Atoi(s)
}

func parseResourcePath(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ".")
	path := make([]string, 0, len(parts))
	for _, s := range parts {
		// Due to the limitations of the regexp match below, the path match has
		// some noise in it we have to filter out :|
		if s == "" || s == "module" {
			continue
		}
		path = append(path, s)
	}
	return path
}

func parseInstanceType(s string) (instanceType, error) {
	switch s {
	case "", "primary":
		return typePrimary, nil
	case "deposed":
		return typeDeposed, nil
	case "tainted":
		return typeTainted, nil
	default:
		return typeInvalid, fmt.Errorf("Unexpected value for instanceType field: %q", s)
	}
}

func tokenizeResourceAddress(s string) (map[string]string, error) {
	// Example of portions of the regexp below using the
	// string "aws_instance.web.tainted[1]"
	re := regexp.MustCompile(`\A` +
		// "module.foo.module.bar" (optional)
		`(?P<path>(?:module\.(?P<module_name>[^.]+)\.?)*)` +
		// possibly "data.", if targeting is a data resource
		`(?P<data_prefix>(?:data\.)?)` +
		// "aws_instance.web" (optional when module path specified)
		`(?:(?P<type>[^.]+)\.(?P<name>[^.[]+))?` +
		// "tainted" (optional, omission implies: "primary")
		`(?:\.(?P<instance_type>\w+))?` +
		// "1" (optional, omission implies: "0")
		`(?:\[(?P<index>\d+)\])?` +
		`\z`)

	groupNames := re.SubexpNames()
	rawMatches := re.FindAllStringSubmatch(s, -1)
	if len(rawMatches) != 1 {
		return nil, fmt.Errorf("invalid resource address %q", s)
	}

	matches := make(map[string]string)
	for i, m := range rawMatches[0] {
		matches[groupNames[i]] = m
	}

	return matches, nil
}
