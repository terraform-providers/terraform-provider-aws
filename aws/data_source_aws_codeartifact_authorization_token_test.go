package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSCodeArtifactAuthorizationTokenDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_codeartifact_authorization_token.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codeartifact.EndpointsID, t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCodeArtifactAuthorizationTokenBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "authorization_token"),
					resource.TestCheckResourceAttrSet(dataSourceName, "expiration"),
					testAccCheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
		},
	})
}

func TestAccAWSCodeArtifactAuthorizationTokenDataSource_owner(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_codeartifact_authorization_token.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codeartifact.EndpointsID, t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCodeArtifactAuthorizationTokenOwnerConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "authorization_token"),
					resource.TestCheckResourceAttrSet(dataSourceName, "expiration"),
					testAccCheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
		},
	})
}

func TestAccAWSCodeArtifactAuthorizationTokenDataSource_duration(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_codeartifact_authorization_token.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codeartifact.EndpointsID, t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCodeArtifactAuthorizationTokenDurationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "authorization_token"),
					resource.TestCheckResourceAttrSet(dataSourceName, "expiration"),
					resource.TestCheckResourceAttr(dataSourceName, "duration_seconds", "900"),
					testAccCheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
		},
	})
}

func testAccCheckAWSCodeArtifactAuthorizationTokenBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}
`, rName)
}

func testAccCheckAWSCodeArtifactAuthorizationTokenBasicConfig(rName string) string {
	return composeConfig(
		testAccCheckAWSCodeArtifactAuthorizationTokenBaseConfig(rName),
		`
data "aws_codeartifact_authorization_token" "test" {
  domain = aws_codeartifact_domain.test.domain
}
`)
}

func testAccCheckAWSCodeArtifactAuthorizationTokenOwnerConfig(rName string) string {
	return composeConfig(
		testAccCheckAWSCodeArtifactAuthorizationTokenBaseConfig(rName),
		`
data "aws_codeartifact_authorization_token" "test" {
  domain       = aws_codeartifact_domain.test.domain
  domain_owner = aws_codeartifact_domain.test.owner
}
`)
}

func testAccCheckAWSCodeArtifactAuthorizationTokenDurationConfig(rName string) string {
	return composeConfig(
		testAccCheckAWSCodeArtifactAuthorizationTokenBaseConfig(rName),
		`
data "aws_codeartifact_authorization_token" "test" {
  domain           = aws_codeartifact_domain.test.domain
  duration_seconds = 900
}
`)
}
