package resource

import (
	"encoding/json"
	"fmt"
	"strconv"

	tfjson "github.com/hashicorp/terraform-json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/addrs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/tfdiags"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

type shimmedState struct {
	state *terraform.State
}

func shimStateFromJson(jsonState *tfjson.State) (*terraform.State, error) {
	state := terraform.NewState()
	state.TFVersion = jsonState.TerraformVersion

	if jsonState.Values == nil {
		// the state is empty
		return state, nil
	}

	for key, output := range jsonState.Values.Outputs {
		os, err := shimOutputState(output)
		if err != nil {
			return nil, err
		}
		state.RootModule().Outputs[key] = os
	}

	ss := &shimmedState{state}
	err := ss.shimStateModule(jsonState.Values.RootModule)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func shimOutputState(so *tfjson.StateOutput) (*terraform.OutputState, error) {
	os := &terraform.OutputState{
		Sensitive: so.Sensitive,
	}

	switch v := so.Value.(type) {
	case string:
		os.Type = "string"
		os.Value = v
		return os, nil
	case []interface{}:
		os.Type = "list"
		if len(v) == 0 {
			os.Value = v
			return os, nil
		}
		switch firstElem := v[0].(type) {
		case string:
			elements := make([]interface{}, len(v))
			for i, el := range v {
				elements[i] = el.(string)
			}
			os.Value = elements
		case bool:
			elements := make([]interface{}, len(v))
			for i, el := range v {
				elements[i] = el.(bool)
			}
			os.Value = elements
		// unmarshalled number from JSON will always be json.Number
		case json.Number:
			elements := make([]interface{}, len(v))
			for i, el := range v {
				elements[i] = el.(json.Number)
			}
			os.Value = elements
		case []interface{}:
			os.Value = v
		case map[string]interface{}:
			os.Value = v
		default:
			return nil, fmt.Errorf("unexpected output list element type: %T", firstElem)
		}
		return os, nil
	case map[string]interface{}:
		os.Type = "map"
		os.Value = v
		return os, nil
	case bool:
		os.Type = "string"
		os.Value = strconv.FormatBool(v)
		return os, nil
	// unmarshalled number from JSON will always be json.Number
	case json.Number:
		os.Type = "string"
		os.Value = v.String()
		return os, nil
	}

	return nil, fmt.Errorf("unexpected output type: %T", so.Value)
}

func (ss *shimmedState) shimStateModule(sm *tfjson.StateModule) error {
	var path addrs.ModuleInstance

	if sm.Address == "" {
		path = addrs.RootModuleInstance
	} else {
		var diags tfdiags.Diagnostics
		path, diags = addrs.ParseModuleInstanceStr(sm.Address)
		if diags.HasErrors() {
			return diags.Err()
		}
	}

	mod := ss.state.AddModule(path)
	for _, res := range sm.Resources {
		resourceState, err := shimResourceState(res)
		if err != nil {
			return err
		}

		key, err := shimResourceStateKey(res)
		if err != nil {
			return err
		}

		mod.Resources[key] = resourceState
	}

	if len(sm.ChildModules) > 0 {
		return fmt.Errorf("Modules are not supported. Found %d modules.",
			len(sm.ChildModules))
	}
	return nil
}

func shimResourceStateKey(res *tfjson.StateResource) (string, error) {
	if res.Index == nil {
		return res.Address, nil
	}

	var mode terraform.ResourceMode
	switch res.Mode {
	case tfjson.DataResourceMode:
		mode = terraform.DataResourceMode
	case tfjson.ManagedResourceMode:
		mode = terraform.ManagedResourceMode
	default:
		return "", fmt.Errorf("unexpected resource mode for %q", res.Address)
	}

	var index int
	switch idx := res.Index.(type) {
	case json.Number:
		i, err := idx.Int64()
		if err != nil {
			return "", fmt.Errorf("unexpected index value (%q) for %q, ",
				idx, res.Address)
		}
		index = int(i)
	default:
		return "", fmt.Errorf("unexpected index type (%T) for %q, "+
			"for_each is not supported", res.Index, res.Address)
	}

	rsk := &terraform.ResourceStateKey{
		Mode:  mode,
		Type:  res.Type,
		Name:  res.Name,
		Index: index,
	}

	return rsk.String(), nil
}

func shimResourceState(res *tfjson.StateResource) (*terraform.ResourceState, error) {
	sf := &shimmedFlatmap{}
	err := sf.FromMap(res.AttributeValues)
	if err != nil {
		return nil, err
	}
	attributes := sf.Flatmap()

	if _, ok := attributes["id"]; !ok {
		return nil, fmt.Errorf("no %q found in attributes", "id")
	}

	return &terraform.ResourceState{
		Provider: res.ProviderName,
		Type:     res.Type,
		Primary: &terraform.InstanceState{
			ID:         attributes["id"],
			Attributes: attributes,
			Meta: map[string]interface{}{
				"schema_version": int(res.SchemaVersion),
			},
			Tainted: res.Tainted,
		},
		Dependencies: res.DependsOn,
	}, nil
}

type shimmedFlatmap struct {
	m map[string]string
}

func (sf *shimmedFlatmap) FromMap(attributes map[string]interface{}) error {
	if sf.m == nil {
		sf.m = make(map[string]string, len(attributes))
	}

	return sf.AddMap("", attributes)
}

func (sf *shimmedFlatmap) AddMap(prefix string, m map[string]interface{}) error {
	for key, value := range m {
		k := key
		if prefix != "" {
			k = fmt.Sprintf("%s.%s", prefix, key)
		}

		err := sf.AddEntry(k, value)
		if err != nil {
			return err
		}
	}

	mapLength := "%"
	if prefix != "" {
		mapLength = fmt.Sprintf("%s.%s", prefix, "%")
	}

	sf.AddEntry(mapLength, strconv.Itoa(len(m)))

	return nil
}

func (sf *shimmedFlatmap) AddSlice(name string, elements []interface{}) error {
	for i, elem := range elements {
		key := fmt.Sprintf("%s.%d", name, i)
		err := sf.AddEntry(key, elem)
		if err != nil {
			return err
		}
	}

	sliceLength := fmt.Sprintf("%s.#", name)
	sf.AddEntry(sliceLength, strconv.Itoa(len(elements)))

	return nil
}

func (sf *shimmedFlatmap) AddEntry(key string, value interface{}) error {
	switch el := value.(type) {
	case nil:
		// omit the entry
		return nil
	case bool:
		sf.m[key] = strconv.FormatBool(el)
	case json.Number:
		sf.m[key] = el.String()
	case string:
		sf.m[key] = el
	case map[string]interface{}:
		err := sf.AddMap(key, el)
		if err != nil {
			return err
		}
	case []interface{}:
		err := sf.AddSlice(key, el)
		if err != nil {
			return err
		}
	default:
		// This should never happen unless terraform-json
		// changes how attributes (types) are represented.
		//
		// We handle all types which the JSON unmarshaler
		// can possibly produce
		// https://golang.org/pkg/encoding/json/#Unmarshal

		return fmt.Errorf("%q: unexpected type (%T)", key, el)
	}
	return nil
}

func (sf *shimmedFlatmap) Flatmap() map[string]string {
	return sf.m
}
