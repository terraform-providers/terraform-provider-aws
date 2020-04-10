package check

import (
	"fmt"

	"github.com/bflad/tfproviderdocs/check/sidenavigation"
	"github.com/hashicorp/go-multierror"
)

func SideNavigationLinkCheck(opts *SideNavigationOptions) error {
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

	var errors *multierror.Error

	for _, link := range sideNavigation.DataSourceLinks {
		if opts.ShouldIgnoreDataSource(link.Text) {
			continue
		}

		if err := link.Validate(opts.ProviderName); err != nil {
			errors = multierror.Append(errors, fmt.Errorf("%s: error validating link: %s", path, err))
		}
	}

	for _, link := range sideNavigation.ResourceLinks {
		if opts.ShouldIgnoreResource(link.Text) {
			continue
		}

		if err := link.Validate(opts.ProviderName); err != nil {
			errors = multierror.Append(errors, fmt.Errorf("%s: error validating link: %s", path, err))
		}
	}

	return errors.ErrorOrNil()
}
