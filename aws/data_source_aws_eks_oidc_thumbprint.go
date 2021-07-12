package aws

import (
	"crypto/sha1"
	"crypto/tls"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsEksOIDCThumbprint() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEksOIDCThumbprintRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"sha1_hash": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsEksOIDCThumbprintRead(d *schema.ResourceData, meta interface{}) error {
	d.SetId(resource.UniqueId())
	s, err := oidcEksSha1FingerprintFor(d.Get("region").(string))
	if err != nil {
		return fmt.Errorf("could not get thumbprint: %v", err)
	}
	if err := d.Set("sha1_hash", s); err != nil {
		return fmt.Errorf("error setting certificate sha1_hash: %v", err)
	}
	return nil
}

func oidcEksSha1FingerprintFor(region string) (string, error) {
	endpoint := fmt.Sprintf("oidc.eks.%s.amazonaws.com:443", region)
	conn, err := tls.Dial("tcp", endpoint, &tls.Config{})
	if err != nil {
		return "", fmt.Errorf("connection error trying to reach %s: %v", endpoint, err)
	}
	defer conn.Close()
	certChain := conn.ConnectionState().PeerCertificates
	// Read the x509.Certificate raw content (DER encoded) and hash it with SHA1,
	// we're only interested in the CA certificate, the last one
	fingerprint := sha1.Sum(certChain[len(certChain)-1].Raw)
	return fmt.Sprintf("%x", fingerprint), nil
}
