package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSCodeArtifactRepositoryEndpointDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_codeartifact_repository_endpoint.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codeartifact.EndpointsID, t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCodeArtifactRepositoryEndpointBasicConfig(rName, "npm"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					testAccCheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
			{
				Config: testAccCheckAWSCodeArtifactRepositoryEndpointBasicConfig(rName, "pypi"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					testAccCheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
			{
				Config: testAccCheckAWSCodeArtifactRepositoryEndpointBasicConfig(rName, "maven"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					testAccCheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
			{
				Config: testAccCheckAWSCodeArtifactRepositoryEndpointBasicConfig(rName, "nuget"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					testAccCheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
		},
	})
}

func TestAccAWSCodeArtifactRepositoryEndpointDataSource_owner(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_codeartifact_repository_endpoint.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codeartifact.EndpointsID, t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCodeArtifactRepositoryEndpointOwnerConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "repository_endpoint"),
					testAccCheckResourceAttrAccountID(dataSourceName, "domain_owner"),
				),
			},
		},
	})
}

func testAccCheckAWSCodeArtifactRepositoryEndpointBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}

resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain
}
`, rName)
}

func testAccCheckAWSCodeArtifactRepositoryEndpointBasicConfig(rName, format string) string {
	return composeConfig(
		testAccCheckAWSCodeArtifactRepositoryEndpointBaseConfig(rName),
		fmt.Sprintf(`
data "aws_codeartifact_repository_endpoint" "test" {
  domain     = aws_codeartifact_domain.test.domain
  repository = aws_codeartifact_repository.test.repository
  format     = %[1]q
}
`, format))
}

func testAccCheckAWSCodeArtifactRepositoryEndpointOwnerConfig(rName string) string {
	return composeConfig(
		testAccCheckAWSCodeArtifactRepositoryEndpointBaseConfig(rName),
		`
data "aws_codeartifact_repository_endpoint" "test" {
  domain       = aws_codeartifact_domain.test.domain
  repository   = aws_codeartifact_repository.test.repository
  domain_owner = aws_codeartifact_domain.test.owner
  format       = "npm"
}
`)
}
