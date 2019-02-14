package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_iam_service_linked_role", &resource.Sweeper{
		Name: "aws_iam_service_linked_role",
		F:    testSweepIamServiceLinkedRoles,
	})
}

func testSweepIamServiceLinkedRoles(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).iamconn

	input := &iam.ListRolesInput{
		PathPrefix: aws.String("/aws-service-role/"),
	}
	customSuffixRegex := regexp.MustCompile(`_tf-acc-test-\d+$`)
	err = conn.ListRolesPages(input, func(page *iam.ListRolesOutput, lastPage bool) bool {
		if len(page.Roles) == 0 {
			log.Printf("[INFO] No IAM Service Roles to sweep")
			return true
		}
		for _, role := range page.Roles {
			roleName := aws.StringValue(role.RoleName)

			if !customSuffixRegex.MatchString(roleName) {
				log.Printf("[INFO] Skipping IAM Service Role: %s", roleName)
				continue
			}

			log.Printf("[INFO] Deleting IAM Service Role: %s", roleName)
			deletionTaskID, err := deleteIamServiceLinkedRole(conn, roleName)
			if err != nil {
				log.Printf("[ERROR] Failed to delete IAM Service Role %s: %s", roleName, err)
				continue
			}
			if deletionTaskID == "" {
				continue
			}

			log.Printf("[INFO] Waiting for deletion of IAM Service Role: %s", roleName)
			err = deleteIamServiceLinkedRoleWaiter(conn, deletionTaskID)
			if err != nil {
				log.Printf("[ERROR] Failed to wait for deletion of IAM Service Role %s: %s", roleName, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping IAM Service Role sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving IAM Service Roles: %s", err)
	}

	return nil
}

func TestDecodeIamServiceLinkedRoleID(t *testing.T) {
	var testCases = []struct {
		Input        string
		ServiceName  string
		RoleName     string
		CustomSuffix string
		ErrCount     int
	}{
		{
			Input:    "not-arn",
			ErrCount: 1,
		},
		{
			Input:    "arn:aws:iam::123456789012:role/not-service-linked-role",
			ErrCount: 1,
		},
		{
			Input:        "arn:aws:iam::123456789012:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling",
			ServiceName:  "autoscaling.amazonaws.com",
			RoleName:     "AWSServiceRoleForAutoScaling",
			CustomSuffix: "",
			ErrCount:     0,
		},
		{
			Input:        "arn:aws:iam::123456789012:role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling_custom-suffix",
			ServiceName:  "autoscaling.amazonaws.com",
			RoleName:     "AWSServiceRoleForAutoScaling_custom-suffix",
			CustomSuffix: "custom-suffix",
			ErrCount:     0,
		},
	}

	for _, tc := range testCases {
		serviceName, roleName, customSuffix, err := decodeIamServiceLinkedRoleID(tc.Input)
		if tc.ErrCount == 0 && err != nil {
			t.Fatalf("expected %q not to trigger an error, received: %s", tc.Input, err)
		}
		if tc.ErrCount > 0 && err == nil {
			t.Fatalf("expected %q to trigger an error", tc.Input)
		}
		if serviceName != tc.ServiceName {
			t.Fatalf("expected service name %q to be %q", serviceName, tc.ServiceName)
		}
		if roleName != tc.RoleName {
			t.Fatalf("expected role name %q to be %q", roleName, tc.RoleName)
		}
		if customSuffix != tc.CustomSuffix {
			t.Fatalf("expected custom suffix %q to be %q", customSuffix, tc.CustomSuffix)
		}
	}
}

func TestAccAWSIAMServiceLinkedRole_basic(t *testing.T) {
	resourceName := "aws_iam_service_linked_role.test"
	awsServiceName := "elasticbeanstalk.amazonaws.com"
	name := "AWSServiceRoleForElasticBeanstalk"
	path := fmt.Sprintf("/aws-service-role/%s/", awsServiceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIAMServiceLinkedRoleDestroy,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Remove existing if possible
					conn := testAccProvider.Meta().(*AWSClient).iamconn
					deletionID, err := deleteIamServiceLinkedRole(conn, name)
					if err != nil {
						t.Fatalf("Error deleting service-linked role %s: %s", name, err)
					}
					if deletionID == "" {
						return
					}

					err = deleteIamServiceLinkedRoleWaiter(conn, deletionID)
					if err != nil {
						t.Fatalf("Error waiting for role (%s) to be deleted: %s", name, err)
					}
				},
				Config: testAccAWSIAMServiceLinkedRoleConfig(awsServiceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIAMServiceLinkedRoleExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:iam::[^:]+:role%s%s$", path, name))),
					resource.TestCheckResourceAttr(resourceName, "aws_service_name", awsServiceName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "path", path),
					resource.TestCheckResourceAttrSet(resourceName, "unique_id"),
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

func TestAccAWSIAMServiceLinkedRole_CustomSuffix(t *testing.T) {
	resourceName := "aws_iam_service_linked_role.test"
	awsServiceName := "autoscaling.amazonaws.com"
	customSuffix := acctest.RandomWithPrefix("tf-acc-test")
	name := fmt.Sprintf("AWSServiceRoleForAutoScaling_%s", customSuffix)
	path := fmt.Sprintf("/aws-service-role/%s/", awsServiceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIAMServiceLinkedRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMServiceLinkedRoleConfig_CustomSuffix(awsServiceName, customSuffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIAMServiceLinkedRoleExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:iam::[^:]+:role%s%s$", path, name))),
					resource.TestCheckResourceAttr(resourceName, "name", name),
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

func TestAccAWSIAMServiceLinkedRole_Description(t *testing.T) {
	resourceName := "aws_iam_service_linked_role.test"
	awsServiceName := "autoscaling.amazonaws.com"
	customSuffix := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIAMServiceLinkedRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIAMServiceLinkedRoleConfig_Description(awsServiceName, customSuffix, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIAMServiceLinkedRoleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				Config: testAccAWSIAMServiceLinkedRoleConfig_Description(awsServiceName, customSuffix, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIAMServiceLinkedRoleExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
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

func testAccCheckAWSIAMServiceLinkedRoleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_service_linked_role" {
			continue
		}

		_, roleName, _, err := decodeIamServiceLinkedRoleID(rs.Primary.ID)
		if err != nil {
			return err
		}

		params := &iam.GetRoleInput{
			RoleName: aws.String(roleName),
		}

		_, err = conn.GetRole(params)

		if err == nil {
			return fmt.Errorf("Service-Linked Role still exists: %q", rs.Primary.ID)
		}

		if !isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
			return err
		}
	}

	return nil

}

func testAccCheckAWSIAMServiceLinkedRoleExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).iamconn
		_, roleName, _, err := decodeIamServiceLinkedRoleID(rs.Primary.ID)
		if err != nil {
			return err
		}

		params := &iam.GetRoleInput{
			RoleName: aws.String(roleName),
		}

		_, err = conn.GetRole(params)

		if err != nil {
			if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
				return fmt.Errorf("Service-Linked Role doesn't exists: %q", rs.Primary.ID)
			}
			return err
		}

		return nil
	}
}

func testAccAWSIAMServiceLinkedRoleConfig(awsServiceName string) string {
	return fmt.Sprintf(`
resource "aws_iam_service_linked_role" "test" {
  aws_service_name = "%s"
}
`, awsServiceName)
}

func testAccAWSIAMServiceLinkedRoleConfig_CustomSuffix(awsServiceName, customSuffix string) string {
	return fmt.Sprintf(`
resource "aws_iam_service_linked_role" "test" {
  aws_service_name = "%s"
  custom_suffix    = "%s"
}
`, awsServiceName, customSuffix)
}

func testAccAWSIAMServiceLinkedRoleConfig_Description(awsServiceName, customSuffix, description string) string {
	return fmt.Sprintf(`
resource "aws_iam_service_linked_role" "test" {
  aws_service_name = "%s"
  custom_suffix    = "%s"
  description      = "%s"
}
`, awsServiceName, customSuffix, description)
}
