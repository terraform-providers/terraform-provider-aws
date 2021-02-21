package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_ssoadmin_permission_set", &resource.Sweeper{
		Name: "aws_ssoadmin_permission_set",
		F:    testSweepSsoAdminPermissionSets,
		Dependencies: []string{
			"aws_ssoadmin_account_assignment",
		},
	})
}

func testSweepSsoAdminPermissionSets(region string) error {
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
		log.Printf("[WARN] Skipping SSO Permission Set sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return err
	}

	instanceArn := dsData.Get("arns").(*schema.Set).List()[0].(string)

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

			arn := aws.StringValue(permissionSet)

			log.Printf("[INFO] Deleting SSO Permission Set: %s", arn)

			r := resourceAwsSsoAdminPermissionSet()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s,%s", arn, instanceArn))

			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !isLast
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping SSO Permission Set sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving SSO Permission Set: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSSSOAdminPermissionSet_basic(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAdminPermissionSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSOAdminPermissionSetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "session_duration", "PT1H"),
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

func TestAccAWSSSOAdminPermissionSet_tags(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAdminPermissionSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSOAdminPermissionSetConfigTagsSingle(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSOAdminPermissionSetExists(resourceName),
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
				Config: testAccAWSSSOAdminPermissionSetConfigTagsMultiple(rName, "key1", "updatedvalue1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "updatedvalue1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSOAdminPermissionSetConfigTagsSingle(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

func TestAccAWSSSOAdminPermissionSet_updateDescription(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAdminPermissionSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSOAdminPermissionSetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				Config: testAccAWSSSOAdminPermissionSetUpdateDescriptionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
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

func TestAccAWSSSOAdminPermissionSet_updateRelayState(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAdminPermissionSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSOAdminPermissionSetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "relay_state", ""),
				),
			},
			{
				Config: testAccAWSSSOAdminPermissionSetUpdateRelayStateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "relay_state", "https://example.com"),
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

func TestAccAWSSSOAdminPermissionSet_updateSessionDuration(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAdminPermissionSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSOAdminPermissionSetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSOAdminPermissionSetExists(resourceName),
				),
			},
			{
				Config: testAccAWSSSOAdminPermissionSetUpdateSessionDurationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "session_duration", "PT2H"),
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

// TestAccAWSSSOAdminPermissionSet_relayState_updateSessionDuration validates
// the resource's unchanged values (primarily relay_state) after updating the session_duration argument
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17411
func TestAccAWSSSOAdminPermissionSet_relayState_updateSessionDuration(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAdminPermissionSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSOAdminPermissionSetRelayStateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "relay_state", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "session_duration", "PT1H"),
				),
			},
			{
				Config: testAccAWSSSOAdminPermissionSetRelayStateConfig_updateSessionDuration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSOAdminPermissionSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", rName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "relay_state", "https://example.com"),
					resource.TestCheckResourceAttr(resourceName, "session_duration", "PT2H"),
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

func TestAccAWSSSOAdminPermissionSet_mixedPolicyAttachments(t *testing.T) {
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSOAdminPermissionSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSOAdminPermissionSetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSOAdminPermissionSetExists(resourceName),
				),
			},
			{
				Config: testAccAWSSSOAdminPermissionSetMixedPolicyAttachmentsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSOAdminPermissionSetExists(resourceName),
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

func testAccCheckAWSSSOAdminPermissionSetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ssoadminconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssoadmin_permission_set" {
			continue
		}

		arn, instanceArn, err := parseSsoAdminResourceID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing SSO Permission Set ID (%s): %w", rs.Primary.ID, err)
		}

		input := &ssoadmin.DescribePermissionSetInput{
			InstanceArn:      aws.String(instanceArn),
			PermissionSetArn: aws.String(arn),
		}

		_, err = conn.DescribePermissionSet(input)

		if tfawserr.ErrCodeEquals(err, ssoadmin.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("SSO Permission Set (%s) still exists", arn)
	}

	return nil
}

func testAccCheckAWSSOAdminPermissionSetExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Resource (%s) ID not set", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).ssoadminconn

		arn, instanceArn, err := parseSsoAdminResourceID(rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error parsing SSO Permission Set ID (%s): %w", rs.Primary.ID, err)
		}

		_, err = conn.DescribePermissionSet(&ssoadmin.DescribePermissionSetInput{
			InstanceArn:      aws.String(instanceArn),
			PermissionSetArn: aws.String(arn),
		})

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSSSOAdminPermissionSetBasicConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`, rName)
}

func testAccAWSSSOAdminPermissionSetUpdateDescriptionConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  description  = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}
`, rName)
}

func testAccAWSSSOAdminPermissionSetUpdateRelayStateConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  relay_state  = "https://example.com"
}
`, rName)
}

func testAccAWSSSOAdminPermissionSetUpdateSessionDurationConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name             = %[1]q
  instance_arn     = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  session_duration = "PT2H"
}
`, rName)
}

func testAccAWSSSOAdminPermissionSetRelayStateConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  description      = %[1]q
  name             = %[1]q
  instance_arn     = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  relay_state      = "https://example.com"
  session_duration = "PT1H"
}
`, rName)
}

func testAccAWSSSOAdminPermissionSetRelayStateConfig_updateSessionDuration(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  description      = %[1]q
  name             = %[1]q
  instance_arn     = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  relay_state      = "https://example.com"
  session_duration = "PT2H"
}
`, rName)
}

func testAccAWSSSOAdminPermissionSetConfigTagsSingle(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSSSOAdminPermissionSetConfigTagsMultiple(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSSSOAdminPermissionSetMixedPolicyAttachmentsConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
}

resource "aws_ssoadmin_managed_policy_attachment" "test" {
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  managed_policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AlexaForBusinessDeviceSetup"
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}

data "aws_iam_policy_document" "test" {
  statement {
    sid = "1"

    actions = [
      "s3:ListAllMyBuckets",
    ]

    resources = [
      "arn:${data.aws_partition.current.partition}:s3:::*",
    ]
  }
}
resource "aws_ssoadmin_permission_set_inline_policy" "test" {
  inline_policy      = data.aws_iam_policy_document.test.json
  instance_arn       = aws_ssoadmin_permission_set.test.instance_arn
  permission_set_arn = aws_ssoadmin_permission_set.test.arn
}
`, rName)
}
