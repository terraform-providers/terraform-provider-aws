package sqs

import (
	"fmt"
	"net/url"
	"strings"
)

// QueueNameFromURL returns the SQS queue name from the specified URL.
func QueueNameFromURL(u string) (string, error) {
	v, err := url.Parse(u)

	if err != nil {
		return "", err
	}

	// http://sqs.us-west-2.amazonaws.com/123456789012/queueName
	parts := strings.Split(v.Path, "/")

	if len(parts) != 3 {
		return "", fmt.Errorf("SQS Queue URL (%s) is in the incorrect format", u)
	}

	return parts[2], nil
}
