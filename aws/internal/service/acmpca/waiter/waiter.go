package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// CertificateAuthorityCreated waits for a CertificateAuthority to return Active or PendingCertificate
func CertificateAuthorityCreated(conn *acmpca.ACMPCA, arn string, timeout time.Duration) (*acmpca.CertificateAuthority, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"", acmpca.CertificateAuthorityStatusCreating},
		Target:  []string{acmpca.CertificateAuthorityStatusActive, acmpca.CertificateAuthorityStatusPendingCertificate},
		Refresh: CertificateAuthorityStatus(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*acmpca.CertificateAuthority); ok {
		return v, err
	}

	return nil, err
}
