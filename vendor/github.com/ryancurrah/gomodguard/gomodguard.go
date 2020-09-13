package gomodguard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver"

	"golang.org/x/mod/modfile"
)

const (
	goModFilename       = "go.mod"
	errReadingGoModFile = "unable to read go mod file %s: %w"
	errParsingGoModFile = "unable to parsing go mod file %s: %w"
)

var (
	blockReasonNotInAllowedList = "import of package `%s` is blocked because the module is not in the allowed modules list."
	blockReasonInBlockedList    = "import of package `%s` is blocked because the module is in the blocked modules list."
)

// BlockedVersion has a version constraint a reason why the the module version is blocked.
type BlockedVersion struct {
	Version             string `yaml:"version"`
	Reason              string `yaml:"reason"`
	lintedModuleVersion string `yaml:"-"`
}

// Set required values for performing checks. This must be ran before running anything else.
func (r *BlockedVersion) Set(lintedModuleVersion string) {
	r.lintedModuleVersion = lintedModuleVersion
}

// IsAllowed returns true if the blocked module is allowed. You must Set() values first.
func (r *BlockedVersion) IsAllowed() bool {
	return !r.isLintedModuleVersionBlocked()
}

// isLintedModuleVersionBlocked returns true if version constraint specified and the
// linted module version meets the constraint.
func (r *BlockedVersion) isLintedModuleVersionBlocked() bool {
	if r.Version == "" {
		return false
	}

	constraint, err := semver.NewConstraint(r.Version)
	if err != nil {
		return false
	}

	version, err := semver.NewVersion(strings.TrimLeft(r.lintedModuleVersion, "v"))
	if err != nil {
		return false
	}

	return constraint.Check(version)
}

// Message returns the reason why the module version is blocked.
func (r *BlockedVersion) Message() string {
	msg := ""

	// Add version contraint to message
	msg += fmt.Sprintf("version `%s` is blocked because it does not meet the version constraint `%s`.", r.lintedModuleVersion, r.Version)

	if r.Reason == "" {
		return msg
	}

	// Add reason to message
	msg += fmt.Sprintf(" %s.", strings.TrimRight(r.Reason, "."))

	return msg
}

// BlockedModule has alternative modules to use and a reason why the module is blocked.
type BlockedModule struct {
	Recommendations   []string `yaml:"recommendations"`
	Reason            string   `yaml:"reason"`
	currentModuleName string   `yaml:"-"`
}

// Set required values for performing checks. This must be ran before running anything else.
func (r *BlockedModule) Set(currentModuleName string) {
	r.currentModuleName = currentModuleName
}

// IsAllowed returns true if the blocked module is allowed. You must Set() values first.
func (r *BlockedModule) IsAllowed() bool {
	// If the current go.mod file being linted is a recommended module of a
	// blocked module and it imports that blocked module, do not set as blocked.
	// This could mean that the linted module is a wrapper for that blocked module.
	return r.isCurrentModuleARecommendation()
}

// isCurrentModuleARecommendation returns true if the current module is in the Recommendations list.
func (r *BlockedModule) isCurrentModuleARecommendation() bool {
	if r == nil {
		return false
	}

	for n := range r.Recommendations {
		if strings.TrimSpace(r.currentModuleName) == strings.TrimSpace(r.Recommendations[n]) {
			return true
		}
	}

	return false
}

// Message returns the reason why the module is blocked and a list of recommended modules if provided.
func (r *BlockedModule) Message() string {
	msg := ""

	// Add recommendations to message
	for i := range r.Recommendations {
		switch {
		case len(r.Recommendations) == 1:
			msg += fmt.Sprintf("`%s` is a recommended module.", r.Recommendations[i])
		case (i+1) != len(r.Recommendations) && (i+1) == (len(r.Recommendations)-1):
			msg += fmt.Sprintf("`%s` ", r.Recommendations[i])
		case (i + 1) != len(r.Recommendations):
			msg += fmt.Sprintf("`%s`, ", r.Recommendations[i])
		default:
			msg += fmt.Sprintf("and `%s` are recommended modules.", r.Recommendations[i])
		}
	}

	if r.Reason == "" {
		return msg
	}

	// Add reason to message
	if msg == "" {
		msg = fmt.Sprintf("%s.", strings.TrimRight(r.Reason, "."))
	} else {
		msg += fmt.Sprintf(" %s.", strings.TrimRight(r.Reason, "."))
	}

	return msg
}

// HasRecommendations returns true if the blocked package has
// recommended modules.
func (r *BlockedModule) HasRecommendations() bool {
	if r == nil {
		return false
	}

	return len(r.Recommendations) > 0
}

// BlockedVersions a list of blocked modules by a version constraint.
type BlockedVersions []map[string]BlockedVersion

// Get returns the module names that are blocked.
func (b BlockedVersions) Get() []string {
	modules := make([]string, len(b))

	for n := range b {
		for module := range b[n] {
			modules[n] = module
			break
		}
	}

	return modules
}

