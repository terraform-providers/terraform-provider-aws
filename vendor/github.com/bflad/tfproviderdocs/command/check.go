package command

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/bflad/tfproviderdocs/check"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/mitchellh/cli"
)

type CheckCommandConfig struct {
	AllowedGuideSubcategories        string
	AllowedGuideSubcategoriesFile    string
	AllowedResourceSubcategories     string
	AllowedResourceSubcategoriesFile string
	IgnoreSideNavigationDataSources  string
	IgnoreSideNavigationResources    string
	LogLevel                         string
	Path                             string
	ProviderName                     string
	ProvidersSchemaJson              string
	RequireGuideSubcategory          bool
	RequireResourceSubcategory       bool
}

// CheckCommand is a Command implementation
type CheckCommand struct {
	Ui cli.Ui
}

func (*CheckCommand) Help() string {
	optsBuffer := bytes.NewBuffer([]byte{})
	opts := tabwriter.NewWriter(optsBuffer, 0, 0, 1, ' ', 0)
	LogLevelFlagHelp(opts)
	fmt.Fprintf(opts, CommandHelpOptionFormat, "-allowed-guide-subcategories", "Comma separated list of allowed guide frontmatter subcategories.")
	fmt.Fprintf(opts, CommandHelpOptionFormat, "-allowed-guide-subcategories-file", "Path to newline separated file of allowed guide frontmatter subcategories.")
	fmt.Fprintf(opts, CommandHelpOptionFormat, "-allowed-resource-subcategories", "Comma separated list of allowed data source and resource frontmatter subcategories.")
	fmt.Fprintf(opts, CommandHelpOptionFormat, "-allowed-resource-subcategories-file", "Path to newline separated file of allowed data source and resource frontmatter subcategories.")
	fmt.Fprintf(opts, CommandHelpOptionFormat, "-ignore-side-navigation-data-sources", "Comma separated list of data sources to ignore side navigation validation.")
	fmt.Fprintf(opts, CommandHelpOptionFormat, "-ignore-side-navigation-resources", "Comma separated list of resources to ignore side navigation validation.")
	fmt.Fprintf(opts, CommandHelpOptionFormat, "-provider-name", "Terraform Provider name. Automatically determined if current working directory or provided path is prefixed with terraform-provider-*.")
	fmt.Fprintf(opts, CommandHelpOptionFormat, "-providers-schema-json", "Path to terraform providers schema -json file. Enables enhanced validations.")
	fmt.Fprintf(opts, CommandHelpOptionFormat, "-require-guide-subcategory", "Require guide frontmatter subcategory.")
	fmt.Fprintf(opts, CommandHelpOptionFormat, "-require-resource-subcategory", "Require data source and resource frontmatter subcategory.")
	opts.Flush()

	helpText := fmt.Sprintf(`
Usage: tfproviderdocs check [options] [PATH]

  Performs documentation directory and file checks against the given Terraform Provider codebase.

Options:

%s
`, optsBuffer.String())

	return strings.TrimSpace(helpText)
}

func (c *CheckCommand) Name() string { return "check" }

