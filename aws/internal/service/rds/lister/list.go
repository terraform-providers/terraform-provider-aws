//go:generate go run ../../../generators/listpages/main.go -function=DescribeDBClusterParameterGroups -paginator Marker github.com/aws/aws-sdk-go/service/rds

package lister

import (
	"github.com/aws/aws-sdk-go/service/rds"
)

func ListAllClusterParameterGroupsPages(conn *rds.RDS, fn func(*rds.DescribeDBClusterParameterGroupsOutput, bool) bool) error {
	input := &rds.DescribeDBClusterParameterGroupsInput{}

	return DescribeDBClusterParameterGroupsPages(conn, input, fn)
}