// GetBlockReason returns a block version if one is set for the provided linted module name.
func (b BlockedVersions) GetBlockReason(lintedModuleName, lintedModuleVersion string) *BlockedVersion {
	for _, blockedModule := range b {
		for blockedModuleName, blockedVersion := range blockedModule {
			if strings.EqualFold(strings.TrimSpace(lintedModuleName), strings.TrimSpace(blockedModuleName)) {
				blockedVersion.Set(lintedModuleVersion)
				return &blockedVersion
			}
		}
	}

	return nil
}

// BlockedModules a list of blocked modules.
type BlockedModules []map[string]BlockedModule

// Get returns the module names that are blocked.
func (b BlockedModules) Get() []string {
	modules := make([]string, len(b))

	for n := range b {
		for module := range b[n] {
			modules[n] = module
			break
		}
	}

	return modules
}

// GetBlockReason returns a block module if one is set for the provided linted module name.
func (b BlockedModules) GetBlockReason(currentModuleName, lintedModuleName string) *BlockedModule {
	for _, blockedModule := range b {
		for blockedModuleName, blockedModule := range blockedModule {
			if strings.EqualFold(strings.TrimSpace(lintedModuleName), strings.TrimSpace(blockedModuleName)) {
				blockedModule.Set(currentModuleName)
				return &blockedModule
			}
		}
	}

	return nil
}

// Allowed is a list of modules and module
// domains that are allowed to be used.
type Allowed struct {
	Modules []string `yaml:"modules"`
	Domains []string `yaml:"domains"`
}

// IsAllowedModule returns true if the given module
// name is in the allowed modules list.
func (a *Allowed) IsAllowedModule(moduleName string) bool {
	allowedModules := a.Modules

	for i := range allowedModules {
		if strings.EqualFold(strings.TrimSpace(moduleName), strings.TrimSpace(allowedModules[i])) {
			return true
		}
	}

	return false
}

// IsAllowedModuleDomain returns true if the given modules domain is
// in the allowed module domains list.
func (a *Allowed) IsAllowedModuleDomain(moduleName string) bool {
	allowedDomains := a.Domains

	for i := range allowedDomains {
		if strings.HasPrefix(strings.TrimSpace(strings.ToLower(moduleName)), strings.TrimSpace(strings.ToLower(allowedDomains[i]))) {
			return true
		}
	}

	return false
}

// Blocked is a list of modules that are
// blocked and not to be used.
type Blocked struct {
	Modules  BlockedModules  `yaml:"modules"`
	Versions BlockedVersions `yaml:"versions"`
}

// Configuration of gomodguard allow and block lists.
type Configuration struct {
	Allowed Allowed `yaml:"allowed"`
	Blocked Blocked `yaml:"blocked"`
}

// Result represents the result of one error.
type Result struct {
	FileName   string
	LineNumber int
	Position   token.Position
	Reason     string
}

// String returns the filename, line
// number and reason of a Result.
func (r *Result) String() string {
	return fmt.Sprintf("%s:%d:1 %s", r.FileName, r.LineNumber, r.Reason)
}

// Processor processes Go files.
type Processor struct {
	Config                    Configuration
	Logger                    *log.Logger
	Modfile                   *modfile.File
	blockedModulesFromModFile map[string][]string
	Result                    []Result
}

// NewProcessor will create a Processor to lint blocked packages.
func NewProcessor(config Configuration, logger *log.Logger) (*Processor, error) {
	goModFileBytes, err := loadGoModFile()
	if err != nil {
		return nil, fmt.Errorf(errReadingGoModFile, goModFilename, err)
	}

	mfile, err := modfile.Parse(goModFilename, goModFileBytes, nil)
	if err != nil {
		return nil, fmt.Errorf(errParsingGoModFile, goModFilename, err)
	}

	logger.Printf("info: allowed modules, %+v", config.Allowed.Modules)
	logger.Printf("info: allowed module domains, %+v", config.Allowed.Domains)
	logger.Printf("info: blocked modules, %+v", config.Blocked.Modules.Get())
	logger.Printf("info: blocked modules with version constraints, %+v", config.Blocked.Versions.Get())

	p := &Processor{
		Config:  config,
		Logger:  logger,
		Modfile: mfile,
		Result:  []Result{},
	}

	p.SetBlockedModulesFromModFile()

	return p, nil
}

// ProcessFiles takes a string slice with file names (full paths)
// and lints them.
func (p *Processor) ProcessFiles(filenames []string) []Result {
	pluralModuleMsg := "s"
	if len(p.blockedModulesFromModFile) == 1 {
		pluralModuleMsg = ""
	}

	blockedModules := make([]string, 0, len(p.blockedModulesFromModFile))
	for blockedModuleName := range p.blockedModulesFromModFile {
		blockedModules = append(blockedModules, blockedModuleName)
	}

	p.Logger.Printf("info: found %d blocked module%s in %s: %+v",
		len(p.blockedModulesFromModFile), pluralModuleMsg, goModFilename, blockedModules)

	for _, filename := range filenames {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			p.Result = append(p.Result, Result{
				FileName:   filename,
				LineNumber: 0,
				Reason:     fmt.Sprintf("unable to read file, file cannot be linted (%s)", err.Error()),
			})
		}

		p.process(filename, data)
	}

	return p.Result
}