func (c *CheckCommand) Run(args []string) int {
	var config CheckCommandConfig

	flags := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	flags.Usage = func() { c.Ui.Info(c.Help()) }
	LogLevelFlag(flags, &config.LogLevel)
	flags.StringVar(&config.AllowedGuideSubcategories, "allowed-guide-subcategories", "", "")
	flags.StringVar(&config.AllowedGuideSubcategoriesFile, "allowed-guide-subcategories-file", "", "")
	flags.StringVar(&config.AllowedResourceSubcategories, "allowed-resource-subcategories", "", "")
	flags.StringVar(&config.AllowedResourceSubcategoriesFile, "allowed-resource-subcategories-file", "", "")
	flags.StringVar(&config.IgnoreSideNavigationDataSources, "ignore-side-navigation-data-sources", "", "")
	flags.StringVar(&config.IgnoreSideNavigationResources, "ignore-side-navigation-resources", "", "")
	flags.StringVar(&config.ProviderName, "provider-name", "", "")
	flags.StringVar(&config.ProvidersSchemaJson, "providers-schema-json", "", "")
	flags.BoolVar(&config.RequireGuideSubcategory, "require-guide-subcategory", false, "")
	flags.BoolVar(&config.RequireResourceSubcategory, "require-resource-subcategory", false, "")

	if err := flags.Parse(args); err != nil {
		flags.Usage()
		return 1
	}

	args = flags.Args()

	if len(args) == 1 {
		config.Path = args[0]
	}

	ConfigureLogging(c.Name(), config.LogLevel)

	if config.ProviderName == "" {
		if config.Path == "" {
			config.ProviderName = providerNameFromCurrentDirectory()
		} else {
			config.ProviderName = providerNameFromPath(config.Path)
		}

		if config.ProviderName == "" {
			log.Printf("[WARN] Unable to determine provider name. Enhanced validations may fail.")
		} else {
			log.Printf("[DEBUG] Found provider name: %s", config.ProviderName)
		}
	}

	directories, err := check.GetDirectories(config.Path)

	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error getting Terraform Provider documentation directories: %s", err))
		return 1
	}

	if len(directories) == 0 {
		if config.Path == "" {
			c.Ui.Error("No Terraform Provider documentation directories found in current path")
		} else {
			c.Ui.Error(fmt.Sprintf("No Terraform Provider documentation directories found in path: %s", config.Path))
		}

		return 1
	}

	var allowedGuideSubcategories, allowedResourceSubcategories, ignoreSideNavigationDataSources, ignoreSideNavigationResources []string

	if v := config.AllowedGuideSubcategories; v != "" {
		allowedGuideSubcategories = strings.Split(v, ",")
	}

	if v := config.AllowedGuideSubcategoriesFile; v != "" {
		var err error
		allowedGuideSubcategories, err = allowedSubcategoriesFile(v)

		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error getting allowed guide subcategories: %s", err))
			return 1
		}
	}

	if v := config.AllowedResourceSubcategories; v != "" {
		allowedResourceSubcategories = strings.Split(v, ",")
	}

	if v := config.AllowedResourceSubcategoriesFile; v != "" {
		var err error
		allowedResourceSubcategories, err = allowedSubcategoriesFile(v)

		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error getting allowed resource subcategories: %s", err))
			return 1
		}
	}

	if v := config.IgnoreSideNavigationDataSources; v != "" {
		ignoreSideNavigationDataSources = strings.Split(v, ",")
	}

	if v := config.IgnoreSideNavigationResources; v != "" {
		ignoreSideNavigationResources = strings.Split(v, ",")
	}

	fileOpts := &check.FileOptions{
		BasePath: config.Path,
	}
	checkOpts := &check.CheckOptions{
		LegacyDataSourceFile: &check.LegacyDataSourceFileOptions{
			FileOptions: fileOpts,
			FrontMatter: &check.FrontMatterOptions{
				AllowedSubcategories: allowedResourceSubcategories,
				RequireSubcategory:   config.RequireResourceSubcategory,
			},
		},
		LegacyGuideFile: &check.LegacyGuideFileOptions{
			FileOptions: fileOpts,
			FrontMatter: &check.FrontMatterOptions{
				AllowedSubcategories: allowedGuideSubcategories,
				RequireSubcategory:   config.RequireGuideSubcategory,
			},
		},
		LegacyIndexFile: &check.LegacyIndexFileOptions{
			FileOptions: fileOpts,
		},
		LegacyResourceFile: &check.LegacyResourceFileOptions{
			FileOptions: fileOpts,
			FrontMatter: &check.FrontMatterOptions{
				AllowedSubcategories: allowedResourceSubcategories,
				RequireSubcategory:   config.RequireResourceSubcategory,
			},
		},
		ProviderName: config.ProviderName,
		RegistryDataSourceFile: &check.RegistryDataSourceFileOptions{
			FileOptions: fileOpts,
			FrontMatter: &check.FrontMatterOptions{
				AllowedSubcategories: allowedResourceSubcategories,
				RequireSubcategory:   config.RequireResourceSubcategory,
			},
		},
		RegistryGuideFile: &check.RegistryGuideFileOptions{
			FileOptions: fileOpts,
			FrontMatter: &check.FrontMatterOptions{
				AllowedSubcategories: allowedGuideSubcategories,
				RequireSubcategory:   config.RequireGuideSubcategory,
			},
		},
		RegistryIndexFile: &check.RegistryIndexFileOptions{
			FileOptions: fileOpts,
		},
		RegistryResourceFile: &check.RegistryResourceFileOptions{
			FileOptions: fileOpts,
			FrontMatter: &check.FrontMatterOptions{
				AllowedSubcategories: allowedResourceSubcategories,
				RequireSubcategory:   config.RequireResourceSubcategory,
			},
		},
		SideNavigation: &check.SideNavigationOptions{
			FileOptions:       fileOpts,
			IgnoreDataSources: ignoreSideNavigationDataSources,
			IgnoreResources:   ignoreSideNavigationResources,
			ProviderName:      config.ProviderName,
		},
	}

	if config.ProvidersSchemaJson != "" {
		ps, err := providerSchemas(config.ProvidersSchemaJson)

		if err != nil {
			c.Ui.Error(fmt.Sprintf("Error enabling Terraform Provider schema checks: %s", err))
			return 1
		}

		if config.ProviderName == "" {
			msg := `Unknown provider name for enabling Terraform Provider schema checks.

Check that the current working directory or provided path is prefixed with terraform-provider-*.`
			c.Ui.Error(msg)
			return 1
		}

		checkOpts.SchemaDataSources = providerSchemasDataSources(ps, config.ProviderName)
		checkOpts.SchemaResources = providerSchemasResources(ps, config.ProviderName)
	}

	if err := check.NewCheck(checkOpts).Run(directories); err != nil {
		c.Ui.Error(fmt.Sprintf("Error checking Terraform Provider documentation: %s", err))
		return 1
	}

	return 0
}

