package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCodeArtifactRepositoryPermissionsPolicy_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_repository_permissions_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codeartifact.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeArtifactRepositoryPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeArtifactRepositoryPermissionsPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactRepositoryPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", "aws_codeartifact_repository.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain", rName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexp.MustCompile("codeartifact:CreateRepository")),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", "owner"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeArtifactRepositoryPermissionsPolicyUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactRepositoryPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", "aws_codeartifact_repository.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain", rName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexp.MustCompile("codeartifact:CreateRepository")),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexp.MustCompile("codeartifact:ListRepositoriesInDomain")),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", "owner"),
				),
			},
		},
	})
}

func TestAccAWSCodeArtifactRepositoryPermissionsPolicy_owner(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_repository_permissions_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codeartifact.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeArtifactRepositoryPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeArtifactRepositoryPermissionsPolicyOwnerConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactRepositoryPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", "aws_codeartifact_repository.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain", rName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexp.MustCompile("codeartifact:CreateRepository")),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", "owner"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodeArtifactRepositoryPermissionsPolicy_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_repository_permissions_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codeartifact.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeArtifactRepositoryPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeArtifactRepositoryPermissionsPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactRepositoryPermissionsExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCodeArtifactRepositoryPermissionsPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCodeArtifactRepositoryPermissionsPolicy_disappears_domain(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_repository_permissions_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codeartifact.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeArtifactRepositoryPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeArtifactRepositoryPermissionsPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactRepositoryPermissionsExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCodeArtifactRepositoryPermissionsPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSCodeArtifactRepositoryPermissionsExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no CodeArtifact domain set")
		}

		conn := testAccProvider.Meta().(*AWSClient).codeartifactconn

		domainOwner, domainName, repoName, err := decodeCodeArtifactRepositoryID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.GetRepositoryPermissionsPolicy(&codeartifact.GetRepositoryPermissionsPolicyInput{
			Domain:      aws.String(domainName),
			DomainOwner: aws.String(domainOwner),
			Repository:  aws.String(repoName),
		})

		return err
	}
}

func testAccCheckAWSCodeArtifactRepositoryPermissionsDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codeartifact_repository_permissions_policy" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).codeartifactconn

		domainOwner, domainName, repoName, err := decodeCodeArtifactRepositoryID(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := conn.GetRepositoryPermissionsPolicy(&codeartifact.GetRepositoryPermissionsPolicyInput{
			Domain:      aws.String(domainName),
			DomainOwner: aws.String(domainOwner),
			Repository:  aws.String(repoName),
		})

		if err == nil {
			if aws.StringValue(resp.Policy.ResourceArn) == rs.Primary.ID {
				return fmt.Errorf("CodeArtifact Domain %s still exists", rs.Primary.ID)
			}
		}

		if isAWSErr(err, codeartifact.ErrCodeResourceNotFoundException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccAWSCodeArtifactRepositoryPermissionsPolicyBasicConfig(rName string) string {
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

resource "aws_codeartifact_repository_permissions_policy" "test" {
  domain          = aws_codeartifact_domain.test.domain
  repository      = aws_codeartifact_repository.test.repository
  policy_document = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "codeartifact:CreateRepository",
            "Effect": "Allow",
            "Principal": "*",
            "Resource": "${aws_codeartifact_domain.test.arn}"
        }
    ]
}
EOF
}
`, rName)
}

func testAccAWSCodeArtifactRepositoryPermissionsPolicyOwnerConfig(rName string) string {
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

resource "aws_codeartifact_repository_permissions_policy" "test" {
  domain          = aws_codeartifact_domain.test.domain
  domain_owner    = aws_codeartifact_domain.test.owner
  repository      = aws_codeartifact_repository.test.repository
  policy_document = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "codeartifact:CreateRepository",
            "Effect": "Allow",
            "Principal": "*",
            "Resource": "${aws_codeartifact_domain.test.arn}"
        }
    ]
}
EOF
}
`, rName)
}

func testAccAWSCodeArtifactRepositoryPermissionsPolicyUpdatedConfig(rName string) string {
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

resource "aws_codeartifact_repository_permissions_policy" "test" {
  domain          = aws_codeartifact_domain.test.domain
  repository      = aws_codeartifact_repository.test.repository
  policy_document = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
 				"codeartifact:CreateRepository",
				"codeartifact:ListRepositoriesInDomain"
			],
            "Effect": "Allow",
            "Principal": "*",
            "Resource": "${aws_codeartifact_domain.test.arn}"
        }
    ]
}
EOF
}
`, rName)
}
