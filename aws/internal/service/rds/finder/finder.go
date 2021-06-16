package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	tfrds "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/rds"
)

// DBProxyTarget returns matching DBProxyTarget.
func DBProxyTarget(conn *rds.RDS, dbProxyName, targetGroupName, targetType, rdsResourceId string) (*rds.DBProxyTarget, error) {
	input := &rds.DescribeDBProxyTargetsInput{
		DBProxyName:     aws.String(dbProxyName),
		TargetGroupName: aws.String(targetGroupName),
	}
	var dbProxyTarget *rds.DBProxyTarget

	err := conn.DescribeDBProxyTargetsPages(input, func(page *rds.DescribeDBProxyTargetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, target := range page.Targets {
			if aws.StringValue(target.Type) == targetType && aws.StringValue(target.RdsResourceId) == rdsResourceId {
				dbProxyTarget = target
				return false
			}
		}

		return !lastPage
	})

	return dbProxyTarget, err
}

// DBProxyEndpoint returns matching DBProxyEndpoint.
func DBProxyEndpoint(conn *rds.RDS, id string) (*rds.DBProxyEndpoint, error) {
	dbProxyName, dbProxyEndpointName, err := tfrds.ResourceAwsDbProxyEndpointParseID(id)
	if err != nil {
		return nil, err
	}

	input := &rds.DescribeDBProxyEndpointsInput{
		DBProxyName:         aws.String(dbProxyName),
		DBProxyEndpointName: aws.String(dbProxyEndpointName),
	}
	var dbProxyEndpoint *rds.DBProxyEndpoint

	err = conn.DescribeDBProxyEndpointsPages(input, func(page *rds.DescribeDBProxyEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, endpoint := range page.DBProxyEndpoints {
			if aws.StringValue(endpoint.DBProxyEndpointName) == dbProxyEndpointName &&
				aws.StringValue(endpoint.DBProxyName) == dbProxyName {
				dbProxyEndpoint = endpoint
				return false
			}
		}

		return !lastPage
	})

	return dbProxyEndpoint, err
}
