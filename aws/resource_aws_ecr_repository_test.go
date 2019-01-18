package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSEcrRepository_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecr_repository.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRepositoryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "ecr", fmt.Sprintf("repository/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccCheckAWSEcrRepositoryRegistryID(resourceName),
					testAccCheckAWSEcrRepositoryRepositoryURL(resourceName, rName),
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

func TestAccAWSEcrRepository_tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecr_repository.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRepositoryConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
				),
			},
			{
				Config: testAccAWSEcrRepositoryConfig_tagsChanged(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "changed"),
				),
			},
		},
	})
}

func testAccCheckAWSEcrRepositoryDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ecrconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecr_repository" {
			continue
		}

		input := ecr.DescribeRepositoriesInput{
			RepositoryNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		}

		out, err := conn.DescribeRepositories(&input)

		if isAWSErr(err, ecr.ErrCodeRepositoryNotFoundException, "") {
			return nil
		}

		if err != nil {
			return err
		}

		for _, repository := range out.Repositories {
			if aws.StringValue(repository.RepositoryName) == rs.Primary.Attributes["name"] {
				return fmt.Errorf("ECR repository still exists: %s", rs.Primary.Attributes["name"])
			}
		}
	}

	return nil
}

func testAccCheckAWSEcrRepositoryExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccCheckAWSEcrRepositoryRegistryID(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := testAccGetAccountID()
		return resource.TestCheckResourceAttr(resourceName, "registry_id", attributeValue)(s)
	}
}

func testAccCheckAWSEcrRepositoryRepositoryURL(resourceName, repositoryName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		attributeValue := fmt.Sprintf("%s.dkr.%s/%s", testAccGetAccountID(), testAccGetServiceEndpoint("ecr"), repositoryName)
		return resource.TestCheckResourceAttr(resourceName, "repository_url", attributeValue)(s)
	}
}

func testAccAWSEcrRepositoryConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "default" {
  name = %q
}
`, rName)
}

func testAccAWSEcrRepositoryConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "default" {
  name = %q

  tags = {
    Environment = "production"
    Usage = "original"
  }
}
`, rName)
}

func testAccAWSEcrRepositoryConfig_tagsChanged(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "default" {
  name = %q

  tags = {
    Usage = "changed"
  }
}
`, rName)
}
