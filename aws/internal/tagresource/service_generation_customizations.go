package tagresource

import (
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

// ServiceCheckDestroyIgnoreError determines additional CheckDestroy error handling.
//
// Use this to ignore errors returned by the list tags API for missing parent resources.
// For generator simplicity, AWS Go SDK service specific constants are not available.
//
// This handling should be in the form of:
// if CONDITIONAL {
//   continue
// }
func ServiceCheckDestroyIgnoreError(serviceName string) string {
	switch serviceName {
	case "dynamodb":
		return `
if isAWSErr(err, "ResourceNotFoundException", "") {
	continue
}
`
	case "ecs":
		return `
if isAWSErr(err, "InvalidParameterException", "The specified cluster is inactive. Specify an active cluster and try again.") {
	continue
}
`
	default:
		return ""
	}
}

// ServiceIdentifierAttributeName determines the schema identifier attribute name.
func ServiceIdentifierAttributeName(serviceName string) string {
	switch serviceName {
	case "ec2":
		return "resource_id"
	default:
		return toSnakeCase(keyvaluetags.ServiceTagInputIdentifierField(serviceName))
	}
}
