package check

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/hashicorp/go-multierror"
)

type RegistryDataSourceFileOptions struct {
	*FileOptions

	FrontMatter *FrontMatterOptions
}

type RegistryDataSourceFileCheck struct {
	FileCheck

	Options *RegistryDataSourceFileOptions
}

func NewRegistryDataSourceFileCheck(opts *RegistryDataSourceFileOptions) *RegistryDataSourceFileCheck {
	check := &RegistryDataSourceFileCheck{
		Options: opts,
	}

	if check.Options == nil {
		check.Options = &RegistryDataSourceFileOptions{}
	}

	if check.Options.FileOptions == nil {
		check.Options.FileOptions = &FileOptions{}
	}

	if check.Options.FrontMatter == nil {
		check.Options.FrontMatter = &FrontMatterOptions{}
	}

	check.Options.FrontMatter.NoLayout = true
	check.Options.FrontMatter.NoSidebarCurrent = true

	return check
}

func (check *RegistryDataSourceFileCheck) Run(path string) error {
	fullpath := check.Options.FullPath(path)

	log.Printf("[DEBUG] Checking file: %s", fullpath)

	if err := RegistryFileExtensionCheck(path); err != nil {
		return fmt.Errorf("%s: error checking file extension: %w", path, err)
	}

	if err := FileSizeCheck(fullpath); err != nil {
		return fmt.Errorf("%s: error checking file size: %w", path, err)
	}

	content, err := ioutil.ReadFile(fullpath)

	if err != nil {
		return fmt.Errorf("%s: error reading file: %w", path, err)
	}

	if err := NewFrontMatterCheck(check.Options.FrontMatter).Run(content); err != nil {
		return fmt.Errorf("%s: error checking file frontmatter: %w", path, err)
	}

	return nil
}

func (check *RegistryDataSourceFileCheck) RunAll(files []string) error {
	var result *multierror.Error

	for _, file := range files {
		if err := check.Run(file); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}
