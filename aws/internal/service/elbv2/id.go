package elbv2

import (
	"fmt"
	"strings"
)

const listenerCertificateIDSeparator = "_"

func ListenerCertificateParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, listenerCertificateIDSeparator, 2)
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "",
		fmt.Errorf("unexpected format for ID (%q), expected listener-arn"+listenerCertificateIDSeparator+
			"certificate-arn", id)
}

func ListenerCertificateCreateID(listenerArn, certificateArn string) string {
	return strings.Join([]string{listenerArn, listenerCertificateIDSeparator, certificateArn}, "")
}
