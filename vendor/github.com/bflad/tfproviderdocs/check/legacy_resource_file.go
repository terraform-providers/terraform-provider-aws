package check

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/hashicorp/go-multierror"
)

type LegacyResourceFileOptions struct {
	*FileOptions

	FrontMatter *FrontMatterOptions
}

type LegacyResourceFileCheck struct {
	FileCheck

	Options *LegacyResourceFileOptions
}

func NewLegacyResourceFileCheck(opts *LegacyResourceFileOptions) *LegacyResourceFileCheck {
	check := &LegacyResourceFileCheck{
		Options: opts,
	}

	if check.Options == nil {
		check.Options = &LegacyResourceFileOptions{}
	}

	if check.Options.FileOptions == nil {
		check.Options.FileOptions = &FileOptions{}
	}

	if check.Options.FrontMatter == nil {
		check.Options.FrontMatter = &FrontMatterOptions{}
	}

	check.Options.FrontMatter.NoSidebarCurrent = true
	check.Options.FrontMatter.RequireDescription = true
	check.Options.FrontMatter.RequireLayout = true
	check.Options.FrontMatter.RequirePageTitle = true

	return check
}

func (check *LegacyResourceFileCheck) Run(path string) error {
	fullpath := check.Options.FullPath(path)

	log.Printf("[DEBUG] Checking file: %s", fullpath)

	if err := LegacyFileExtensionCheck(path); err != nil {
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

func (check *LegacyResourceFileCheck) RunAll(files []string) error {
	var result *multierror.Error

	for _, file := range files {
		if err := check.Run(file); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}
