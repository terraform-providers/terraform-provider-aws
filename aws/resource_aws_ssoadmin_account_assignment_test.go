package aws

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ssoadmin/finder"
)

func init() {
	resource.AddTestSweepers("aws_ssoadmin_account_assignment", &resource.Sweeper{
		Name: "aws_ssoadmin_account_assignment",
		F:    testSweepSsoAdminAccountAssignments,
	})
}

func testSweepSsoAdminAccountAssignments(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).ssoadminconn
	var sweeperErrs *multierror.Error

	// Need to Read the SSO Instance first; assumes the first instance returned
	// is where the permission sets exist as AWS SSO currently supports only 1 instance
	ds := dataSourceAwsSsoAdminInstances()
	dsData := ds.Data(nil)

	err = ds.Read(dsData, client)

	if testSweepSkipResourceError(err) {
		log.Printf("[WARN] Skipping SSO Account Assignment sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return err
	}

	instanceArn := dsData.Get("arns").(*schema.Set).List()[0].(string)

	// To sweep account assignments, we need to first determine which Permission Sets
	// are available and then search for their respective assignments
	input := &ssoadmin.ListPermissionSetsInput{
		InstanceArn: aws.String(instanceArn),
	}

	err = conn.ListPermissionSetsPages(input, func(page *ssoadmin.ListPermissionSetsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, permissionSet := range page.PermissionSets {
			if permissionSet == nil {
				continue
			}

			permissionSetArn := aws.StringValue(permissionSet)

			input := &ssoadmin.ListAccountAssignmentsInput{
				AccountId:        aws.String(client.(*AWSClient).accountid),
				InstanceArn:      aws.String(instanceArn),
				PermissionSetArn: permissionSet,
			}

			err := conn.ListAccountAssignmentsPages(input, func(page *ssoadmin.ListAccountAssignmentsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, a := range page.AccountAssignments {
					if a == nil {
						continue
					}

					principalID := aws.StringValue(a.PrincipalId)
					principalType := aws.StringValue(a.PrincipalType)
					targetID := aws.StringValue(a.AccountId)
					targetType := ssoadmin.TargetTypeAwsAccount // only valid value currently accepted by API

					r := resourceAwsSsoAdminAccountAssignment()
					d := r.Data(nil)
					d.SetId(fmt.Sprintf("%s,%s,%s,%s,%s,%s", principalID, principalType, targetID, targetType, permissionSetArn, instanceArn))

					err = r.Delete(d, client)

					if err != nil {
						log.Printf("[ERROR] %s", err)
						sweeperErrs = multierror.Append(sweeperErrs, err)
						continue
					}
				}

				return !lastPage
			})

			if testSweepSkipSweepError(err) {
				log.Printf("[WARN] Skipping SSO Account Assignment sweep (PermissionSet %s) for %s: %s", permissionSetArn, region, err)
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving SSO Account Assignments for Permission Set (%s): %w", permissionSetArn, err))
			}
		}

		return !isLast
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SSO Account Assignment sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving SSO Permission Sets for Account Assignment sweep: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSSSOAdminAccountAssignment_Basic_Group(t *testing.T) {
	resourceName := "aws_ssoadmin_account_assignment.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	groupName := os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSSOAdminInstances(t)
			testAccPreCheckAWSIdentityStoreGroupName(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAdminAccountAssignmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSOAdminAccountAssignmentBasicGroupConfig(groupName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOAdminAccountAssignmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_type", "AWS_ACCOUNT"),
					resource.TestCheckResourceAttr(resourceName, "principal_type", "GROUP"),
					resource.TestMatchResourceAttr(resourceName, "principal_id", regexp.MustCompile("^([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}")),
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

func TestAccAWSSSOAdminAccountAssignment_Basic_User(t *testing.T) {
	resourceName := "aws_ssoadmin_account_assignment.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	userName := os.Getenv("AWS_IDENTITY_STORE_USER_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSSOAdminInstances(t)
			testAccPreCheckAWSIdentityStoreUserName(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAdminAccountAssignmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSOAdminAccountAssignmentBasicUserConfig(userName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOAdminAccountAssignmentExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "target_type", "AWS_ACCOUNT"),
					resource.TestCheckResourceAttr(resourceName, "principal_type", "USER"),
					resource.TestMatchResourceAttr(resourceName, "principal_id", regexp.MustCompile("^([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}")),
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

func TestAccAWSSSOAdminAccountAssignment_Disappears(t *testing.T) {
	resourceName := "aws_ssoadmin_account_assignment.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	groupName := os.Getenv("AWS_IDENTITY_STORE_GROUP_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSSOAdminInstances(t)
			testAccPreCheckAWSIdentityStoreGroupName(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAdminAccountAssignmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSOAdminAccountAssignmentBasicGroupConfig(groupName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSOAdminAccountAssignmentExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSsoAdminAccountAssignment(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})

}

func testAccCheckAWSSSOAdminAccountAssignmentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ssoadminconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssoadmin_account_assignment" {
			continue
		}

		idParts, err := parseSsoAdminAccountAssignmentID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing SSO Account Assignment ID (%s): %w", rs.Primary.ID, err)
		}

		principalID := idParts[0]
		principalType := idParts[1]
		targetID := idParts[2]
		permissionSetArn := idParts[4]
		instanceArn := idParts[5]

		accountAssignment, err := finder.AccountAssignment(conn, principalID, principalType, targetID, permissionSetArn, instanceArn)

		if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading SSO Account Assignment for Principal (%s): %w", principalID, err)
		}

		if accountAssignment != nil {
			return fmt.Errorf("SSO Account Assignment for Principal (%s) still exists", principalID)
		}
	}

	return nil
}

func testAccCheckAWSSSOAdminAccountAssignmentExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).ssoadminconn

		idParts, err := parseSsoAdminAccountAssignmentID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing SSO Account Assignment ID (%s): %w", rs.Primary.ID, err)
		}

		principalID := idParts[0]
		principalType := idParts[1]
		targetID := idParts[2]
		permissionSetArn := idParts[4]
		instanceArn := idParts[5]

		accountAssignment, err := finder.AccountAssignment(conn, principalID, principalType, targetID, permissionSetArn, instanceArn)

		if err != nil {
			return err
		}

		if accountAssignment == nil {
			return fmt.Errorf("Account Assignment for Principal (%s) not found", principalID)
		}

		return nil
	}
}

