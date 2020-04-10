package check

type SideNavigationOptions struct {
	*FileOptions

	IgnoreDataSources []string
	IgnoreResources   []string

	ProviderName string
}

func (opts *SideNavigationOptions) ShouldIgnoreDataSource(name string) bool {
	for _, ignoreDataSource := range opts.IgnoreDataSources {
		if ignoreDataSource == name {
			return true
		}
	}

	return false
}

func (opts *SideNavigationOptions) ShouldIgnoreResource(name string) bool {
	for _, ignoreResource := range opts.IgnoreResources {
		if ignoreResource == name {
			return true
		}
	}

	return false
}
