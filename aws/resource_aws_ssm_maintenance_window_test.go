package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSSMMaintenanceWindow_basic(t *testing.T) {
	var winId ssm.MaintenanceWindowIdentity
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &winId),
					resource.TestCheckResourceAttr(resourceName, "cutoff", "1"),
					resource.TestCheckResourceAttr(resourceName, "duration", "3"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "end_date", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule_timezone", ""),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 16 ? * TUE *)"),
					resource.TestCheckResourceAttr(resourceName, "start_date", ""),
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

func TestAccAWSSSMMaintenanceWindow_disappears(t *testing.T) {
	var winId ssm.MaintenanceWindowIdentity
	name := acctest.RandString(10)
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &winId),
					testAccCheckAWSSSMMaintenanceWindowDisappears(&winId),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindow_multipleUpdates(t *testing.T) {
	var maintenanceWindow1, maintenanceWindow2 ssm.MaintenanceWindowIdentity
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowConfig(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "cutoff", "1"),
					resource.TestCheckResourceAttr(resourceName, "duration", "3"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 16 ? * TUE *)"),
				),
			},
			{
				Config: testAccAWSSSMMaintenanceWindowConfigMultipleUpdates(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "cutoff", "8"),
					resource.TestCheckResourceAttr(resourceName, "duration", "10"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 16 ? * WED *)"),
				),
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindow_Cutoff(t *testing.T) {
	var maintenanceWindow1, maintenanceWindow2 ssm.MaintenanceWindowIdentity
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowConfigCutoff(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "cutoff", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMMaintenanceWindowConfigCutoff(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "cutoff", "2"),
				),
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindow_Duration(t *testing.T) {
	var maintenanceWindow1, maintenanceWindow2 ssm.MaintenanceWindowIdentity
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowConfigDuration(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "duration", "3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMMaintenanceWindowConfigDuration(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "duration", "10"),
				),
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindow_Enabled(t *testing.T) {
	var maintenanceWindow1, maintenanceWindow2 ssm.MaintenanceWindowIdentity
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowConfigEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMMaintenanceWindowConfigEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindow_EndDate(t *testing.T) {
	var maintenanceWindow1, maintenanceWindow2, maintenanceWindow3 ssm.MaintenanceWindowIdentity
	endDate1 := time.Now().UTC().Add(365 * 24 * time.Hour).Format(time.RFC3339)
	endDate2 := time.Now().UTC().Add(730 * 24 * time.Hour).Format(time.RFC3339)
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowConfigEndDate(rName, endDate1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "end_date", endDate1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMMaintenanceWindowConfigEndDate(rName, endDate2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "end_date", endDate2),
				),
			},
			{
				Config: testAccAWSSSMMaintenanceWindowConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &maintenanceWindow3),
					resource.TestCheckResourceAttr(resourceName, "end_date", ""),
				),
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindow_Schedule(t *testing.T) {
	var maintenanceWindow1, maintenanceWindow2 ssm.MaintenanceWindowIdentity
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowConfigSchedule(rName, "cron(0 16 ? * TUE *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 16 ? * TUE *)"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMMaintenanceWindowConfigSchedule(rName, "cron(0 16 ? * WED *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(0 16 ? * WED *)"),
				),
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindow_ScheduleTimezone(t *testing.T) {
	var maintenanceWindow1, maintenanceWindow2, maintenanceWindow3 ssm.MaintenanceWindowIdentity
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowConfigScheduleTimezone(rName, "America/Los_Angeles"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "schedule_timezone", "America/Los_Angeles"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMMaintenanceWindowConfigScheduleTimezone(rName, "America/New_York"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "schedule_timezone", "America/New_York"),
				),
			},
			{
				Config: testAccAWSSSMMaintenanceWindowConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &maintenanceWindow3),
					resource.TestCheckResourceAttr(resourceName, "schedule_timezone", ""),
				),
			},
		},
	})
}

func TestAccAWSSSMMaintenanceWindow_StartDate(t *testing.T) {
	var maintenanceWindow1, maintenanceWindow2, maintenanceWindow3 ssm.MaintenanceWindowIdentity
	startDate1 := time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339)
	startDate2 := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_maintenance_window.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMMaintenanceWindowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMMaintenanceWindowConfigStartDate(rName, startDate1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &maintenanceWindow1),
					resource.TestCheckResourceAttr(resourceName, "start_date", startDate1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMMaintenanceWindowConfigStartDate(rName, startDate2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &maintenanceWindow2),
					resource.TestCheckResourceAttr(resourceName, "start_date", startDate2),
				),
			},
			{
				Config: testAccAWSSSMMaintenanceWindowConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMMaintenanceWindowExists(resourceName, &maintenanceWindow3),
					resource.TestCheckResourceAttr(resourceName, "start_date", ""),
				),
			},
		},
	})
}

