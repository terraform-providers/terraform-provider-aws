package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ssoadmin/finder"
)

func TestAccAWSSSOAdminManagedPolicyAttachment_basic(t *testing.T) {
	resourceName := "aws_ssoadmin_managed_policy_attachment.test"
	permissionSetResourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAdminManagedPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSSOAdminManagedPolicyAttachmentBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOAdminManagedPolicyAttachmentExists(resourceName),
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "managed_policy_arn", regexp.MustCompile(`policy/AlexaForBusinessDeviceSetup`)),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_name", "AlexaForBusinessDeviceSetup"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_arn", permissionSetResourceName, "instance_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "permission_set_arn", permissionSetResourceName, "arn"),
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

func TestAccAWSSSOAdminManagedPolicyAttachment_forceNew(t *testing.T) {
	resourceName := "aws_ssoadmin_managed_policy_attachment.test"
	permissionSetResourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAdminManagedPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSSOAdminManagedPolicyAttachmentBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOAdminManagedPolicyAttachmentExists(resourceName),
				),
			},
			{
				Config: testAccSSOAdminManagedPolicyAttachmentConfig_forceNew(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOAdminManagedPolicyAttachmentExists(resourceName),
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(resourceName, "managed_policy_arn", regexp.MustCompile(`policy/AmazonCognitoReadOnly`)),
					resource.TestCheckResourceAttr(resourceName, "managed_policy_name", "AmazonCognitoReadOnly"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_arn", permissionSetResourceName, "instance_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "permission_set_arn", permissionSetResourceName, "arn"),
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

func TestAccAWSSSOAdminManagedPolicyAttachment_disappears(t *testing.T) {
	resourceName := "aws_ssoadmin_managed_policy_attachment.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAdminManagedPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSSOAdminManagedPolicyAttachmentBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOAdminManagedPolicyAttachmentExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSsoAdminManagedPolicyAttachment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSSOAdminManagedPolicyAttachment_disappears_permissionSet(t *testing.T) {
	resourceName := "aws_ssoadmin_managed_policy_attachment.test"
	permissionSetResourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAdminManagedPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSSOAdminManagedPolicyAttachmentBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOAdminManagedPolicyAttachmentExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSsoAdminPermissionSet(), permissionSetResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSSOAdminManagedPolicyAttachment_multipleManagedPolicies(t *testing.T) {
	resourceName := "aws_ssoadmin_managed_policy_attachment.test"
	otherResourceName := "aws_ssoadmin_managed_policy_attachment.other"
	permissionSetResourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAdminManagedPolicyAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSSOAdminManagedPolicyAttachmentBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOAdminManagedPolicyAttachmentExists(resourceName),
				),
			},
			{
				Config: testAccSSOAdminManagedPolicyAttachmentConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOAdminManagedPolicyAttachmentExists(resourceName),
					testAccCheckAWSSSOAdminManagedPolicyAttachmentExists(otherResourceName),
					//lintignore:AWSAT001
					resource.TestMatchResourceAttr(otherResourceName, "managed_policy_arn", regexp.MustCompile(`policy/AmazonDynamoDBReadOnlyAccess`)),
					resource.TestCheckResourceAttr(otherResourceName, "managed_policy_name", "AmazonDynamoDBReadOnlyAccess"),
					resource.TestCheckResourceAttrPair(otherResourceName, "instance_arn", permissionSetResourceName, "instance_arn"),
					resource.TestCheckResourceAttrPair(otherResourceName, "permission_set_arn", permissionSetResourceName, "arn"),
				),
			},
			{
				ResourceName:      otherResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSSSOAdminManagedPolicyAttachmentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ssoadminconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssoadmin_managed_policy_attachment" {
			continue
		}

		managedPolicyArn, permissionSetArn, instanceArn, err := parseSsoAdminManagedPolicyAttachmentID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error parsing SSO Managed Policy Attachment ID (%s): %w", rs.Primary.ID, err)
		}

		policy, err := finder.ManagedPolicy(conn, managedPolicyArn, permissionSetArn, instanceArn)

		if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if policy == nil {
			continue
		}

		return fmt.Errorf("Managed Policy (%s) for SSO Permission Set (%s) still exists", managedPolicyArn, permissionSetArn)

	}

	return nil
}

func testAccCheckAWSSSOAdminManagedPolicyAttachmentExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		managedPolicyArn, permissionSetArn, instanceArn, err := parseSsoAdminManagedPolicyAttachmentID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing SSO Managed Policy Attachment ID (%s): %w", rs.Primary.ID, err)
		}

		conn := testAccProvider.Meta().(*AWSClient).ssoadminconn

		policy, err := finder.ManagedPolicy(conn, managedPolicyArn, permissionSetArn, instanceArn)

		if err != nil {
			return err
		}

		if policy == nil {
			return fmt.Errorf("Managed Policy (%s) for SSO Permission Set (%s) not found", managedPolicyArn, permissionSetArn)
		}

		return nil
	}
}

func testAccSSOAdminManagedPolicyAttachmentBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`, rName)
}

func testAccSSOAdminManagedPolicyAttachmentBasicConfig(rName string) string {
	return composeConfig(
		testAccSSOAdminManagedPolicyAttachmentBaseConfig(rName),
		`
resource "aws_ssoadmin_managed_policy_attachment" "test" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  managed_policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AlexaForBusinessDeviceSetup"
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}
`)
}

func testAccSSOAdminManagedPolicyAttachmentConfig_forceNew(rName string) string {
	return composeConfig(
		testAccSSOAdminManagedPolicyAttachmentBaseConfig(rName),
		`
resource "aws_ssoadmin_managed_policy_attachment" "test" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  managed_policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonCognitoReadOnly"
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}
`)
}

func testAccSSOAdminManagedPolicyAttachmentConfig_multiple(rName string) string {
	return composeConfig(
		testAccSSOAdminManagedPolicyAttachmentBasicConfig(rName),
		`
resource "aws_ssoadmin_managed_policy_attachment" "other" {
  instance_arn       = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  managed_policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonDynamoDBReadOnlyAccess"
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}
`)
}