// process file imports and add lint error if blocked package is imported.
func (p *Processor) process(filename string, data []byte) {
	fileSet := token.NewFileSet()

	file, err := parser.ParseFile(fileSet, filename, data, parser.ParseComments)
	if err != nil {
		p.Result = append(p.Result, Result{
			FileName:   filename,
			LineNumber: 0,
			Reason:     fmt.Sprintf("invalid syntax, file cannot be linted (%s)", err.Error()),
		})

		return
	}

	imports := file.Imports
	for n := range imports {
		importedPkg := strings.TrimSpace(strings.Trim(imports[n].Path.Value, "\""))

		blockReasons := p.isBlockedPackageFromModFile(importedPkg)
		if blockReasons == nil {
			continue
		}

		for _, blockReason := range blockReasons {
			p.addError(fileSet, imports[n].Pos(), blockReason)
		}
	}
}

// addError adds an error for the file and line number for the current token.Pos
// with the given reason.
func (p *Processor) addError(fileset *token.FileSet, pos token.Pos, reason string) {
	position := fileset.Position(pos)

	p.Result = append(p.Result, Result{
		FileName:   position.Filename,
		LineNumber: position.Line,
		Position:   position,
		Reason:     reason,
	})
}

// SetBlockedModulesFromModFile determines which modules are blocked by reading
// the go.mod file and comparing the require modules to the allowed modules.
func (p *Processor) SetBlockedModulesFromModFile() {
	blockedModules := make(map[string][]string, len(p.Modfile.Require))
	currentModuleName := p.Modfile.Module.Mod.Path
	lintedModules := p.Modfile.Require

	for i := range lintedModules {
		if lintedModules[i].Indirect {
			continue
		}

		lintedModuleName := strings.TrimSpace(lintedModules[i].Mod.Path)
		lintedModuleVersion := strings.TrimSpace(lintedModules[i].Mod.Version)

		var isAllowed bool

		switch {
		case len(p.Config.Allowed.Modules) == 0 && len(p.Config.Allowed.Domains) == 0:
			isAllowed = true
		case p.Config.Allowed.IsAllowedModuleDomain(lintedModuleName):
			isAllowed = true
		case p.Config.Allowed.IsAllowedModule(lintedModuleName):
			isAllowed = true
		default:
			isAllowed = false
		}

		blockModuleReason := p.Config.Blocked.Modules.GetBlockReason(currentModuleName, lintedModuleName)
		blockVersionReason := p.Config.Blocked.Versions.GetBlockReason(lintedModuleName, lintedModuleVersion)

		if !isAllowed && blockModuleReason == nil && blockVersionReason == nil {
			blockedModules[lintedModuleName] = append(blockedModules[lintedModuleName], blockReasonNotInAllowedList)
			continue
		}

		if blockModuleReason != nil && !blockModuleReason.IsAllowed() {
			blockedModules[lintedModuleName] = append(blockedModules[lintedModuleName], fmt.Sprintf("%s %s", blockReasonInBlockedList, blockModuleReason.Message()))
		}

		if blockVersionReason != nil && !blockVersionReason.IsAllowed() {
			blockedModules[lintedModuleName] = append(blockedModules[lintedModuleName], fmt.Sprintf("%s %s", blockReasonInBlockedList, blockVersionReason.Message()))
		}
	}

	p.blockedModulesFromModFile = blockedModules
}

// isBlockedPackageFromModFile returns the block reason if the package is blocked.
func (p *Processor) isBlockedPackageFromModFile(packageName string) []string {
	for blockedModuleName, blockReasons := range p.blockedModulesFromModFile {
		if strings.HasPrefix(strings.TrimSpace(packageName), strings.TrimSpace(blockedModuleName)) {
			formattedReasons := make([]string, 0, len(blockReasons))

			for _, blockReason := range blockReasons {
				formattedReasons = append(formattedReasons, fmt.Sprintf(blockReason, packageName))
			}

			return formattedReasons
		}
	}

	return nil
}

func loadGoModFile() ([]byte, error) {
	cmd := exec.Command("go", "env", "-json")
	stdout, _ := cmd.StdoutPipe()
	_ = cmd.Start()

	if stdout == nil {
		return ioutil.ReadFile(goModFilename)
	}

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(stdout)

	goEnv := make(map[string]string)

	err := json.Unmarshal(buf.Bytes(), &goEnv)
	if err != nil {
		return ioutil.ReadFile(goModFilename)
	}

	if _, ok := goEnv["GOMOD"]; !ok {
		return ioutil.ReadFile(goModFilename)
	}

	if _, err := os.Stat(goEnv["GOMOD"]); os.IsNotExist(err) {
		return ioutil.ReadFile(goModFilename)
	}

	return ioutil.ReadFile(goEnv["GOMOD"])
}
