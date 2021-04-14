package finder

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	tfroute53 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53"
)

func HostedZoneDnssec(conn *route53.Route53, hostedZoneID string) (*route53.GetDNSSECOutput, error) {
	input := &route53.GetDNSSECInput{
		HostedZoneId: aws.String(hostedZoneID),
	}

	output, err := conn.GetDNSSEC(input)

	if err != nil {
		return nil, err
	}

	return output, nil
}

func KeySigningKey(conn *route53.Route53, hostedZoneID string, name string) (*route53.KeySigningKey, error) {
	input := &route53.GetDNSSECInput{
		HostedZoneId: aws.String(hostedZoneID),
	}

	var result *route53.KeySigningKey

	output, err := conn.GetDNSSEC(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, keySigningKey := range output.KeySigningKeys {
		if keySigningKey == nil {
			continue
		}

		if aws.StringValue(keySigningKey.Name) == name {
			result = keySigningKey
			break
		}
	}

	return result, err
}

func KeySigningKeyByResourceID(conn *route53.Route53, resourceID string) (*route53.KeySigningKey, error) {
	hostedZoneID, name, err := tfroute53.KeySigningKeyParseResourceID(resourceID)

	if err != nil {
		return nil, fmt.Errorf("error parsing Route 53 Key Signing Key (%s) identifier: %w", resourceID, err)
	}

	return KeySigningKey(conn, hostedZoneID, name)
}
