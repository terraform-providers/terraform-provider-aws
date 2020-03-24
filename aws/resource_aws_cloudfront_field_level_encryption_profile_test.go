package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSCloudfrontFieldLevelEncryptionProfile_basic(t *testing.T) {
	var profile cloudfront.GetFieldLevelEncryptionProfileOutput
	resourceName := "aws_cloudfront_field_level_encryption_profile.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFront(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFieldLevelEncryptionProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontFieldLevelEncryptionProfileConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudfrontFieldLevelEncryptionProfileExists(resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, "comment", "some comment"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "encryption_entities.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudfrontFieldLevelEncryptionProfileExtendedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudfrontFieldLevelEncryptionProfileExists(resourceName, &profile),
					resource.TestCheckResourceAttr(resourceName, "comment", "some other comment"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "encryption_entities.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "etag"),
				),
			},
		},
	})
}

func TestAccAWSCloudfrontFieldLevelEncryptionProfile_disappears(t *testing.T) {
	var profile cloudfront.GetFieldLevelEncryptionProfileOutput
	resourceName := "aws_cloudfront_field_level_encryption_profile.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFront(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudfrontFieldLevelEncryptionProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudfrontFieldLevelEncryptionProfileConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudfrontFieldLevelEncryptionProfileExists(resourceName, &profile),
					testAccCheckCloudfrontFieldLevelEncryptionProfileDisappears(&profile),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCloudfrontFieldLevelEncryptionProfileDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_field_level_encryption_profile" {
			continue
		}

		params := &cloudfront.GetFieldLevelEncryptionProfileInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetFieldLevelEncryptionProfile(params)
		if err == nil {
			return fmt.Errorf("cloudfront Field Level Encryption Profile was not deleted")
		}
	}

	return nil
}

func testAccCheckCloudfrontFieldLevelEncryptionProfileExists(r string, profile *cloudfront.GetFieldLevelEncryptionProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("Not found: %s", r)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Id is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

		params := &cloudfront.GetFieldLevelEncryptionProfileInput{
			Id: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetFieldLevelEncryptionProfile(params)
		if err != nil {
			return fmt.Errorf("Error retrieving Cloudfront Field Level Encryption Profile: %s", err)
		}

		*profile = *resp

		return nil
	}
}

func testAccCheckCloudfrontFieldLevelEncryptionProfileDisappears(profile *cloudfront.GetFieldLevelEncryptionProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

		params := &cloudfront.DeleteFieldLevelEncryptionProfileInput{
			Id:      profile.FieldLevelEncryptionProfile.Id,
			IfMatch: profile.ETag,
		}

		_, err := conn.DeleteFieldLevelEncryptionProfile(params)

		return err
	}
}

func testAccAWSCloudfrontFieldLevelEncryptionProfileConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test" {
  comment     = "test key"
  encoded_key = "${file("test-fixtures/cloudfront-public-key.pem")}"
  name        = %[1]q
}

resource "aws_cloudfront_field_level_encryption_profile" "test" {
  comment = "some comment"
  name    = %[1]q

  encryption_entities {
    public_key_id  = "${aws_cloudfront_public_key.test.id}"
    provider_id    = %[1]q
    field_patterns = ["DateOfBirth"]
  }
}
`, rName)
}

func testAccAWSCloudfrontFieldLevelEncryptionProfileExtendedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test" {
  comment     = "test key"
  encoded_key = "${file("test-fixtures/cloudfront-public-key.pem")}"
  name        = %[1]q
}

resource "aws_cloudfront_field_level_encryption_profile" "test" {
  comment = "some other comment"
  name    = %[1]q

  encryption_entities {
    public_key_id  = "${aws_cloudfront_public_key.test.id}"
    provider_id    = %[1]q
    field_patterns = ["FirstName", "DateOfBirth"]
  }
}
`, rName)
}
