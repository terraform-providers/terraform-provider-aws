//go:generate go run ../../../generators/listpages/main.go -function=DescribeCapacityProviders -paginator=NextToken github.com/aws/aws-sdk-go/service/ecs

package lister
