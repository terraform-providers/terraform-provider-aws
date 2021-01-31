package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for a QueryLogConfigAssociation to return ACTIVE
	QueryLogConfigAssociationCreatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a QueryLogConfigAssociation to be deleted
	QueryLogConfigAssociationDeletedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a QueryLogConfig to return CREATED
	QueryLogConfigCreatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a QueryLogConfig to be deleted
	QueryLogConfigDeletedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a DnssecConfig to return ENABLED
	DnssecConfigCreatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a DnssecConfig to return DISABLED
	DnssecConfigDeletedTimeout = 5 * time.Minute
)

// QueryLogConfigAssociationCreated waits for a QueryLogConfig to return ACTIVE
func QueryLogConfigAssociationCreated(conn *route53resolver.Route53Resolver, queryLogConfigAssociationID string) (*route53resolver.ResolverQueryLogConfigAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverQueryLogConfigAssociationStatusCreating},
		Target:  []string{route53resolver.ResolverQueryLogConfigAssociationStatusActive},
		Refresh: QueryLogConfigAssociationStatus(conn, queryLogConfigAssociationID),
		Timeout: QueryLogConfigAssociationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverQueryLogConfigAssociation); ok {
		return v, err
	}

	return nil, err
}

// QueryLogConfigAssociationCreated waits for a QueryLogConfig to be deleted
func QueryLogConfigAssociationDeleted(conn *route53resolver.Route53Resolver, queryLogConfigAssociationID string) (*route53resolver.ResolverQueryLogConfigAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverQueryLogConfigAssociationStatusDeleting},
		Target:  []string{},
		Refresh: QueryLogConfigAssociationStatus(conn, queryLogConfigAssociationID),
		Timeout: QueryLogConfigAssociationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverQueryLogConfigAssociation); ok {
		return v, err
	}

	return nil, err
}

// QueryLogConfigCreated waits for a QueryLogConfig to return CREATED
func QueryLogConfigCreated(conn *route53resolver.Route53Resolver, queryLogConfigID string) (*route53resolver.ResolverQueryLogConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverQueryLogConfigStatusCreating},
		Target:  []string{route53resolver.ResolverQueryLogConfigStatusCreated},
		Refresh: QueryLogConfigStatus(conn, queryLogConfigID),
		Timeout: QueryLogConfigCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverQueryLogConfig); ok {
		return v, err
	}

	return nil, err
}

// QueryLogConfigCreated waits for a QueryLogConfig to be deleted
func QueryLogConfigDeleted(conn *route53resolver.Route53Resolver, queryLogConfigID string) (*route53resolver.ResolverQueryLogConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverQueryLogConfigStatusDeleting},
		Target:  []string{},
		Refresh: QueryLogConfigStatus(conn, queryLogConfigID),
		Timeout: QueryLogConfigDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverQueryLogConfig); ok {
		return v, err
	}

	return nil, err
}

// DnssecConfigCreated waits for a DnssecConfig to return ENABLED
func DnssecConfigCreated(conn *route53resolver.Route53Resolver, dnssecConfigID string) (*route53resolver.ResolverDnssecConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverDNSSECValidationStatusEnabling},
		Target:  []string{route53resolver.ResolverDNSSECValidationStatusEnabled},
		Refresh: DnssecConfigStatus(conn, dnssecConfigID),
		Timeout: DnssecConfigCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverDnssecConfig); ok {
		return v, err
	}

	return nil, err
}

// DnssecConfigCreated waits for a DnssecConfig to return DELETED
func DnssecConfigDeleted(conn *route53resolver.Route53Resolver, dnssecConfigID string) (*route53resolver.ResolverDnssecConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverDNSSECValidationStatusDisabling},
		Target:  []string{route53resolver.ResolverDNSSECValidationStatusDisabled},
		Refresh: DnssecConfigStatus(conn, dnssecConfigID),
		Timeout: DnssecConfigDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*route53resolver.ResolverDnssecConfig); ok {
		return v, err
	}

	return nil, err
}
