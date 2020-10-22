package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

			_, err = conn.DeleteRepository(&ecr.DeleteRepositoryInput{
				// We should probably sweep repositories even if there are images.
				Force:          aws.Bool(true),
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
	var v ecr.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRepositoryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "ecr", fmt.Sprintf("repository/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccCheckAWSEcrRepositoryRegistryID(resourceName),
					testAccCheckAWSEcrRepositoryRepositoryURL(resourceName, rName),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_type", ecr.EncryptionTypeAes256),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.kms_key", ""),
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
	var v1, v2 ecr.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRepositoryConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
				),
			},
			{
				Config: testAccAWSEcrRepositoryConfig_tagsChanged(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "changed"),
				),
			},
		},
	})
}

func TestAccAWSEcrRepository_immutability(t *testing.T) {
	var v ecr.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRepositoryConfig_immutability(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryExists(resourceName, &v),
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
	var v1, v2 ecr.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRepositoryConfig_image_scanning_configuration(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryExists(resourceName, &v1),
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
					testAccCheckAWSEcrRepositoryExists(resourceName, &v2),
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

func TestAccAWSEcrRepository_encryption_kms(t *testing.T) {
	var v1, v2 ecr.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecr_repository.test"
	kmsKeyDataSourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrRepositoryConfig_encryption_kms_defaultkey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_type", ecr.EncryptionTypeKms),
					// This will be the default ECR service KMS key. We don't currently have a way to look this up.
					testAccMatchResourceAttrRegionalARN(resourceName, "encryption_configuration.0.kms_key", "kms", regexp.MustCompile(fmt.Sprintf("key/%s$", uuidRegexPattern))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEcrRepositoryConfig_encryption_kms_customkey(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryExists(resourceName, &v2),
					testAccCheckAWSEcrRepositoryRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_type", ecr.EncryptionTypeKms),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.kms_key", kmsKeyDataSourceName, "arn"),
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

func TestAccAWSEcrRepository_encryption_aes256(t *testing.T) {
	var v1, v2 ecr.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecr_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				// Test that the addition of the default encryption_configuration doesn't recreation in the next step
				Config: testAccAWSEcrRepositoryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryExists(resourceName, &v1),
				),
			},
			{
				Config: testAccAWSEcrRepositoryConfig_encryption_aes256(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrRepositoryExists(resourceName, &v2),
					testAccCheckAWSEcrRepositoryNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.encryption_type", ecr.EncryptionTypeAes256),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.kms_key", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Test that the removal of the default encryption_configuration doesn't cause any plan changes
				Config:   testAccAWSEcrRepositoryConfig(rName),
				PlanOnly: true,
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

func testAccCheckAWSEcrRepositoryExists(name string, res *ecr.Repository) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ECR repository ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ecrconn

		output, err := conn.DescribeRepositories(&ecr.DescribeRepositoriesInput{
			RepositoryNames: aws.StringSlice([]string{rs.Primary.ID}),
		})
		if err != nil {
			return err
		}
		if len(output.Repositories) == 0 {
			return fmt.Errorf("ECR repository %s not found", rs.Primary.ID)
		}

		*res = *output.Repositories[0]

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

func testAccCheckAWSEcrRepositoryRecreated(i, j *ecr.Repository) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreatedAt) == aws.TimeValue(j.CreatedAt) {
			return fmt.Errorf("ECR repository was not recreated")
		}

		return nil
	}
}

func testAccCheckAWSEcrRepositoryNotRecreated(i, j *ecr.Repository) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreatedAt) != aws.TimeValue(j.CreatedAt) {
			return fmt.Errorf("ECR repository was recreated")
		}

		return nil
	}
}

func testAccAWSEcrRepositoryConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %q
}
`, rName)
}

func testAccAWSEcrRepositoryConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
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
resource "aws_ecr_repository" "test" {
  name = %q

  tags = {
    Usage = "changed"
  }
}
`, rName)
}

func testAccAWSEcrRepositoryConfig_immutability(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name                 = %q
  image_tag_mutability = "IMMUTABLE"
}
`, rName)
}

func testAccAWSEcrRepositoryConfig_image_scanning_configuration(rName string, scanOnPush bool) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %q

  image_scanning_configuration {
    scan_on_push = %t
  }
}
`, rName, scanOnPush)
}

func testAccAWSEcrRepositoryConfig_encryption_kms_defaultkey(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %q

  encryption_configuration {
    encryption_type = "KMS"
  }
}
`, rName)
}

func testAccAWSEcrRepositoryConfig_encryption_kms_customkey(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {}

resource "aws_ecr_repository" "test" {
  name = %q

  encryption_configuration {
    encryption_type = "KMS"
    kms_key         = aws_kms_key.test.arn
  }
}
`, rName)
}

func testAccAWSEcrRepositoryConfig_encryption_aes256(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "test" {
  name = %q

  encryption_configuration {
    encryption_type = "AES256"
  }
}
`, rName)
}
