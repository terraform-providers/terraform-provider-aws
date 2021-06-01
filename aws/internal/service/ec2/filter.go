package ec2

import (
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// BuildAttributeFilterList takes a flat map of scalar attributes (most
// likely values extracted from a *schema.ResourceData on an EC2-querying
// data source) and produces a []*ec2.Filter representing an exact match
// for each of the given non-empty attributes.
//
// The keys of the given attributes map are the attribute names expected
// by the EC2 API, which are usually either in camelcase or with dash-separated
// words. We conventionally map these to underscore-separated identifiers
// with the same words when presenting these as data source query attributes
// in Terraform.
//
// It's the callers responsibility to transform any non-string values into
// the appropriate string serialization required by the AWS API when
// encoding the given filter. Any attributes given with empty string values
// are ignored, assuming that the user wishes to leave that attribute
// unconstrained while filtering.
//
// The purpose of this function is to create values to pass in
// for the "Filters" attribute on most of the "Describe..." API functions in
// the EC2 API, to aid in the implementation of Terraform data sources that
// retrieve data about EC2 objects.
func BuildAttributeFilterList(attrs map[string]string) []*ec2.Filter {
	var filters []*ec2.Filter

	// sort the filters by name to make the output deterministic
	var names []string
	for filterName := range attrs {
		names = append(names, filterName)
	}

	sort.Strings(names)

	for _, filterName := range names {
		value := attrs[filterName]
		if value == "" {
			continue
		}

		filters = append(filters, &ec2.Filter{
			Name:   aws.String(filterName),
			Values: []*string{aws.String(value)},
		})
	}

	return filters
}
