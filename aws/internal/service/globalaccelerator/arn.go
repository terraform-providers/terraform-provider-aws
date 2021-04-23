package globalaccelerator

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
)

const (
	ARNSeparator = "/"
	ARNService   = "globalaccelerator"
)

// EndpointGroupARNToListenerARN converts an endpoint group ARN to a listener ARN.
// See https://docs.aws.amazon.com/service-authorization/latest/reference/list_awsglobalaccelerator.html#awsglobalaccelerator-resources-for-iam-policies.
func EndpointGroupARNToListenerARN(inputARN string) (string, error) {
	parsedARN, err := arn.Parse(inputARN)

	if err != nil {
		return "", fmt.Errorf("error parsing ARN (%s): %w", inputARN, err)
	}

	if actual, expected := parsedARN.Service, ARNService; actual != expected {
		return "", fmt.Errorf("expected service %s in ARN (%s), got: %s", expected, inputARN, actual)
	}

	resourceParts := strings.Split(parsedARN.Resource, ARNSeparator)

	if actual, expected := len(resourceParts), 6; actual < expected {
		return "", fmt.Errorf("expected at least %d resource parts in ARN (%s), got: %d", expected, inputARN, actual)
	}

	outputARN := arn.ARN{
		Partition: parsedARN.Partition,
		Service:   parsedARN.Service,
		Region:    parsedARN.Region,
		AccountID: parsedARN.AccountID,
		Resource:  strings.Join(resourceParts[0:4], ARNSeparator),
	}.String()

	return outputARN, nil
}

// ListenerOrEndpointGroupARNToAcceleratorARN converts a listener or endpoint group ARN to an accelerator ARN.
// See https://docs.aws.amazon.com/service-authorization/latest/reference/list_awsglobalaccelerator.html#awsglobalaccelerator-resources-for-iam-policies.
func ListenerOrEndpointGroupARNToAcceleratorARN(inputARN string) (string, error) {
	parsedARN, err := arn.Parse(inputARN)

	if err != nil {
		return "", fmt.Errorf("error parsing ARN (%s): %w", inputARN, err)
	}

	if actual, expected := parsedARN.Service, ARNService; actual != expected {
		return "", fmt.Errorf("expected service %s in ARN (%s), got: %s", expected, inputARN, actual)
	}

	resourceParts := strings.Split(parsedARN.Resource, ARNSeparator)

	if actual, expected := len(resourceParts), 4; actual < expected {
		return "", fmt.Errorf("expected at least %d resource parts in ARN (%s), got: %d", expected, inputARN, actual)
	}

	outputARN := arn.ARN{
		Partition: parsedARN.Partition,
		Service:   parsedARN.Service,
		Region:    parsedARN.Region,
		AccountID: parsedARN.AccountID,
		Resource:  strings.Join(resourceParts[0:2], ARNSeparator),
	}.String()

	return outputARN, nil
}
