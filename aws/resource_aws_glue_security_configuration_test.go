package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_glue_security_configuration", &resource.Sweeper{
		Name: "aws_glue_security_configuration",
		F:    testSweepGlueSecurityConfigurations,
	})
}

func testSweepGlueSecurityConfigurations(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).glueconn

	input := &glue.GetSecurityConfigurationsInput{}

	for {
		output, err := conn.GetSecurityConfigurations(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Glue Security Configuration sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving Glue Security Configurations: %s", err)
		}

		for _, securityConfiguration := range output.SecurityConfigurations {
			name := aws.StringValue(securityConfiguration.Name)

			if !strings.HasPrefix(name, "tf-acc-test") {
				log.Printf("[INFO] Skipping Glue Security Configuration: %s", name)
				continue
			}

			log.Printf("[INFO] Deleting Glue Security Configuration: %s", name)
			err := deleteGlueSecurityConfiguration(conn, name)
			if err != nil {
				log.Printf("[ERROR] Failed to delete Glue Security Configuration %s: %s", name, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSGlueSecurityConfiguration_Basic(t *testing.T) {
	var securityConfiguration glue.SecurityConfiguration

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_security_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueSecurityConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueSecurityConfigurationConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueSecurityConfigurationExists(resourceName, &securityConfiguration),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.cloudwatch_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.cloudwatch_encryption.0.cloudwatch_encryption_mode", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.cloudwatch_encryption.0.kms_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.job_bookmarks_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.job_bookmarks_encryption.0.job_bookmarks_encryption_mode", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.job_bookmarks_encryption.0.kms_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.0.kms_key_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.0.s3_encryption_mode", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccAWSGlueSecurityConfiguration_CloudWatchEncryption_CloudWatchEncryptionMode_SSEKMS(t *testing.T) {
	var securityConfiguration glue.SecurityConfiguration

	rName := acctest.RandomWithPrefix("tf-acc-test")
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_glue_security_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueSecurityConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueSecurityConfigurationConfig_CloudWatchEncryption_CloudWatchEncryptionMode_SSEKMS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueSecurityConfigurationExists(resourceName, &securityConfiguration),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.cloudwatch_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.cloudwatch_encryption.0.cloudwatch_encryption_mode", "SSE-KMS"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.cloudwatch_encryption.0.kms_key_arn", kmsKeyResourceName, "arn"),
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

func TestAccAWSGlueSecurityConfiguration_JobBookmarksEncryption_JobBookmarksEncryptionMode_CSEKMS(t *testing.T) {
	var securityConfiguration glue.SecurityConfiguration

	rName := acctest.RandomWithPrefix("tf-acc-test")
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_glue_security_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueSecurityConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueSecurityConfigurationConfig_JobBookmarksEncryption_JobBookmarksEncryptionMode_CSEKMS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueSecurityConfigurationExists(resourceName, &securityConfiguration),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.job_bookmarks_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.job_bookmarks_encryption.0.job_bookmarks_encryption_mode", "CSE-KMS"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.job_bookmarks_encryption.0.kms_key_arn", kmsKeyResourceName, "arn"),
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

func TestAccAWSGlueSecurityConfiguration_S3Encryption_S3EncryptionMode_SSEKMS(t *testing.T) {
	var securityConfiguration glue.SecurityConfiguration

	rName := acctest.RandomWithPrefix("tf-acc-test")
	kmsKeyResourceName := "aws_kms_key.test"
	resourceName := "aws_glue_security_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueSecurityConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueSecurityConfigurationConfig_S3Encryption_S3EncryptionMode_SSEKMS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueSecurityConfigurationExists(resourceName, &securityConfiguration),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.0.s3_encryption_mode", "SSE-KMS"),
					resource.TestCheckResourceAttrPair(resourceName, "encryption_configuration.0.s3_encryption.0.kms_key_arn", kmsKeyResourceName, "arn"),
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

func TestAccAWSGlueSecurityConfiguration_S3Encryption_S3EncryptionMode_SSES3(t *testing.T) {
	var securityConfiguration glue.SecurityConfiguration

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_glue_security_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGlueSecurityConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGlueSecurityConfigurationConfig_S3Encryption_S3EncryptionMode_SSES3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGlueSecurityConfigurationExists(resourceName, &securityConfiguration),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.0.s3_encryption_mode", "SSE-S3"),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.s3_encryption.0.kms_key_arn", ""),
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

func testAccCheckAWSGlueSecurityConfigurationExists(resourceName string, securityConfiguration *glue.SecurityConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Security Configuration ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn

		output, err := conn.GetSecurityConfiguration(&glue.GetSecurityConfigurationInput{
			Name: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if output.SecurityConfiguration == nil {
			return fmt.Errorf("Glue Security Configuration (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(output.SecurityConfiguration.Name) == rs.Primary.ID {
			*securityConfiguration = *output.SecurityConfiguration
			return nil
		}

		return fmt.Errorf("Glue Security Configuration (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckAWSGlueSecurityConfigurationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_glue_security_configuration" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).glueconn

		output, err := conn.GetSecurityConfiguration(&glue.GetSecurityConfigurationInput{
			Name: aws.String(rs.Primary.ID),
		})

		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		securityConfiguration := output.SecurityConfiguration
		if securityConfiguration != nil && aws.StringValue(securityConfiguration.Name) == rs.Primary.ID {
			return fmt.Errorf("Glue Security Configuration %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAWSGlueSecurityConfigurationConfig_Basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_security_configuration" "test" {
  name = %q

  encryption_configuration {
    cloudwatch_encryption {
      cloudwatch_encryption_mode = "DISABLED"
    }

    job_bookmarks_encryption {
      job_bookmarks_encryption_mode = "DISABLED"
    }

    s3_encryption {
      s3_encryption_mode = "DISABLED"
    }
  }
}
`, rName)
}

func testAccAWSGlueSecurityConfigurationConfig_CloudWatchEncryption_CloudWatchEncryptionMode_SSEKMS(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_glue_security_configuration" "test" {
  name = %q

  encryption_configuration {
    cloudwatch_encryption {
      cloudwatch_encryption_mode = "SSE-KMS"
      kms_key_arn                = "${aws_kms_key.test.arn}"
    }

    job_bookmarks_encryption {
      job_bookmarks_encryption_mode = "DISABLED"
    }

    s3_encryption {
      s3_encryption_mode = "DISABLED"
    }
  }
}
`, rName)
}

func testAccAWSGlueSecurityConfigurationConfig_JobBookmarksEncryption_JobBookmarksEncryptionMode_CSEKMS(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_glue_security_configuration" "test" {
  name = %q

  encryption_configuration {
    cloudwatch_encryption {
      cloudwatch_encryption_mode = "DISABLED"
    }

    job_bookmarks_encryption {
      job_bookmarks_encryption_mode = "CSE-KMS"
      kms_key_arn                   = "${aws_kms_key.test.arn}"
    }

    s3_encryption {
      s3_encryption_mode = "DISABLED"
    }
  }
}
`, rName)
}

func testAccAWSGlueSecurityConfigurationConfig_S3Encryption_S3EncryptionMode_SSEKMS(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_glue_security_configuration" "test" {
  name = %q

  encryption_configuration {
    cloudwatch_encryption {
      cloudwatch_encryption_mode = "DISABLED"
    }

    job_bookmarks_encryption {
      job_bookmarks_encryption_mode = "DISABLED"
    }

    s3_encryption {
      kms_key_arn        = "${aws_kms_key.test.arn}"
      s3_encryption_mode = "SSE-KMS"
    }
  }
}
`, rName)
}

func testAccAWSGlueSecurityConfigurationConfig_S3Encryption_S3EncryptionMode_SSES3(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_security_configuration" "test" {
  name = %q

  encryption_configuration {
    cloudwatch_encryption {
      cloudwatch_encryption_mode = "DISABLED"
    }

    job_bookmarks_encryption {
      job_bookmarks_encryption_mode = "DISABLED"
    }

    s3_encryption {
      s3_encryption_mode = "SSE-S3"
    }
  }
}
`, rName)
}
