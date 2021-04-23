package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
)

// ResolverQueryLogConfigAssociationByID returns the query logging configuration association corresponding to the specified ID.
// Returns nil if no configuration is found.
func ResolverQueryLogConfigAssociationByID(conn *route53resolver.Route53Resolver, queryLogConfigAssociationID string) (*route53resolver.ResolverQueryLogConfigAssociation, error) {
	input := &route53resolver.GetResolverQueryLogConfigAssociationInput{
		ResolverQueryLogConfigAssociationId: aws.String(queryLogConfigAssociationID),
	}

	output, err := conn.GetResolverQueryLogConfigAssociation(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.ResolverQueryLogConfigAssociation, nil
}

// ResolverQueryLogConfigByID returns the query logging configuration corresponding to the specified ID.
// Returns nil if no configuration is found.
func ResolverQueryLogConfigByID(conn *route53resolver.Route53Resolver, queryLogConfigID string) (*route53resolver.ResolverQueryLogConfig, error) {
	input := &route53resolver.GetResolverQueryLogConfigInput{
		ResolverQueryLogConfigId: aws.String(queryLogConfigID),
	}

	output, err := conn.GetResolverQueryLogConfig(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.ResolverQueryLogConfig, nil
}

// ResolverDnssecConfigByID returns the dnssec configuration corresponding to the specified ID.
// Returns nil if no configuration is found.
func ResolverDnssecConfigByID(conn *route53resolver.Route53Resolver, dnssecConfigID string) (*route53resolver.ResolverDnssecConfig, error) {
	input := &route53resolver.ListResolverDnssecConfigsInput{}

	var config *route53resolver.ResolverDnssecConfig
	// GetResolverDnssecConfigs does not support query with id
	err := conn.ListResolverDnssecConfigsPages(input, func(page *route53resolver.ListResolverDnssecConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, c := range page.ResolverDnssecConfigs {
			if aws.StringValue(c.Id) == dnssecConfigID {
				config = c
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if config == nil {
		return nil, nil
	}

	return config, nil
}
