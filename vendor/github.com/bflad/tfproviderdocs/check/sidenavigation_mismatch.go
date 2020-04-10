package check

import (
	"fmt"

	"github.com/bflad/tfproviderdocs/check/sidenavigation"
	"github.com/hashicorp/go-multierror"
)

func SideNavigationMismatchCheck(opts *SideNavigationOptions, dataSourceFiles []string, resourceFiles []string) error {
	if opts == nil || opts.ProviderName == "" {
		return nil
	}

	path := fmt.Sprintf("website/%s%s", opts.ProviderName, FileExtensionErb)

	sideNavigation, err := sidenavigation.FindFile(opts.FullPath(path))

	if err != nil {
		return fmt.Errorf("%s: error finding side navigation: %s", path, err)
	}

	if sideNavigation == nil {
		return fmt.Errorf("%s: error finding side navigation: not found in file", path)
	}

	var missingDataSources, missingResources []string
	var result *multierror.Error

	for _, dataSourceFile := range dataSourceFiles {
		dataSourceName := fileResourceName(opts.ProviderName, dataSourceFile)

		if sideNavigation.HasDataSourceLink(dataSourceName) {
			continue
		}

		missingDataSources = append(missingDataSources, dataSourceName)
	}

	for _, resourceFile := range resourceFiles {
		resourceName := fileResourceName(opts.ProviderName, resourceFile)

		if sideNavigation.HasResourceLink(resourceName) {
			continue
		}

		missingResources = append(missingResources, resourceName)
	}

	for _, missingDataSource := range missingDataSources {
		if opts.ShouldIgnoreDataSource(missingDataSource) {
			continue
		}

		err := fmt.Errorf("%s: missing side navigation link for data source: %s", path, missingDataSource)
		result = multierror.Append(result, err)
	}

	for _, missingResource := range missingResources {
		if opts.ShouldIgnoreResource(missingResource) {
			continue
		}

		err := fmt.Errorf("%s: missing side navigation link for resource: %s", path, missingResource)
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}
