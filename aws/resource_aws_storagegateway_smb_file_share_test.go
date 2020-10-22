package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSStorageGatewaySmbFileShare_Authentication_ActiveDirectory(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"
	gatewayResourceName := "aws_storagegateway_gateway.test"
	bucketResourceName := "aws_s3_bucket.test"
	iamResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_Authentication_ActiveDirectory(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`share/share-.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication", "ActiveDirectory"),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_STANDARD"),
					resource.TestMatchResourceAttr(resourceName, "fileshare_id", regexp.MustCompile(`^share-`)),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", gatewayResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_arn", ""),
					resource.TestCheckResourceAttrPair(resourceName, "location_arn", bucketResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "object_acl", storagegateway.ObjectACLPrivate),
					resource.TestMatchResourceAttr(resourceName, "path", regexp.MustCompile(`^/.+`)),
					resource.TestCheckResourceAttr(resourceName, "read_only", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", iamResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "admin_user_list.#", "0"),
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

func TestAccAWSStorageGatewaySmbFileShare_Authentication_GuestAccess(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"
	gatewayResourceName := "aws_storagegateway_gateway.test"
	bucketResourceName := "aws_s3_bucket.test"
	iamResourceName := "aws_iam_role.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_Authentication_GuestAccess(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "admin_user_list.#", "0"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`share/share-.+`)),
					resource.TestCheckResourceAttr(resourceName, "authentication", "GuestAccess"),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "case_sensitivity", "ClientSpecified"),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_STANDARD"),
					resource.TestMatchResourceAttr(resourceName, "fileshare_id", regexp.MustCompile(`^share-`)),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_arn", gatewayResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "kms_key_arn", ""),
					resource.TestCheckResourceAttrPair(resourceName, "location_arn", bucketResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "object_acl", storagegateway.ObjectACLPrivate),
					resource.TestCheckResourceAttr(resourceName, "read_only", "false"),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", iamResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", "0"),
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

func TestAccAWSStorageGatewaySmbFileShare_DefaultStorageClass(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_DefaultStorageClass(rName, "S3_STANDARD_IA"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_STANDARD_IA"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_DefaultStorageClass(rName, "S3_ONEZONE_IA"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "default_storage_class", "S3_ONEZONE_IA"),
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

func TestAccAWSStorageGatewaySmbFileShare_Tags(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSStorageGatewaySmbFileShare_GuessMIMETypeEnabled(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_GuessMIMETypeEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", "false"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_GuessMIMETypeEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "guess_mime_type_enabled", "true"),
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

func TestAccAWSStorageGatewaySmbFileShare_InvalidUserList(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_InvalidUserList_Single(rName, "invaliduser1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", "1"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_InvalidUserList_Multiple(rName, "invaliduser2", "invaliduser3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", "2"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_InvalidUserList_Single(rName, "invaliduser4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "invalid_user_list.#", "1"),
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

func TestAccAWSStorageGatewaySmbFileShare_KMSEncrypted(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSStorageGatewaySmbFileShareConfig_KMSEncrypted(rName, true),
				ExpectError: regexp.MustCompile(`KMSKey is missing`),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_KMSEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
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

func TestAccAWSStorageGatewaySmbFileShare_KMSKeyArn(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"
	keyName := "aws_kms_key.test.0"
	keyUpdatedName := "aws_kms_key.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_KMSKeyArn(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", keyName, "arn"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_KMSKeyArn_Update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", keyUpdatedName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_KMSEncrypted(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "kms_encrypted", "false"),
				),
			},
		},
	})
}

func TestAccAWSStorageGatewaySmbFileShare_ObjectACL(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_ObjectACL(rName, storagegateway.ObjectACLPublicRead),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "object_acl", storagegateway.ObjectACLPublicRead),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_ObjectACL(rName, storagegateway.ObjectACLPublicReadWrite),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "object_acl", storagegateway.ObjectACLPublicReadWrite),
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

func TestAccAWSStorageGatewaySmbFileShare_ReadOnly(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_ReadOnly(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "read_only", "false"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_ReadOnly(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "read_only", "true"),
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

func TestAccAWSStorageGatewaySmbFileShare_RequesterPays(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_RequesterPays(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", "false"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_RequesterPays(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "requester_pays", "true"),
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

func TestAccAWSStorageGatewaySmbFileShare_ValidUserList(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_ValidUserList_Single(rName, "validuser1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", "1"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_ValidUserList_Multiple(rName, "validuser2", "validuser3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", "2"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_ValidUserList_Single(rName, "validuser4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "valid_user_list.#", "1"),
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

func TestAccAWSStorageGatewaySmbFileShare_smb_acl(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareSMBACLConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "smb_acl_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareSMBACLConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "smb_acl_enabled", "false"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareSMBACLConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "smb_acl_enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSStorageGatewaySmbFileShare_audit(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"
	logResourceName := "aws_cloudwatch_log_group.test"
	logResourceNameSecond := "aws_cloudwatch_log_group.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareAuditDestinationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttrPair(resourceName, "audit_destination_arn", logResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareAuditDestinationUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttrPair(resourceName, "audit_destination_arn", logResourceNameSecond, "arn"),
				),
			},
		},
	})
}

func TestAccAWSStorageGatewaySmbFileShare_cacheAttributes(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareCacheAttributesConfig(rName, 300),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.0.cache_stale_timeout_in_seconds", "300"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareCacheAttributesConfig(rName, 500),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.0.cache_stale_timeout_in_seconds", "500"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareCacheAttributesConfig(rName, 300),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_attributes.0.cache_stale_timeout_in_seconds", "300"),
				),
			},
		},
	})
}

func TestAccAWSStorageGatewaySmbFileShare_caseSensitivity(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareCaseSensitivityConfig(rName, "CaseSensitive"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "case_sensitivity", "CaseSensitive"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareCaseSensitivityConfig(rName, "ClientSpecified"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "case_sensitivity", "ClientSpecified"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareCaseSensitivityConfig(rName, "CaseSensitive"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "case_sensitivity", "CaseSensitive"),
				),
			},
		},
	})
}

func TestAccAWSStorageGatewaySmbFileShare_disappears(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_Authentication_GuestAccess(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsStorageGatewaySmbFileShare(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSStorageGatewaySmbFileShare_AdminUserList(t *testing.T) {
	var smbFileShare storagegateway.SMBFileShareInfo
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_smb_file_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewaySmbFileShareDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_AdminUserList_Single(rName, "adminuser1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "admin_user_list.#", "1"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_AdminUserList_Multiple(rName, "adminuser2", "adminuser3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "admin_user_list.#", "2"),
				),
			},
			{
				Config: testAccAWSStorageGatewaySmbFileShareConfig_AdminUserList_Single(rName, "adminuser4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName, &smbFileShare),
					resource.TestCheckResourceAttr(resourceName, "admin_user_list.#", "1"),
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

func testAccCheckAWSStorageGatewaySmbFileShareDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_storagegateway_smb_file_share" {
			continue
		}

		input := &storagegateway.DescribeSMBFileSharesInput{
			FileShareARNList: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeSMBFileShares(input)

		if err != nil {
			if isAWSErr(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified file share was not found.") {
				continue
			}
			return err
		}

		if output != nil && len(output.SMBFileShareInfoList) > 0 && output.SMBFileShareInfoList[0] != nil {
			return fmt.Errorf("Storage Gateway SMB File Share %q still exists", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAWSStorageGatewaySmbFileShareExists(resourceName string, smbFileShare *storagegateway.SMBFileShareInfo) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn
		input := &storagegateway.DescribeSMBFileSharesInput{
			FileShareARNList: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeSMBFileShares(input)

		if err != nil {
			return err
		}

		if output == nil || len(output.SMBFileShareInfoList) == 0 || output.SMBFileShareInfoList[0] == nil {
			return fmt.Errorf("Storage Gateway SMB File Share %q does not exist", rs.Primary.ID)
		}

		*smbFileShare = *output.SMBFileShareInfoList[0]

		return nil
	}
}

func testAccAWSStorageGateway_SmbFileShare_ActiveDirectoryBase(rName string) string {
	return testAccAWSStorageGatewayGatewayConfig_SmbActiveDirectorySettings(rName) + fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "storagegateway.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ],
      "Effect": "Allow"
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "test" {
  bucket        = %q
  force_destroy = true
}
`, rName, rName)
}

func testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName string) string {
	return testAccAWSStorageGatewayGatewayConfig_SmbGuestPassword(rName, "smbguestpassword") + fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "storagegateway.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:*"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ],
      "Effect": "Allow"
    }
  ]
}
POLICY
}

resource "aws_s3_bucket" "test" {
  bucket        = %q
  force_destroy = true
}
`, rName, rName)
}

func testAccAWSStorageGatewaySmbFileShareConfig_Authentication_ActiveDirectory(rName string) string {
	return testAccAWSStorageGateway_SmbFileShare_ActiveDirectoryBase(rName) + `
resource "aws_storagegateway_smb_file_share" "test" {
  authentication = "ActiveDirectory"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  role_arn       = aws_iam_role.test.arn
}
`
}

func testAccAWSStorageGatewaySmbFileShareConfig_Authentication_GuestAccess(rName string) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + `
resource "aws_storagegateway_smb_file_share" "test" {
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  role_arn       = aws_iam_role.test.arn
}
`
}

func testAccAWSStorageGatewaySmbFileShareConfig_DefaultStorageClass(rName, defaultStorageClass string) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication        = "GuestAccess"
  default_storage_class = %q
  gateway_arn           = aws_storagegateway_gateway.test.arn
  location_arn          = aws_s3_bucket.test.arn
  role_arn              = aws_iam_role.test.arn
}
`, defaultStorageClass)
}

