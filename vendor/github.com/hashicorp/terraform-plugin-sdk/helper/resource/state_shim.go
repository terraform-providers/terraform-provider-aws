package resource

import (
	"encoding/json"
	"fmt"
	"strconv"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/internal/addrs"
	"github.com/hashicorp/terraform-plugin-sdk/internal/configs/hcl2shim"
	"github.com/hashicorp/terraform-plugin-sdk/internal/states"
	"github.com/hashicorp/terraform-plugin-sdk/internal/tfdiags"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/zclconf/go-cty/cty"
)

// shimState takes a new *states.State and reverts it to a legacy state for the provider ACC tests
func shimNewState(newState *states.State, providers map[string]terraform.ResourceProvider) (*terraform.State, error) {
	state := terraform.NewState()

	// in the odd case of a nil state, let the helper packages handle it
	if newState == nil {
		return nil, nil
	}

	for _, newMod := range newState.Modules {
		mod := state.AddModule(newMod.Addr)

		for name, out := range newMod.OutputValues {
			outputType := ""
			val := hcl2shim.ConfigValueFromHCL2(out.Value)
			ty := out.Value.Type()
			switch {
			case ty == cty.String:
				outputType = "string"
			case ty.IsTupleType() || ty.IsListType():
				outputType = "list"
			case ty.IsMapType():
				outputType = "map"
			}

			mod.Outputs[name] = &terraform.OutputState{
				Type:      outputType,
				Value:     val,
				Sensitive: out.Sensitive,
			}
		}

		for _, res := range newMod.Resources {
			resType := res.Addr.Type
			providerType := res.ProviderConfig.ProviderConfig.Type

			resource := getResource(providers, providerType, res.Addr)

			for key, i := range res.Instances {
				resState := &terraform.ResourceState{
					Type:     resType,
					Provider: res.ProviderConfig.String(),
				}

				// We should always have a Current instance here, but be safe about checking.
				if i.Current != nil {
					flatmap, err := shimmedAttributes(i.Current, resource)
					if err != nil {
						return nil, fmt.Errorf("error decoding state for %q: %s", resType, err)
					}

					var meta map[string]interface{}
					if i.Current.Private != nil {
						err := json.Unmarshal(i.Current.Private, &meta)
						if err != nil {
							return nil, err
						}
					}

					resState.Primary = &terraform.InstanceState{
						ID:         flatmap["id"],
						Attributes: flatmap,
						Tainted:    i.Current.Status == states.ObjectTainted,
						Meta:       meta,
					}

					if i.Current.SchemaVersion != 0 {
						if resState.Primary.Meta == nil {
							resState.Primary.Meta = map[string]interface{}{}
						}
						resState.Primary.Meta["schema_version"] = i.Current.SchemaVersion
					}

					for _, dep := range i.Current.Dependencies {
						resState.Dependencies = append(resState.Dependencies, dep.String())
					}

					// convert the indexes to the old style flapmap indexes
					idx := ""
					switch key.(type) {
					case addrs.IntKey:
						// don't add numeric index values to resources with a count of 0
						if len(res.Instances) > 1 {
							idx = fmt.Sprintf(".%d", key)
						}
					case addrs.StringKey:
						idx = "." + key.String()
					}

					mod.Resources[res.Addr.String()+idx] = resState
				}

				// add any deposed instances
				for _, dep := range i.Deposed {
					flatmap, err := shimmedAttributes(dep, resource)
					if err != nil {
						return nil, fmt.Errorf("error decoding deposed state for %q: %s", resType, err)
					}

					var meta map[string]interface{}
					if dep.Private != nil {
						err := json.Unmarshal(dep.Private, &meta)
						if err != nil {
							return nil, err
						}
					}

					deposed := &terraform.InstanceState{
						ID:         flatmap["id"],
						Attributes: flatmap,
						Tainted:    dep.Status == states.ObjectTainted,
						Meta:       meta,
					}
					if dep.SchemaVersion != 0 {
						deposed.Meta = map[string]interface{}{
							"schema_version": dep.SchemaVersion,
						}
					}

					resState.Deposed = append(resState.Deposed, deposed)
				}
			}
		}
	}

	return state, nil
}

func getResource(providers map[string]terraform.ResourceProvider, providerName string, addr addrs.Resource) *schema.Resource {
	p := providers[providerName]
	if p == nil {
		panic(fmt.Sprintf("provider %q not found in test step", providerName))
	}

	// this is only for tests, so should only see schema.Providers
	provider := p.(*schema.Provider)

	switch addr.Mode {
	case addrs.ManagedResourceMode:
		resource := provider.ResourcesMap[addr.Type]
		if resource != nil {
			return resource
		}
	case addrs.DataResourceMode:
		resource := provider.DataSourcesMap[addr.Type]
		if resource != nil {
			return resource
		}
	}

	panic(fmt.Sprintf("resource %s not found in test step", addr.Type))
}

func shimmedAttributes(instance *states.ResourceInstanceObjectSrc, res *schema.Resource) (map[string]string, error) {
	flatmap := instance.AttrsFlat
	if flatmap != nil {
		return flatmap, nil
	}

	// if we have json attrs, they need to be decoded
	rio, err := instance.Decode(res.CoreConfigSchema().ImpliedType())
	if err != nil {
		return nil, err
	}

	instanceState, err := res.ShimInstanceStateFromValue(rio.Value)
	if err != nil {
		return nil, err
	}

	return instanceState.Attributes, nil
}

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
		// unmarshalled number from JSON will always be float64
		case float64:
			elements := make([]interface{}, len(v))
			for i, el := range v {
				elements[i] = el.(float64)
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
	// unmarshalled number from JSON will always be float64
	case float64:
		os.Type = "string"
		os.Value = strconv.FormatFloat(v, 'f', -1, 64)
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
	case float64:
		index = int(idx)
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
	case float64:
		sf.m[key] = strconv.FormatFloat(el, 'f', -1, 64)
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
