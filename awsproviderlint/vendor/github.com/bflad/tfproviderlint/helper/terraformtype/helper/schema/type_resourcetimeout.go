package schema

import "time"

const (
	ResourceTimeoutTypeCreateField  = `Create`
	ResourceTimeoutTypeDefaultField = `Default`
	ResourceTimeoutTypeDeleteField  = `Delete`
	ResourceTimeoutTypeReadField    = `Read`
	ResourceTimeoutTypeUpdateField  = `Update`
)

// resourceTimeoutType is an internal representation of the SDK helper/schema.ResourceTimeout type
type resourceTimeoutType struct {
	Create  *time.Duration
	Default *time.Duration
	Delete  *time.Duration
	Read    *time.Duration
	Update  *time.Duration
}