func testAccAWSStorageGatewaySmbFileShareConfig_GuessMIMETypeEnabled(rName string, guessMimeTypeEnabled bool) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication          = "GuestAccess"
  gateway_arn             = aws_storagegateway_gateway.test.arn
  guess_mime_type_enabled = %t
  location_arn            = aws_s3_bucket.test.arn
  role_arn                = aws_iam_role.test.arn
}
`, guessMimeTypeEnabled)
}

func testAccAWSStorageGatewaySmbFileShareConfig_InvalidUserList_Single(rName, invalidUser1 string) string {
	return testAccAWSStorageGateway_SmbFileShare_ActiveDirectoryBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Must be ActiveDirectory
  authentication    = "ActiveDirectory"
  gateway_arn       = aws_storagegateway_gateway.test.arn
  invalid_user_list = [%q]
  location_arn      = aws_s3_bucket.test.arn
  role_arn          = aws_iam_role.test.arn
}
`, invalidUser1)
}

func testAccAWSStorageGatewaySmbFileShareConfig_InvalidUserList_Multiple(rName, invalidUser1, invalidUser2 string) string {
	return testAccAWSStorageGateway_SmbFileShare_ActiveDirectoryBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Must be ActiveDirectory
  authentication    = "ActiveDirectory"
  gateway_arn       = aws_storagegateway_gateway.test.arn
  invalid_user_list = [%q, %q]
  location_arn      = aws_s3_bucket.test.arn
  role_arn          = aws_iam_role.test.arn
}
`, invalidUser1, invalidUser2)
}

func testAccAWSStorageGatewaySmbFileShareConfig_KMSEncrypted(rName string, kmsEncrypted bool) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  kms_encrypted  = %t
  location_arn   = aws_s3_bucket.test.arn
  role_arn       = aws_iam_role.test.arn
}
`, kmsEncrypted)
}

func testAccAWSStorageGatewaySmbFileShareConfig_KMSKeyArn(rName string) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + `
resource "aws_kms_key" "test" {
  count = 2

  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
}

resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  kms_encrypted  = true
  kms_key_arn    = aws_kms_key.test[0].arn
  location_arn   = aws_s3_bucket.test.arn
  role_arn       = aws_iam_role.test.arn
}
`
}

func testAccAWSStorageGatewaySmbFileShareConfig_KMSKeyArn_Update(rName string) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + `
resource "aws_kms_key" "test" {
  count = 2

  deletion_window_in_days = 7
  description             = "Terraform Acceptance Testing"
}

resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  kms_encrypted  = true
  kms_key_arn    = aws_kms_key.test[1].arn
  location_arn   = aws_s3_bucket.test.arn
  role_arn       = aws_iam_role.test.arn
}
`
}

func testAccAWSStorageGatewaySmbFileShareConfig_ObjectACL(rName, objectACL string) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  object_acl     = %q
  role_arn       = aws_iam_role.test.arn
}
`, objectACL)
}

func testAccAWSStorageGatewaySmbFileShareConfig_ReadOnly(rName string, readOnly bool) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  read_only      = %t
  role_arn       = aws_iam_role.test.arn
}
`, readOnly)
}

func testAccAWSStorageGatewaySmbFileShareConfig_RequesterPays(rName string, requesterPays bool) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  requester_pays = %t
  role_arn       = aws_iam_role.test.arn
}
`, requesterPays)
}

func testAccAWSStorageGatewaySmbFileShareConfig_ValidUserList_Single(rName, validUser1 string) string {
	return testAccAWSStorageGateway_SmbFileShare_ActiveDirectoryBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Must be ActiveDirectory
  authentication  = "ActiveDirectory"
  gateway_arn     = aws_storagegateway_gateway.test.arn
  location_arn    = aws_s3_bucket.test.arn
  role_arn        = aws_iam_role.test.arn
  valid_user_list = [%q]
}
`, validUser1)
}