func testAccCheckAWSSSMMaintenanceWindowExists(n string, res *ssm.MaintenanceWindowIdentity) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Maintenance Window ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ssmconn

		resp, err := conn.DescribeMaintenanceWindows(&ssm.DescribeMaintenanceWindowsInput{
			Filters: []*ssm.MaintenanceWindowFilter{
				{
					Key:    aws.String("Name"),
					Values: []*string{aws.String(rs.Primary.Attributes["name"])},
				},
			},
		})
		if err != nil {
			return err
		}

		for _, i := range resp.WindowIdentities {
			if *i.WindowId == rs.Primary.ID {
				*res = *i
				return nil
			}
		}

		return fmt.Errorf("No AWS SSM Maintenance window found")
	}
}

func testAccCheckAWSSSMMaintenanceWindowDisappears(maintenanceWindowIdentity *ssm.MaintenanceWindowIdentity) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ssmconn

		id := aws.StringValue(maintenanceWindowIdentity.WindowId)
		_, err := conn.DeleteMaintenanceWindow(&ssm.DeleteMaintenanceWindowInput{
			WindowId: aws.String(id),
		})
		if err != nil {
			return fmt.Errorf("error deleting maintenance window %s: %s", id, err)
		}
		return nil
	}
}

func testAccCheckAWSSSMMaintenanceWindowDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ssmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_maintenance_window" {
			continue
		}

		out, err := conn.DescribeMaintenanceWindows(&ssm.DescribeMaintenanceWindowsInput{
			Filters: []*ssm.MaintenanceWindowFilter{
				{
					Key:    aws.String("Name"),
					Values: []*string{aws.String(rs.Primary.Attributes["name"])},
				},
			},
		})

		if err != nil {
			return err
		}

		if len(out.WindowIdentities) > 0 {
			return fmt.Errorf("Expected AWS SSM Maintenance Document to be gone, but was still found")
		}

		return nil
	}

	return nil
}

func testAccAWSSSMMaintenanceWindowConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = 3
  name     = %q
  schedule = "cron(0 16 ? * TUE *)"
}
`, rName)
}

func testAccAWSSSMMaintenanceWindowConfigCutoff(rName string, cutoff int) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = %d
  duration = 3
  name     = %q
  schedule = "cron(0 16 ? * TUE *)"
}
`, cutoff, rName)
}

func testAccAWSSSMMaintenanceWindowConfigDuration(rName string, duration int) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = %d
  name     = %q
  schedule = "cron(0 16 ? * TUE *)"
}
`, duration, rName)
}

func testAccAWSSSMMaintenanceWindowConfigEnabled(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = 3
  enabled  = %t
  name     = %q
  schedule = "cron(0 16 ? * TUE *)"
}
`, enabled, rName)
}

func testAccAWSSSMMaintenanceWindowConfigEndDate(rName, endDate string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = 3
  end_date = %q
  name     = %q
  schedule = "cron(0 16 ? * TUE *)"
}
`, endDate, rName)
}

func testAccAWSSSMMaintenanceWindowConfigMultipleUpdates(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 8
  duration = 10
  enabled  = false
  name     = %q
  schedule = "cron(0 16 ? * WED *)"
}
`, rName)
}

func testAccAWSSSMMaintenanceWindowConfigSchedule(rName, schedule string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff   = 1
  duration = 3
  name     = %q
  schedule = %q
}
`, rName, schedule)
}

func testAccAWSSSMMaintenanceWindowConfigScheduleTimezone(rName, scheduleTimezone string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff            = 1
  duration          = 3
  name              = %q
  schedule          = "cron(0 16 ? * TUE *)"
  schedule_timezone = %q
}
`, rName, scheduleTimezone)
}

func testAccAWSSSMMaintenanceWindowConfigStartDate(rName, startDate string) string {
	return fmt.Sprintf(`
resource "aws_ssm_maintenance_window" "test" {
  cutoff     = 1
  duration   = 3
  name       = %q
  schedule   = "cron(0 16 ? * TUE *)"
  start_date = %q
}
`, rName, startDate)
}
