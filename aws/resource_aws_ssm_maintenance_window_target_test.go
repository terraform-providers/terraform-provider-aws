package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSSMMaintenanceWindowTarget_basic(t *testing.T) {
	var maint ssm.MaintenanceWindowTarget
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_maintenance_window_target.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTargetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTargetExists(resourceName, &maint),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.key", "tag:Name2"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.1", "acceptance_test2"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "This resource is for test purpose only"),
					resource.TestCheckResourceAttr(resourceName, "resource_type", ssm.MaintenanceWindowResourceTypeInstance),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSSSMMaintenanceWindowTargetImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindowTarget_noNameOrDescription(t *testing.T) {
	var maint ssm.MaintenanceWindowTarget
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_maintenance_window_target.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTargetNoNameOrDescriptionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTargetExists(resourceName, &maint),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.key", "tag:Name2"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.1", "acceptance_test2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSSSMMaintenanceWindowTargetImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindowTarget_validation(t *testing.T) {
	name := acctest.RandString(10)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSSSMMaintenanceWindowTargetBasicConfigWithTarget(name, "Bäd Name!@#$%^", "good description"),
				ExpectError: regexp.MustCompile(`Only alphanumeric characters, hyphens, dots & underscores allowed`),
			},
			{
				Config:      testAccAWSSSMMaintenanceWindowTargetBasicConfigWithTarget(name, "goodname", "bd"),
				ExpectError: regexp.MustCompile(`expected length of [\w]+ to be in the range \(3 - 128\), got [\w]+`),
			},
			{
				Config:      testAccAWSSSMMaintenanceWindowTargetBasicConfigWithTarget(name, "goodname", "This description is tooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooo long"),
				ExpectError: regexp.MustCompile(`expected length of [\w]+ to be in the range \(3 - 128\), got [\w]+`),
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindowTarget_update(t *testing.T) {
	var maint ssm.MaintenanceWindowTarget
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_maintenance_window_target.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTargetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTargetExists(resourceName, &maint),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.key", "tag:Name2"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.1", "acceptance_test2"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "This resource is for test purpose only"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSSSMMaintenanceWindowTargetImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMMaintenanceWindowTargetBasicConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTargetExists(resourceName, &maint),
					resource.TestCheckResourceAttr(resourceName, "owner_information", "something"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "acceptance_test"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.key", "tag:Updated"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.0", "new-value"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "This resource is for test purpose only - updated"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSSSMMaintenanceWindowTargetImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindowTarget_resourceGroup(t *testing.T) {
	var maint ssm.MaintenanceWindowTarget
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_maintenance_window_target.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTargetBasicResourceGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTargetExists(resourceName, &maint),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "resource-groups:ResourceTypeFilters"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "AWS::EC2::Instance"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.key", "resource-groups:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.0", "resource-group-name"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "This resource is for test purpose only"),
					resource.TestCheckResourceAttr(resourceName, "resource_type", ssm.MaintenanceWindowResourceTypeResourceGroup),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSSSMMaintenanceWindowTargetImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindowTarget_disappears(t *testing.T) {
	var maint ssm.MaintenanceWindowTarget
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_maintenance_window_target.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowTargetBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowTargetExists(resourceName, &maint),
					testAccCheckAWSSSMMaintenanceWindowTargetDisappears(&maint),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSSMMaintenanceWindowTargetExists(n string, mWindTarget *ssm.MaintenanceWindowTarget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Maintenance Window Target Window ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ssmconn

		resp, err := conn.DescribeMaintenanceWindowTargets(&ssm.DescribeMaintenanceWindowTargetsInput{
			WindowId: aws.String(rs.Primary.Attributes["window_id"]),
			Filters: []*ssm.MaintenanceWindowFilter{
				{
					Key:    aws.String("WindowTargetId"),
					Values: []*string{aws.String(rs.Primary.ID)},
				},
			},
		})
		if err != nil {
			return err
		}

		for _, i := range resp.Targets {
			if *i.WindowTargetId == rs.Primary.ID {
				*mWindTarget = *resp.Targets[0]
				return nil
			}
		}

		return fmt.Errorf("No AWS SSM Maintenance window target found")
	}
}

func testAccCheckAWSSSMMaintenanceWindowTargetDisappears(mWindTarget *ssm.MaintenanceWindowTarget) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ssmconn
		input := &ssm.DeregisterTargetFromMaintenanceWindowInput{
			WindowId:       mWindTarget.WindowId,
			WindowTargetId: mWindTarget.WindowTargetId,
		}

		_, err := conn.DeregisterTargetFromMaintenanceWindow(input)

		return err
	}
}

func testAccCheckAWSSSMMaintenanceWindowTargetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ssmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_maintenance_window_target" {
			continue
		}

		out, err := conn.DescribeMaintenanceWindowTargets(&ssm.DescribeMaintenanceWindowTargetsInput{
			WindowId: aws.String(rs.Primary.Attributes["window_id"]),
			Filters: []*ssm.MaintenanceWindowFilter{
				{
					Key:    aws.String("WindowTargetId"),
					Values: []*string{aws.String(rs.Primary.ID)},
				},
			},
		})

		if err != nil {
			// Verify the error is what we want
			if isAWSErr(err, ssm.ErrCodeDoesNotExistException, "") {
				continue
			}
			return err
		}

		if len(out.Targets) > 0 {
			return fmt.Errorf("Expected AWS SSM Maintenance Target to be gone, but was still found")
		}

		return nil
	}

	return nil
}

