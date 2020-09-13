# go-header
[![Actions Status](https://github.com/denis-tingajkin/go-header/workflows/ci/badge.svg)](https://github.com/denis-tingajkin/go-header/actions)

Go source code linter providing checks for license headers.

# Installation

For installation you can simply use `go get`.

```
go get github.com/denis-tingajkin/go-header/cmd/go-header
```

# Configuration

To configuring `go-header.yml` linter you simply need to fill the next structures in YAML format.
```go
// Configuration represents go-header linter setup parameters
type Configuration struct {
	// Values is map of values. Supports two types 'const` and `regexp`. Values can be used recursively.
	Values       map[string]map[string]string `yaml:"values"'`
	// Template is template for checking. Uses values.
	Template     string                       `yaml:"template"`
	// TemplatePath path to the template file. Useful if need to load the template from a specific file.
	TemplatePath string                       `yaml:"template-path"`
}
```
Where supported two kinds of values: `const` and `regexp`. NOTE: values can be used recursively. 
Values ​​with type `const` checks on equality.
Values ​​with type `regexp` checks on the match.

# Execution

`go-header` linter expects file path on input. If you want to run `go-header` only on diff files, then you can use this command
```bash
go-header $(git diff --name-only)
```

# Setup example

## Step 1
Create configuration file  `.go-header.yaml` in the root of project.
```yaml
---
values:
  const:
    MY COMPANY: mycompany.com
template-path: ./mypath/mytemplate.txt
```
## Step 2 
Write the template file. For example for config above `mytemplate.txt` could be
```text
{{ MY COMPANY }}
SPDX-License-Identifier: Apache-2.0

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at:

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
## Step 3 
You are ready! Execute `go-header {FILES}` from the root of the project. 
