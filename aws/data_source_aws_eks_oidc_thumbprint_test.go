package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	eksThumbprintDataSourceTestRegion = "eu-central-1"
	// CA Thumbprint for oidc.eks.eu-central-1.amazonaws.com
	// in the format expected by the IAM OIDC provider resource configuration
	expectedSHA1Thumbprint = "9e99a48a9960b14926bb7f3b02e22da2b0ab7280"
	//expectedSHA1Thumbprints = []string{"9e99a48a9960b14926bb7f3b02e22da2b0ab7280","02f8cdf9a7cc82efafcf86b5a626a3242bfe978d"}
)

func TestAccEksOIDCThumbprint_readCertificateForRegion(t *testing.T) {
	var providers []*schema.Provider
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEksThumbprintDataSourceConfig(eksThumbprintDataSourceTestRegion),
				Check:  resource.TestCheckResourceAttr("data.aws_eks_oidc_thumbprint.frankfurt", "sha1_hash", expectedSHA1Thumbprint),
			},
		},
	})
}

func TestAccEksOIDCThumbprint_failsToReadCertificateForNonExistingRegion(t *testing.T) {
	wrongRegion := "bogusregion"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckAwsEksThumbprintDataSourceConfig(wrongRegion),
				Check:       resource.TestCheckNoResourceAttr("data.aws_eks_oidc_thumbprint.frankfurt", "sha1_hash"),
				ExpectError: regexp.MustCompile(`.*could not get thumbprint: connection error trying to reach.*`),
			},
		},
	})
}

func testAccCheckAwsEksThumbprintDataSourceConfig(region string) string {
	return fmt.Sprintf(`
data "aws_eks_oidc_thumbprint" "frankfurt" {
  region = "%s"
}
`, region)
}