func testAccAWSSSMMaintenanceWindowTargetBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "test" {
  name          = %[1]q
  description   = "This resource is for test purpose only"
  window_id     = aws_ssm_maintenance_window.test.id
  resource_type = "INSTANCE"

  targets {
    key    = "tag:Name"
    values = ["acceptance_test"]
  }

  targets {
    key    = "tag:Name2"
    values = ["acceptance_test", "acceptance_test2"]
  }
}
`, rName)
}

func testAccAWSSSMMaintenanceWindowTargetBasicResourceGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "test" {
  name          = %[1]q
  description   = "This resource is for test purpose only"
  window_id     = aws_ssm_maintenance_window.test.id
  resource_type = "RESOURCE_GROUP"

  targets {
    key    = "resource-groups:ResourceTypeFilters"
    values = ["AWS::EC2::Instance"]
  }

  targets {
    key    = "resource-groups:Name"
    values = ["resource-group-name"]
  }
}
`, rName)
}

func testAccAWSSSMMaintenanceWindowTargetNoNameOrDescriptionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "test" {
  window_id     = aws_ssm_maintenance_window.test.id
  resource_type = "INSTANCE"

  targets {
    key    = "tag:Name"
    values = ["acceptance_test"]
  }

  targets {
    key    = "tag:Name2"
    values = ["acceptance_test", "acceptance_test2"]
  }
}
`, rName)
}

func testAccAWSSSMMaintenanceWindowTargetBasicConfigUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "test" {
  name              = %[1]q
  description       = "This resource is for test purpose only - updated"
  window_id         = aws_ssm_maintenance_window.test.id
  resource_type     = "INSTANCE"
  owner_information = "something"

  targets {
    key    = "tag:Name"
    values = ["acceptance_test"]
  }

  targets {
    key    = "tag:Updated"
    values = ["new-value"]
  }
}
`, rName)
}

func testAccAWSSSMMaintenanceWindowTargetBasicConfigWithTarget(rName, tName, tDesc string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  name     = %[1]q
  schedule = "cron(0 16 ? * TUE *)"
  duration = 3
  cutoff   = 1
}

resource "aws_ssm_maintenance_window_target" "test" {
  name              = %[2]q
  description       = %[3]q
  window_id         = aws_ssm_maintenance_window.test.id
  resource_type     = "INSTANCE"
  owner_information = "something"

  targets {
    key    = "tag:Name"
    values = ["acceptance_test"]
  }

  targets {
    key    = "tag:Updated"
    values = ["new-value"]
  }
}
`, rName, tName, tDesc)
}

func testAccAWSSSMMaintenanceWindowTargetImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["window_id"], rs.Primary.ID), nil
	}
}