func testAccAWSSSOAdminAccountAssignmentBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

data "aws_caller_identity" "current" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`, rName)
}

func testAccAWSSSOAdminAccountAssignmentBasicGroupConfig(groupName, rName string) string {
	return composeConfig(
		testAccAWSSSOAdminAccountAssignmentBaseConfig(rName),
		fmt.Sprintf(`
data "aws_identitystore_group" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  filter {
    attribute_path  = "DisplayName"
    attribute_value = %q
  }
}

resource "aws_ssoadmin_account_assignment" "test" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
  target_type        = "AWS_ACCOUNT"
  target_id          = data.aws_caller_identity.current.account_id
  principal_type     = "GROUP"
  principal_id       = data.aws_identitystore_group.test.group_id
}
`, groupName))
}

func testAccAWSSSOAdminAccountAssignmentBasicUserConfig(userName, rName string) string {
	return composeConfig(
		testAccAWSSSOAdminAccountAssignmentBaseConfig(rName),
		fmt.Sprintf(`
data "aws_identitystore_user" "test" {
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
  filter {
    attribute_path  = "UserName"
    attribute_value = %q
  }
}

resource "aws_ssoadmin_account_assignment" "test" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
  target_type        = "AWS_ACCOUNT"
  target_id          = data.aws_caller_identity.current.account_id
  principal_type     = "USER"
  principal_id       = data.aws_identitystore_user.test.user_id
}
`, userName))
}