func testAccAWSStorageGatewaySmbFileShareConfig_ValidUserList_Multiple(rName, validUser1, validUser2 string) string {
	return testAccAWSStorageGateway_SmbFileShare_ActiveDirectoryBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Must be ActiveDirectory
  authentication  = "ActiveDirectory"
  gateway_arn     = aws_storagegateway_gateway.test.arn
  location_arn    = aws_s3_bucket.test.arn
  role_arn        = aws_iam_role.test.arn
  valid_user_list = [%q, %q]
}
`, validUser1, validUser2)
}

func testAccAWSStorageGatewaySmbFileShareConfig_AdminUserList_Single(rName, adminUser1 string) string {
	return testAccAWSStorageGateway_SmbFileShare_ActiveDirectoryBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Must be ActiveDirectory
  authentication  = "ActiveDirectory"
  gateway_arn     = aws_storagegateway_gateway.test.arn
  location_arn    = aws_s3_bucket.test.arn
  role_arn        = aws_iam_role.test.arn
  admin_user_list = [%q]
}
`, adminUser1)
}

func testAccAWSStorageGatewaySmbFileShareConfig_AdminUserList_Multiple(rName, adminUser1, adminUser2 string) string {
	return testAccAWSStorageGateway_SmbFileShare_ActiveDirectoryBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Must be ActiveDirectory
  authentication  = "ActiveDirectory"
  gateway_arn     = aws_storagegateway_gateway.test.arn
  location_arn    = aws_s3_bucket.test.arn
  role_arn        = aws_iam_role.test.arn
  admin_user_list = [%q, %q]
}
`, adminUser1, adminUser2)
}

func testAccAWSStorageGatewaySmbFileShareConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  role_arn       = aws_iam_role.test.arn

  tags = {
    %q = %q
  }
}
`, tagKey1, tagValue1)
}

func testAccAWSStorageGatewaySmbFileShareConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  role_arn       = aws_iam_role.test.arn

  tags = {
    %q = %q
    %q = %q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSStorageGatewaySmbFileShareSMBACLConfig(rName string, enabled bool) string {
	return testAccAWSStorageGateway_SmbFileShare_ActiveDirectoryBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  authentication  = "ActiveDirectory"
  gateway_arn     = aws_storagegateway_gateway.test.arn
  location_arn    = aws_s3_bucket.test.arn
  role_arn        = aws_iam_role.test.arn
  smb_acl_enabled = %[1]t
}
`, enabled)
}

func testAccAWSStorageGatewaySmbFileShareAuditDestinationConfig(rName string) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication        = "GuestAccess"
  gateway_arn           = aws_storagegateway_gateway.test.arn
  location_arn          = aws_s3_bucket.test.arn
  role_arn              = aws_iam_role.test.arn
  audit_destination_arn = aws_cloudwatch_log_group.test.arn
}
`, rName)
}

func testAccAWSStorageGatewaySmbFileShareAuditDestinationUpdatedConfig(rName string) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_group" "test2" {
  name = "%[1]s-updated"
}

resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication        = "GuestAccess"
  gateway_arn           = aws_storagegateway_gateway.test.arn
  location_arn          = aws_s3_bucket.test.arn
  role_arn              = aws_iam_role.test.arn
  audit_destination_arn = aws_cloudwatch_log_group.test2.arn
}
`, rName)
}

func testAccAWSStorageGatewaySmbFileShareCacheAttributesConfig(rName string, timeout int) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication = "GuestAccess"
  gateway_arn    = aws_storagegateway_gateway.test.arn
  location_arn   = aws_s3_bucket.test.arn
  role_arn       = aws_iam_role.test.arn

  cache_attributes {
    cache_stale_timeout_in_seconds = %[1]d
  }
}
`, timeout)
}

func testAccAWSStorageGatewaySmbFileShareCaseSensitivityConfig(rName, option string) string {
	return testAccAWSStorageGateway_SmbFileShare_GuestAccessBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_smb_file_share" "test" {
  # Use GuestAccess to simplify testing
  authentication   = "GuestAccess"
  gateway_arn      = aws_storagegateway_gateway.test.arn
  location_arn     = aws_s3_bucket.test.arn
  role_arn         = aws_iam_role.test.arn
  case_sensitivity = %[1]q
}
`, option)
}
