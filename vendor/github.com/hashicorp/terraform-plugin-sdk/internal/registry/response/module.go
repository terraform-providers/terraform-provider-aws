package response

// ModuleSubmodule is the metadata about a specific submodule within
// a module. This includes the root module as a special case.
type ModuleSubmodule struct {
	Path   string `json:"path"`
	Readme string `json:"readme"`
	Empty  bool   `json:"empty"`

	Inputs       []*ModuleInput    `json:"inputs"`
	Outputs      []*ModuleOutput   `json:"outputs"`
	Dependencies []*ModuleDep      `json:"dependencies"`
	Resources    []*ModuleResource `json:"resources"`
}

// ModuleInput is an input for a module.
type ModuleInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Default     string `json:"default"`
}

// ModuleOutput is an output for a module.
type ModuleOutput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ModuleDep is an output for a module.
type ModuleDep struct {
	Name    string `json:"name"`
	Source  string `json:"source"`
	Version string `json:"version"`
}

// ModuleProviderDep is the output for a provider dependency
type ModuleProviderDep struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ModuleResource is an output for a module.
type ModuleResource struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
