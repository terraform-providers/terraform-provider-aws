package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
)

// CertificateAuthorityByARN returns the certificate authority corresponding to the specified ARN.
// Returns nil if no certificate authority is found.
func CertificateAuthorityByARN(conn *acmpca.ACMPCA, arn string) (*acmpca.CertificateAuthority, error) {
	input := &acmpca.DescribeCertificateAuthorityInput{
		CertificateAuthorityArn: aws.String(arn),
	}

	output, err := conn.DescribeCertificateAuthority(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.CertificateAuthority, nil
}
