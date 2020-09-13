package goheader

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

// Configuration represents go-header linter setup parameters
type Configuration struct {
	// Values is map of values. Supports two types 'const` and `regexp`. Values can be used recursively.
	Values map[string]map[string]string `yaml:"values"'`
	// Template is template for checking. Uses values.
	Template string `yaml:"template"`
	// TemplatePath path to the template file. Useful if need to load the template from a specific file.
	TemplatePath string `yaml:"template-path"`
}

func (c *Configuration) GetValues() (map[string]Value, error) {
	var result = make(map[string]Value)
	createConst := func(raw string) Value {
		return &ConstValue{RawValue: raw}
	}
	createRegexp := func(raw string) Value {
		return &RegexpValue{RawValue: raw}
	}
	appendValues := func(m map[string]string, create func(string) Value) {
		for k, v := range m {
			key := strings.ToLower(k)
			result[key] = create(v)
		}
	}
	for k, v := range c.Values {
		switch k {
		case "const":
			appendValues(v, createConst)
		case "regexp":
			appendValues(v, createRegexp)
		default:
			return nil, fmt.Errorf("unknown value type %v", k)
		}
	}
	return result, nil
}

func (c *Configuration) GetTemplate() (string, error) {
	if c.Template != "" {
		return c.Template, nil
	}
	if c.TemplatePath == "" {
		return "", errors.New("template has not passed")
	}
	if b, err := ioutil.ReadFile(c.TemplatePath); err != nil {
		return "", err
	} else {
		c.Template = string(b)
		return c.Template, nil
	}
}

func (c *Configuration) Parse(p string) error {
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, c)
}
