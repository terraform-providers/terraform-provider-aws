package tfexec

import (
	"encoding/json"
)

// AllowMissingConfigOption represents the -allow-missing-config flag.
type AllowMissingConfigOption struct {
	allowMissingConfig bool
}

// AllowMissingConfig represents the -allow-missing-config flag.
func AllowMissingConfig(allowMissingConfig bool) *AllowMissingConfigOption {
	return &AllowMissingConfigOption{allowMissingConfig}
}

// BackendOption represents the -backend flag.
type BackendOption struct {
	backend bool
}

// Backend represents the -backend flag.
func Backend(backend bool) *BackendOption {
	return &BackendOption{backend}
}

// BackendConfigOption represents the -backend-config flag.
type BackendConfigOption struct {
	path string
}

// BackendConfig represents the -backend-config flag.
func BackendConfig(backendConfig string) *BackendConfigOption {
	return &BackendConfigOption{backendConfig}
}

// BackupOption represents the -backup flag.
type BackupOption struct {
	path string
}

// Backup represents the -backup flag.
func Backup(path string) *BackupOption {
	return &BackupOption{path}
}

// DisableBackup is a convenience method for Backup("-"), indicating backup state should be disabled.
func DisableBackup() *BackupOption {
	return &BackupOption{"-"}
}

// ConfigOption represents the -config flag.
type ConfigOption struct {
	path string
}

// Config represents the -config flag.
func Config(path string) *ConfigOption {
	return &ConfigOption{path}
}

// CopyStateOption represents the -state flag for terraform workspace new. This flag is used
// to copy an existing state file in to the new workspace.
type CopyStateOption struct {
	path string
}

// CopyState represents the -state flag for terraform workspace new. This flag is used
// to copy an existing state file in to the new workspace.
func CopyState(path string) *CopyStateOption {
	return &CopyStateOption{path}
}

type DirOption struct {
	path string
}

func Dir(path string) *DirOption {
	return &DirOption{path}
}

type DirOrPlanOption struct {
	path string
}

func DirOrPlan(path string) *DirOrPlanOption {
	return &DirOrPlanOption{path}
}

// DestroyFlagOption represents the -destroy flag.
type DestroyFlagOption struct {
	// named to prevent conflict with DestroyOption interface

	destroy bool
}

// Destroy represents the -destroy flag.
func Destroy(destroy bool) *DestroyFlagOption {
	return &DestroyFlagOption{destroy}
}

type ForceCopyOption struct {
	forceCopy bool
}

func ForceCopy(forceCopy bool) *ForceCopyOption {
	return &ForceCopyOption{forceCopy}
}

type FromModuleOption struct {
	source string
}

func FromModule(source string) *FromModuleOption {
	return &FromModuleOption{source}
}

type GetOption struct {
	get bool
}

func Get(get bool) *GetOption {
	return &GetOption{get}
}

type GetPluginsOption struct {
	getPlugins bool
}

func GetPlugins(getPlugins bool) *GetPluginsOption {
	return &GetPluginsOption{getPlugins}
}

// LockOption represents the -lock flag.
type LockOption struct {
	lock bool
}

// Lock represents the -lock flag.
func Lock(lock bool) *LockOption {
	return &LockOption{lock}
}

// LockTimeoutOption represents the -lock-timeout flag.
type LockTimeoutOption struct {
	timeout string
}

// LockTimeout represents the -lock-timeout flag.
func LockTimeout(lockTimeout string) *LockTimeoutOption {
	// TODO: should this just use a duration instead?
	return &LockTimeoutOption{lockTimeout}
}

type OutOption struct {
	path string
}

func Out(path string) *OutOption {
	return &OutOption{path}
}

type ParallelismOption struct {
	parallelism int
}

func Parallelism(n int) *ParallelismOption {
	return &ParallelismOption{n}
}

type PluginDirOption struct {
	pluginDir string
}

func PluginDir(pluginDir string) *PluginDirOption {
	return &PluginDirOption{pluginDir}
}

type ReattachInfo map[string]ReattachConfig

// ReattachConfig holds the information Terraform needs to be able to attach
// itself to a provider process, so it can drive the process.
type ReattachConfig struct {
	Protocol string
	Pid      int
	Test     bool
	Addr     ReattachConfigAddr
}

// ReattachConfigAddr is a JSON-encoding friendly version of net.Addr.
type ReattachConfigAddr struct {
	Network string
	String  string
}

type ReattachOption struct {
	info ReattachInfo
}

func (info ReattachInfo) marshalString() (string, error) {
	reattachStr, err := json.Marshal(info)
	if err != nil {
		return "", err
	}
	return string(reattachStr), nil
}

func Reattach(info ReattachInfo) *ReattachOption {
	return &ReattachOption{info}
}

type ReconfigureOption struct {
	reconfigure bool
}

func Reconfigure(reconfigure bool) *ReconfigureOption {
	return &ReconfigureOption{reconfigure}
}

type RefreshOption struct {
	refresh bool
}

func Refresh(refresh bool) *RefreshOption {
	return &RefreshOption{refresh}
}

type StateOption struct {
	path string
}

// State represents the -state flag.
//
// Deprecated: The -state CLI flag is a legacy flag and should not be used.
// If you need a different state file for every run, you can instead use the
// local backend.
// See https://github.com/hashicorp/terraform/issues/25920#issuecomment-676560799
func State(path string) *StateOption {
	return &StateOption{path}
}

type StateOutOption struct {
	path string
}

func StateOut(path string) *StateOutOption {
	return &StateOutOption{path}
}

type TargetOption struct {
	target string
}

func Target(resource string) *TargetOption {
	return &TargetOption{resource}
}

type UpgradeOption struct {
	upgrade bool
}

func Upgrade(upgrade bool) *UpgradeOption {
	return &UpgradeOption{upgrade}
}

type VarOption struct {
	assignment string
}

func Var(assignment string) *VarOption {
	return &VarOption{assignment}
}

type VarFileOption struct {
	path string
}

func VarFile(path string) *VarFileOption {
	return &VarFileOption{path}
}

type VerifyPluginsOption struct {
	verifyPlugins bool
}

func VerifyPlugins(verifyPlugins bool) *VerifyPluginsOption {
	return &VerifyPluginsOption{verifyPlugins}
}
