// Code generated by generators/servicefilters/main.go; DO NOT EDIT.

package namevaluesfilters

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/rds"
)

// []*SERVICE.Filter handling

// Ec2Filters returns ec2 service filters.
func (filters NameValuesFilters) Ec2Filters() []*ec2.Filter {
	result := make([]*ec2.Filter, 0, len(filters))

	for k, v := range filters.Map() {
		filter := &ec2.Filter{
			Name:   aws.String(k),
			Values: aws.StringSlice(v),
		}

		result = append(result, filter)
	}

	return result
}

// RdsFilters returns rds service filters.
func (filters NameValuesFilters) RdsFilters() []*rds.Filter {
	result := make([]*rds.Filter, 0, len(filters))

	for k, v := range filters.Map() {
		filter := &rds.Filter{
			Name:   aws.String(k),
			Values: aws.StringSlice(v),
		}

		result = append(result, filter)
	}

	return result
}