func (c *CheckCommand) Synopsis() string {
	return "Checks Terraform Provider documentation"
}

func allowedSubcategoriesFile(path string) ([]string, error) {
	log.Printf("[DEBUG] Loading allowed subcategories file: %s", path)

	file, err := os.Open(path)

	if err != nil {
		return nil, fmt.Errorf("error opening allowed subcategories file (%s): %w", path, err)
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	var allowedSubcategories []string

	for scanner.Scan() {
		allowedSubcategories = append(allowedSubcategories, scanner.Text())
	}

	if err != nil {
		return nil, fmt.Errorf("error reading allowed subcategories file (%s): %w", path, err)
	}

	return allowedSubcategories, nil
}

func providerNameFromCurrentDirectory() string {
	path, _ := os.Getwd()

	return providerNameFromPath(path)
}

func providerNameFromPath(path string) string {
	base := filepath.Base(path)

	if strings.ContainsAny(base, "./") {
		return ""
	}

	if !strings.HasPrefix(base, "terraform-provider-") {
		return ""
	}

	return strings.TrimPrefix(base, "terraform-provider-")
}

// providerSchemas reads, parses, and validates a provided terraform provider schema -json path.
func providerSchemas(path string) (*tfjson.ProviderSchemas, error) {
	log.Printf("[DEBUG] Loading providers schema JSON file: %s", path)

	content, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("error reading providers schema JSON file (%s): %w", path, err)
	}

	var ps tfjson.ProviderSchemas

	if err := json.Unmarshal(content, &ps); err != nil {
		return nil, fmt.Errorf("error parsing providers schema JSON file (%s): %w", path, err)
	}

	if err := ps.Validate(); err != nil {
		return nil, fmt.Errorf("error validating providers schema JSON file (%s): %w", path, err)
	}

	return &ps, nil
}

// providerSchemasDataSources returns all data sources from a terraform providers schema -json provider.
func providerSchemasDataSources(ps *tfjson.ProviderSchemas, providerName string) map[string]*tfjson.Schema {
	if ps == nil || providerName == "" {
		return nil
	}

	provider, ok := ps.Schemas[providerName]

	if !ok {
		log.Printf("[WARN] Provider name (%s) not found in provider schema", providerName)
		return nil
	}

	dataSources := make([]string, 0, len(provider.DataSourceSchemas))

	for name := range provider.DataSourceSchemas {
		dataSources = append(dataSources, name)
	}

	sort.Strings(dataSources)

	log.Printf("[DEBUG] Found provider schema data sources: %v", dataSources)

	return provider.DataSourceSchemas
}

// providerSchemasResources returns all resources from a terraform providers schema -json provider.
func providerSchemasResources(ps *tfjson.ProviderSchemas, providerName string) map[string]*tfjson.Schema {
	if ps == nil || providerName == "" {
		return nil
	}

	provider, ok := ps.Schemas[providerName]

	if !ok {
		log.Printf("[WARN] Provider name (%s) not found in provider schema", providerName)
		return nil
	}

	resources := make([]string, 0, len(provider.ResourceSchemas))

	for name := range provider.ResourceSchemas {
		resources = append(resources, name)
	}

	sort.Strings(resources)

	log.Printf("[DEBUG] Found provider schema data sources: %v", resources)

	return provider.ResourceSchemas
}
