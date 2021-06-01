package aws

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/signer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSignerSigningProfile_basic(t *testing.T) {
	resourceName := "aws_signer_signing_profile.test_sp"
	rString := acctest.RandString(48)
	profileName := fmt.Sprintf("tf_acc_sp_basic_%s", rString)

	var conf signer.GetSigningProfileOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSignerSigningProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSignerSigningProfileConfigProvidedName(profileName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSignerSigningProfileExists(resourceName, &conf),
					resource.TestMatchResourceAttr(resourceName, "name",
						regexp.MustCompile("^[a-zA-Z0-9_]{0,64}$")),
					resource.TestCheckResourceAttr(resourceName, "platform_id", "AWSLambda-SHA384-ECDSA"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSSignerSigningProfile_GenerateNameWithNamePrefix(t *testing.T) {
	resourceName := "aws_signer_signing_profile.test_sp"
	namePrefix := "tf_acc_sp_basic_"

	var conf signer.GetSigningProfileOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSignerSigningProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSignerSigningProfileConfig(namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSignerSigningProfileExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "platform_id", "AWSLambda-SHA384-ECDSA"),
				),
			},
		},
	})
}

func TestAccAWSSignerSigningProfile_GenerateName(t *testing.T) {
	resourceName := "aws_signer_signing_profile.test_sp"

	var conf signer.GetSigningProfileOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSignerSigningProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSignerSigningProfileConfigGenerateName(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSignerSigningProfileExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "platform_id", "AWSLambda-SHA384-ECDSA"),
				),
			},
		},
	})
}

func TestAccAWSSignerSigningProfile_tags(t *testing.T) {
	resourceName := "aws_signer_signing_profile.test_sp"
	namePrefix := "tf_acc_sp_basic_"

	var conf signer.GetSigningProfileOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSignerSigningProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSignerSigningProfileConfigTags(namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSignerSigningProfileExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "value2"),
				),
			},
			{
				Config: testAccAWSSignerSigningProfileUpdateTags(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSignerSigningProfileExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "prod"),
				),
			},
		},
	})
}

func TestAccAWSSignerSigningProfile_SignatureValidityPeriod(t *testing.T) {
	resourceName := "aws_signer_signing_profile.test_sp"
	namePrefix := "tf_acc_sp_basic_"

	var conf signer.GetSigningProfileOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSignerSigningProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSignerSigningProfileConfigSVP(namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSignerSigningProfileExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.0.type", "DAYS"),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.0.value", "10"),
				),
			},
			{
				Config: testAccAWSSignerSigningProfileUpdateSVP(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSignerSigningProfileExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.0.type", "MONTHS"),
					resource.TestCheckResourceAttr(resourceName, "signature_validity_period.0.value", "10"),
				),
			},
		},
	})
}

func testAccPreCheckSingerSigningProfile(t *testing.T, platformID string) {
	conn := testAccProvider.Meta().(*AWSClient).signerconn

	input := &signer.ListSigningPlatformsInput{}

	output, err := conn.ListSigningPlatforms(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if output == nil {
		t.Skip("skipping acceptance testing: empty response")
	}

	for _, platform := range output.Platforms {
		if platform == nil {
			continue
		}

		if aws.StringValue(platform.PlatformId) == platformID {
			return
		}
	}

	t.Skipf("skipping acceptance testing: Signing Platform (%s) not found", platformID)
}

func testAccAWSSignerSigningProfileConfig(namePrefix string) string {
	return baseAccAWSSignerSigningProfileConfig(namePrefix)
}

func testAccAWSSignerSigningProfileConfigGenerateName() string {
	return `
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}`
}

func testAccAWSSignerSigningProfileConfigProvidedName(profileName string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name        = "%s"
}`, profileName)
}

func testAccAWSSignerSigningProfileConfigTags(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name_prefix = "%s"
  tags = {
    "tag1" = "value1"
    "tag2" = "value2"
  }
}`, namePrefix)
}

func testAccAWSSignerSigningProfileConfigSVP(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name_prefix = "%s"

  signature_validity_period {
    value = 10
    type  = "DAYS"
  }
}
`, namePrefix)
}

func testAccAWSSignerSigningProfileUpdateSVP() string {
	return `
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"

  signature_validity_period {
    value = 10
    type  = "MONTHS"
  }
}
`
}

func testAccAWSSignerSigningProfileUpdateTags() string {
	return `
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  tags = {
    "tag1" = "prod"
  }
}
`
}

func baseAccAWSSignerSigningProfileConfig(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name_prefix = "%s"
}
`, namePrefix)
}

func testAccCheckAWSSignerSigningProfileExists(res string, sp *signer.GetSigningProfileOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("Signing profile not found: %s", res)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Signing Profile with that ARN does not exist")
		}

		conn := testAccProvider.Meta().(*AWSClient).signerconn

		params := &signer.GetSigningProfileInput{
			ProfileName: aws.String(rs.Primary.ID),
		}

		getSp, err := conn.GetSigningProfile(params)
		if err != nil {
			return err
		}

		*sp = *getSp

		return nil
	}
}

func testAccCheckAWSSignerSigningProfileDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).signerconn

	time.Sleep(5 * time.Second)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_signer_signing_profile" {
			continue
		}

		out, err := conn.GetSigningProfile(&signer.GetSigningProfileInput{
			ProfileName: aws.String(rs.Primary.ID),
		})

		if *out.Status != signer.SigningProfileStatusCanceled && err == nil {
			return fmt.Errorf("Signing Profile not cancelled%s", *out.ProfileName)
		}

	}

	return nil
}
