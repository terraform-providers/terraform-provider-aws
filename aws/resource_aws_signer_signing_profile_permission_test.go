package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/signer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
)

func TestAccAWSSignerSigningProfilePermission_basic(t *testing.T) {
	resourceName := "aws_signer_signing_profile_permission.test_sp_permission"
	profileResourceName := "aws_signer_signing_profile.test_sp"
	rString := acctest.RandString(53)
	profileName := fmt.Sprintf("tf_acc_spp_%s", rString)

	var conf signer.GetSigningProfileOutput
	var sppconf signer.ListProfilePermissionsOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSignerSigningProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccAWSSignerSigningProfilePermissionConfig(profileName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSignerSigningProfileExists(profileResourceName, &conf),
					testAccCheckAWSSignerSigningProfilePermissionExists(resourceName, profileName, &sppconf),
					naming.TestCheckResourceAttrNameGenerated(resourceName, "statement_id"),
				),
			},
			{
				ResourceName:            profileResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSSignerSigningProfilePermission_GetSigningProfile(t *testing.T) {
	resourceName := "aws_signer_signing_profile_permission.test_sp_permission"
	profileResourceName := "aws_signer_signing_profile.test_sp"
	rString := acctest.RandString(53)
	profileName := fmt.Sprintf("tf_acc_spp_%s", rString)

	var conf signer.GetSigningProfileOutput
	var sppconf signer.ListProfilePermissionsOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSignerSigningProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config:  testAccAWSSignerSigningProfilePermissionGetSP(profileName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSignerSigningProfileExists(profileResourceName, &conf),
					testAccCheckAWSSignerSigningProfilePermissionExists(resourceName, profileName, &sppconf),
				),
			},
			{
				ResourceName:            profileResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
			{
				Config:  testAccAWSSignerSigningProfilePermissionRevokeSignature(profileName),
				Destroy: false,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSignerSigningProfileExists(profileResourceName, &conf),
					testAccCheckAWSSignerSigningProfilePermissionExists(resourceName, profileName, &sppconf),
				),
			},
		},
	})
}

func TestAccAWSSignerSigningProfilePermission_StartSigningJob_GetSP(t *testing.T) {
	resourceName1 := "aws_signer_signing_profile_permission.sp1_perm"
	resourceName2 := "aws_signer_signing_profile_permission.sp2_perm"
	profileResourceName := "aws_signer_signing_profile.test_sp"
	rString := acctest.RandString(53)
	profileName := fmt.Sprintf("tf_acc_spp_%s", rString)

	var conf signer.GetSigningProfileOutput
	var sppconf signer.ListProfilePermissionsOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSignerSigningProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSignerSigningProfilePermissionStartSigningJobGetSP(profileName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSignerSigningProfileExists(profileResourceName, &conf),
					testAccCheckAWSSignerSigningProfilePermissionExists(resourceName1, profileName, &sppconf),
					testAccCheckAWSSignerSigningProfilePermissionExists(resourceName2, profileName, &sppconf),
				),
			},
			{
				ResourceName:            profileResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSSignerSigningProfilePermission_StatementPrefix(t *testing.T) {
	resourceName := "aws_signer_signing_profile_permission.sp1_perm"
	profileResourceName := "aws_signer_signing_profile.test_sp"
	rString := acctest.RandString(53)
	profileName := fmt.Sprintf("tf_acc_spp_%s", rString)
	statementNamePrefix := "tf_acc_spp_statement_"

	//var conf signer.GetSigningProfileOutput
	var sppconf signer.ListProfilePermissionsOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckSingerSigningProfile(t, "AWSLambda-SHA384-ECDSA") },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSignerSigningProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSignerSigningProfilePermissionStatementPrefix(statementNamePrefix, profileName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSignerSigningProfilePermissionExists(resourceName, profileName, &sppconf),
					naming.TestCheckResourceAttrNameFromPrefix(resourceName, "statement_id", statementNamePrefix),
				),
			},
			{
				ResourceName:            profileResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func testAccAWSSignerSigningProfilePermissionConfig(profileName string) string {
	return fmt.Sprintf(testAccAWSSignerSigningProfilePermissionConfig_base(profileName) + `
data "aws_caller_identity" "current" {}

resource "aws_signer_signing_profile_permission" "test_sp_permission" {
  profile_name = aws_signer_signing_profile.test_sp.name
  action       = "signer:StartSigningJob"
  principal    = data.aws_caller_identity.current.account_id
}`)
}

func testAccAWSSignerSigningProfilePermissionStartSigningJobGetSP(profileName string) string {
	return fmt.Sprintf(testAccAWSSignerSigningProfilePermissionConfig_base(profileName) + `
data "aws_caller_identity" "current" {}

resource "aws_signer_signing_profile_permission" "sp1_perm" {
  profile_name = aws_signer_signing_profile.test_sp.name
  action       = "signer:StartSigningJob"
  principal    = data.aws_caller_identity.current.account_id
  statement_id = "statementid1"
}

resource "aws_signer_signing_profile_permission" "sp2_perm" {
  profile_name = aws_signer_signing_profile.test_sp.name
  action       = "signer:GetSigningProfile"
  principal    = data.aws_caller_identity.current.account_id
  statement_id = "statementid2"
}`)
}

func testAccAWSSignerSigningProfilePermissionStatementPrefix(statementNamePrefix, profileName string) string {
	return fmt.Sprintf(testAccAWSSignerSigningProfilePermissionConfig_base(profileName)+`
data "aws_caller_identity" "current" {}

resource "aws_signer_signing_profile_permission" "sp1_perm" {
  profile_name        = aws_signer_signing_profile.test_sp.name
  action              = "signer:StartSigningJob"
  principal           = data.aws_caller_identity.current.account_id
  statement_id_prefix = %[1]q
}`, statementNamePrefix)
}

func testAccAWSSignerSigningProfilePermissionGetSP(profileName string) string {
	return fmt.Sprintf(testAccAWSSignerSigningProfilePermissionConfig_base(profileName) + `
data "aws_caller_identity" "current" {}

resource "aws_signer_signing_profile_permission" "test_sp_permission" {
  profile_name = aws_signer_signing_profile.test_sp.name
  action       = "signer:GetSigningProfile"
  principal    = data.aws_caller_identity.current.account_id
}`)
}

func testAccAWSSignerSigningProfilePermissionRevokeSignature(profileName string) string {
	return fmt.Sprintf(testAccAWSSignerSigningProfilePermissionConfig_base(profileName) + `
data "aws_caller_identity" "current" {}

resource "aws_signer_signing_profile_permission" "test_sp_permission" {
  profile_name = aws_signer_signing_profile.test_sp.name
  action       = "signer:RevokeSignature"
  principal    = data.aws_caller_identity.current.account_id
}`)
}

func testAccAWSSignerSigningProfilePermissionConfig_base(profileName string) string {
	return fmt.Sprintf(`
resource "aws_signer_signing_profile" "test_sp" {
  platform_id = "AWSLambda-SHA384-ECDSA"
  name        = "%s"
}`, profileName)
}

func testAccCheckAWSSignerSigningProfilePermissionExists(res, profileName string, spp *signer.ListProfilePermissionsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[res]
		if !ok {
			return fmt.Errorf("Signing profile permission not found: %s", res)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Signing Profile with that ARN does not exist")
		}

		conn := testAccProvider.Meta().(*AWSClient).signerconn

		params := &signer.ListProfilePermissionsInput{
			ProfileName: aws.String(profileName),
		}

		getSp, err := conn.ListProfilePermissions(params)
		if err != nil {
			return err
		}

		*spp = *getSp

		return nil
	}
}
