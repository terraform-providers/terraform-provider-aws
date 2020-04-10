package configload

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/internal/configs"
	"github.com/hashicorp/terraform-plugin-sdk/internal/registry"
	"github.com/hashicorp/terraform-svchost/disco"
	"github.com/spf13/afero"
)

// A Loader instance is the main entry-point for loading configurations via
// this package.
//
// It extends the general config-loading functionality in the parent package
// "configs" to support installation of modules from remote sources and
// loading full configurations using modules that were previously installed.
type Loader struct {
	// parser is used to read configuration
	parser *configs.Parser

	// modules is used to install and locate descendent modules that are
	// referenced (directly or indirectly) from the root module.
	modules moduleMgr
}

// Config is used with NewLoader to specify configuration arguments for the
// loader.
type Config struct {
	// ModulesDir is a path to a directory where descendent modules are
	// (or should be) installed. (This is usually the
	// .terraform/modules directory, in the common case where this package
	// is being loaded from the main Terraform CLI package.)
	ModulesDir string

	// Services is the service discovery client to use when locating remote
	// module registry endpoints. If this is nil then registry sources are
	// not supported, which should be true only in specialized circumstances
	// such as in tests.
	Services *disco.Disco
}

// NewLoader creates and returns a loader that reads configuration from the
// real OS filesystem.
//
// The loader has some internal state about the modules that are currently
// installed, which is read from disk as part of this function. If that
// manifest cannot be read then an error will be returned.
func NewLoader(config *Config) (*Loader, error) {
	fs := afero.NewOsFs()
	parser := configs.NewParser(fs)
	reg := registry.NewClient(config.Services, nil)

	ret := &Loader{
		parser: parser,
		modules: moduleMgr{
			FS:         afero.Afero{Fs: fs},
			CanInstall: true,
			Dir:        config.ModulesDir,
			Services:   config.Services,
			Registry:   reg,
		},
	}

	err := ret.modules.readModuleManifestSnapshot()
	if err != nil {
		return nil, fmt.Errorf("failed to read module manifest: %s", err)
	}

	return ret, nil
}

// ModulesDir returns the path to the directory where the loader will look for
// the local cache of remote module packages.
func (l *Loader) ModulesDir() string {
	return l.modules.Dir
}

// RefreshModules updates the in-memory cache of the module manifest from the
// module manifest file on disk. This is not necessary in normal use because
// module installation and configuration loading are separate steps, but it
// can be useful in tests where module installation is done as a part of
// configuration loading by a helper function.
//
// Call this function after any module installation where an existing loader
// is already alive and may be used again later.
//
// An error is returned if the manifest file cannot be read.
func (l *Loader) RefreshModules() error {
	if l == nil {
		// Nothing to do, then.
		return nil
	}
	return l.modules.readModuleManifestSnapshot()
}

// Sources returns the source code cache for the underlying parser of this
// loader. This is a shorthand for l.Parser().Sources().
func (l *Loader) Sources() map[string][]byte {
	return l.parser.Sources()
}
