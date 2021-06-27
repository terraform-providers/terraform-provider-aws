package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudfront/finder"
)

func TestAccAWSCloudfrontFieldLevelEncryptionConfig_basic(t *testing.T) {
	var profile cloudfront.GetFieldLevelEncryptionConfigOutput
	resourceName := "aws_cloudfront_field_level_encryption_config.test"
	profileResourceName := "aws_cloudfront_field_level_encryption_profile.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFieldLevelEncryptionConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontFieldLevelEncryptionConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudfrontFieldLevelEncryptionConfigExists(resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, "comment", "some comment"),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.0.content_type_profile.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.0.content_type_profile.0.content_type", "application/x-www-form-urlencoded"),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.0.content_type_profile.0.format", "URLEncoded"),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.0.forward_when_content_type_is_unknown", "true"),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.0.query_arg_profile.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.0.query_arg_profile.0.query_arg", "URLEncoded"),
					resource.TestCheckResourceAttrPair(resourceName, "query_arg_profile_config.0.query_arg_profile.0.profile_id", profileResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.0.forward_when_query_arg_is_unknown", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudfrontFieldLevelEncryptionConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudfrontFieldLevelEncryptionConfigExists(resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, "comment", "some comment"),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.0.content_type_profile.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.0.content_type_profile.0.content_type", "application/x-www-form-urlencoded"),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.0.content_type_profile.0.format", "URLEncoded"),
					resource.TestCheckResourceAttr(resourceName, "content_type_profile_config.0.forward_when_content_type_is_unknown", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "content_type_profile_config.0.content_type_profile.0.profile_id", profileResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.0.query_arg_profile.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.0.query_arg_profile.0.query_arg", "URLEncoded2"),
					resource.TestCheckResourceAttrPair(resourceName, "query_arg_profile_config.0.query_arg_profile.0.profile_id", profileResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "query_arg_profile_config.0.forward_when_query_arg_is_unknown", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
				),
			},
		},
	})
}

func TestAccAWSCloudfrontFieldLevelEncryptionConfig_disappears(t *testing.T) {
	var profile cloudfront.GetFieldLevelEncryptionConfigOutput
	resourceName := "aws_cloudfront_field_level_encryption_config.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFieldLevelEncryptionConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontFieldLevelEncryptionConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudfrontFieldLevelEncryptionConfigExists(resourceName, &profile),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudfrontFieldLevelEncryptionConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCloudfrontFieldLevelEncryptionConfigDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_field_level_encryption_config" {
			continue
		}

		_, err := finder.FieldLevelEncryptionConfigByID(conn, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("cloudfront Field Level Encryption Config was not deleted")
		}
	}

	return nil
}

func testAccCheckCloudfrontFieldLevelEncryptionConfigExists(r string, profile *cloudfront.GetFieldLevelEncryptionConfigOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Id is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

		resp, err := finder.FieldLevelEncryptionConfigByID(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error retrieving Cloudfront Field Level Encryption Config: %w", err)
		}

		*profile = *resp

		return nil
	}
}

func testAccAWSCloudfrontFieldLevelEncryptionConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test" {
  comment     = "test key"
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
  name        = %[1]q
}

resource "aws_cloudfront_field_level_encryption_profile" "test" {
  comment = "some comment"
  name    = %[1]q

  encryption_entities {
    public_key_id  = aws_cloudfront_public_key.test.id
    provider_id    = %[1]q
    field_patterns = ["DateOfBirth"]
  }
}
`, rName)
}

func testAccAWSCloudfrontFieldLevelEncryptionConfigBasic(rName string) string {
	return testAccAWSCloudfrontFieldLevelEncryptionConfigBase(rName) + fmt.Sprintf(`
resource "aws_cloudfront_field_level_encryption_config" "test" {
  comment = "some comment"

  content_type_profile_config {
    forward_when_content_type_is_unknown = true

    content_type_profile {
      content_type = "application/x-www-form-urlencoded"
      format       = "URLEncoded"
    }
  }

  query_arg_profile_config {
    forward_when_query_arg_is_unknown = true

    query_arg_profile {
      profile_id = aws_cloudfront_field_level_encryption_profile.test.id
      query_arg  = "URLEncoded"
    }
  }
}
`)
}

func testAccAWSCloudfrontFieldLevelEncryptionConfigUpdated(rName string) string {
	return testAccAWSCloudfrontFieldLevelEncryptionConfigBase(rName) + fmt.Sprintf(`
resource "aws_cloudfront_field_level_encryption_config" "test" {
  comment = "some comment"

  content_type_profile_config {
    forward_when_content_type_is_unknown = false

    content_type_profile {
      content_type = "application/x-www-form-urlencoded"
      format       = "URLEncoded"
      profile_id   = aws_cloudfront_field_level_encryption_profile.test.id
    }
  }

  query_arg_profile_config {
    forward_when_query_arg_is_unknown = false

    query_arg_profile {
      profile_id = aws_cloudfront_field_level_encryption_profile.test.id
      query_arg  = "URLEncoded2"
    }
  }
}
`)
}
