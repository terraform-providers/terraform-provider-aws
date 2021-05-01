package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

// CertificateAuthorityCertificateByARN returns the certificate for the certificate authority corresponding to the specified ARN.
// Returns a resource.NotFoundError if no certificate authority is found or the certificate authority does not have a certificate assigned.
func CertificateAuthorityCertificateByARN(conn *acmpca.ACMPCA, arn string) (*acmpca.GetCertificateAuthorityCertificateOutput, error) {
	input := &acmpca.GetCertificateAuthorityCertificateInput{
		CertificateAuthorityArn: aws.String(arn),
	}

	output, err := conn.GetCertificateAuthorityCertificate(input)
	if tfawserr.ErrCodeEquals(err, acmpca.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &resource.NotFoundError{
			Message:     "empty result",
			LastRequest: input,
		}
	}

	return output, nil
}
