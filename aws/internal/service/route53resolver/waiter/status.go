package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53resolver/finder"
)

const (
	resolverQueryLogConfigAssociationStatusNotFound = "NotFound"
	resolverQueryLogConfigAssociationStatusUnknown  = "Unknown"

	resolverQueryLogConfigStatusNotFound = "NotFound"
	resolverQueryLogConfigStatusUnknown  = "Unknown"

	resolverDnssecConfigStatusNotFound = "NotFound"
	resolverDnssecConfigStatusUnknown  = "Unknown"
)

// QueryLogConfigAssociationStatus fetches the QueryLogConfigAssociation and its Status
func QueryLogConfigAssociationStatus(conn *route53resolver.Route53Resolver, queryLogConfigAssociationID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		queryLogConfigAssociation, err := finder.ResolverQueryLogConfigAssociationByID(conn, queryLogConfigAssociationID)

		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			return nil, resolverQueryLogConfigAssociationStatusNotFound, nil
		}

		if err != nil {
			return nil, resolverQueryLogConfigAssociationStatusUnknown, err
		}

		if queryLogConfigAssociation == nil {
			return nil, resolverQueryLogConfigAssociationStatusNotFound, nil
		}

		return queryLogConfigAssociation, aws.StringValue(queryLogConfigAssociation.Status), nil
	}
}

// QueryLogConfigStatus fetches the QueryLogConfig and its Status
func QueryLogConfigStatus(conn *route53resolver.Route53Resolver, queryLogConfigID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		queryLogConfig, err := finder.ResolverQueryLogConfigByID(conn, queryLogConfigID)

		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			return nil, resolverQueryLogConfigStatusNotFound, nil
		}

		if err != nil {
			return nil, resolverQueryLogConfigStatusUnknown, err
		}

		if queryLogConfig == nil {
			return nil, resolverQueryLogConfigStatusNotFound, nil
		}

		return queryLogConfig, aws.StringValue(queryLogConfig.Status), nil
	}
}

// DnssecConfigStatus fetches the DnssecConfig and its Status
func DnssecConfigStatus(conn *route53resolver.Route53Resolver, dnssecConfigID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		dnssecConfig, err := finder.ResolverDnssecConfigByID(conn, dnssecConfigID)

		if err != nil {
			return nil, resolverDnssecConfigStatusUnknown, err
		}

		if dnssecConfig == nil {
			return nil, resolverDnssecConfigStatusNotFound, nil
		}

		return dnssecConfig, aws.StringValue(dnssecConfig.ValidationStatus), nil
	}
}
