// This file contains code generation customizations for each AWS Go SDK service.

package namevaluesfilters

// ServiceFilterPackage determines the service filter type package.
func ServiceFilterPackage(serviceName string) string {
	switch serviceName {
	default:
		return serviceName
	}
}

// ServiceFilterType determines the service filter type.
func ServiceFilterType(serviceName string) string {
	switch serviceName {
	case "resourcegroupstaggingapi":
		return "TagFilter"
	default:
		return "Filter"
	}
}

// ServiceFilterTypeNameField determines the service filter type name field.
func ServiceFilterTypeNameField(serviceName string) string {
	switch serviceName {
	case "resourcegroupstaggingapi":
		return "Key"
	default:
		return "Name"
	}
}

// ServiceFilterTypeValuesField determines the service filter type values field.
func ServiceFilterTypeValuesField(serviceName string) string {
	switch serviceName {
	default:
		return "Values"
	}
}
