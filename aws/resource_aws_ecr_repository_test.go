package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_ecr_repository", &resource.Sweeper{
		Name: "aws_ecr_repository",
		F:    testSweepEcrRepositories,
	})
}

func testSweepEcrRepositories(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).ecrconn

	var errors error
	err = conn.DescribeRepositoriesPages(&ecr.DescribeRepositoriesInput{}, func(page *ecr.DescribeRepositoriesOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, repository := range page.Repositories {
			repositoryName := aws.StringValue(repository.RepositoryName)
			log.Printf("[INFO] Deleting ECR repository: %s", repositoryName)

			shouldForce := true
			_, err = conn.DeleteRepository(&ecr.DeleteRepositoryInput{
				// We should probably sweep repositories even if there are images.
				Force:          &shouldForce,
				RegistryId:     repository.RegistryId,
				RepositoryName: repository.RepositoryName,
			})
			if err != nil {
				if !isAWSErr(err, ecr.ErrCodeRepositoryNotFoundException, "") {
					sweeperErr := fmt.Errorf("Error deleting ECR repository (%s): %w", repositoryName, err)
					log.Printf("[ERROR] %s", sweeperErr)
					errors = multierror.Append(errors, sweeperErr)
				}
				continue
			}
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping ECR repository sweep for %s: %s", region, err)
			return nil
		}
		errors = multierror.Append(errors, fmt.Errorf("Error retreiving ECR repositories: %w", err))
	}

	return errors
}

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

func TestAccAWSEcrRepository_immutability(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecr_repository.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRepositoryConfig_immutability(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "image_tag_mutability", "IMMUTABLE"),
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

func TestAccAWSEcrRepository_image_scanning_configuration(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecr_repository.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRepositoryConfig_image_scanning_configuration(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.0.scan_on_push", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test that the removal of the non-default image_scanning_configuration causes plan changes
				Config:             testAccAWSEcrRepositoryConfig(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				// Test attribute update
				Config: testAccAWSEcrRepositoryConfig_image_scanning_configuration(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "image_scanning_configuration.0.scan_on_push", "false"),
				),
			},
			{
				// Test that the removal of the default image_scanning_configuration doesn't cause any plan changes
				Config:             testAccAWSEcrRepositoryConfig(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
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
		attributeValue := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s", testAccGetAccountID(), testAccGetRegion(), repositoryName)
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
    Usage       = "original"
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

func testAccAWSEcrRepositoryConfig_immutability(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "default" {
  name = %q
  image_tag_mutability = "IMMUTABLE"
}
`, rName)
}

func testAccAWSEcrRepositoryConfig_image_scanning_configuration(rName string, scanOnPush bool) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "default" {
  name = %q
  image_scanning_configuration {
    scan_on_push = %t
  }
}
`, rName, scanOnPush)
}
