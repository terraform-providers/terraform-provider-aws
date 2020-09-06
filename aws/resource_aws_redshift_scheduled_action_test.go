package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_redshift_scheduled_action", &resource.Sweeper{
		Name: "aws_redshift_scheduled_action",
		Dependencies: []string{
			"aws_iam_role",
			"aws_iam_policy",
			"aws_iam_role_policy_attachment",
		},
		F: testSweepRedshiftScheduledActions,
	})
}

func testSweepRedshiftScheduledActions(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).redshiftconn

	req := &redshift.DescribeScheduledActionsInput{}

	resp, err := conn.DescribeScheduledActions(req)
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Redshift Regional Scheduled Actions sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error describing Redshift Regional Scheduled Actions: %s", err)
	}

	if len(resp.ScheduledActions) == 0 {
		log.Print("[DEBUG] No AWS Redshift Regional Scheduled Actions to sweep")
		return nil
	}

	for _, ScheduledActions := range resp.ScheduledActions {
		identifier := aws.StringValue(ScheduledActions.ScheduledActionName)

		hasPrefix := false
		prefixes := []string{"tf-test-"}

		for _, prefix := range prefixes {
			if strings.HasPrefix(identifier, prefix) {
				hasPrefix = true
				break
			}
		}

		if !hasPrefix {
			log.Printf("[INFO] Skipping Delete Redshift Scheduled Action: %s", identifier)
			continue
		}

		_, err := conn.DeleteScheduledAction(&redshift.DeleteScheduledActionInput{
			ScheduledActionName: ScheduledActions.ScheduledActionName,
		})
		if isAWSErr(err, redshift.ErrCodeScheduledActionNotFoundFault, "") {
			return nil
		}
		if err != nil {
			return fmt.Errorf("Error deleting Redshift Scheduled Action %s: %s", identifier, err)
		}
	}

	return nil
}

func TestAccAWSRedshiftScheduledActionPauseCluster_basic(t *testing.T) {
	var v redshift.ScheduledAction

	rName := acctest.RandString(8)
	resourceName := "aws_redshift_scheduled_action.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftScheduledActionConfigPauseCluster(rName, "cron(00 23 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftScheduledActionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-test-%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "name", resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(00 23 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.action", "PauseCluster"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.cluster_identifier", "tf-test-identifier"),
				),
			},
			{
				Config: testAccAWSRedshiftScheduledActionConfigPauseCluster(rName, "at(2060-03-04T17:27:00)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftScheduledActionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-test-%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "name", resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "schedule", "at(2060-03-04T17:27:00)"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.action", "PauseCluster"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.cluster_identifier", "tf-test-identifier"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"active",
				},
			},
		},
	})
}

func TestAccAWSRedshiftScheduledActionPauseCluster_WithOptions(t *testing.T) {
	var v redshift.ScheduledAction

	rName := acctest.RandString(8)
	resourceName := "aws_redshift_scheduled_action.default"
	startTime := time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().UTC().Add(2 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftScheduledActionConfigPauseClusterWithFullOption(rName, "cron(00 * * * ? *)", "This is test action", "true", startTime, endTime),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftScheduledActionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-test-%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "description", "This is test action"),
					resource.TestCheckResourceAttr(resourceName, "active", "true"),
					resource.TestCheckResourceAttr(resourceName, "start_time", startTime),
					resource.TestCheckResourceAttr(resourceName, "end_time", endTime),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"active",
				},
			},
		},
	})
}

func TestAccAWSRedshiftScheduledActionResumeCluster_basic(t *testing.T) {
	var v redshift.ScheduledAction

	rName := acctest.RandString(8)
	resourceName := "aws_redshift_scheduled_action.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftScheduledActionConfigResumeCluster(rName, "cron(00 23 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftScheduledActionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-test-%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "name", resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(00 23 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.action", "ResumeCluster"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.cluster_identifier", "tf-test-identifier"),
				),
			},
			{
				Config: testAccAWSRedshiftScheduledActionConfigResumeCluster(rName, "at(2060-03-04T17:27:00)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftScheduledActionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-test-%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "name", resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "schedule", "at(2060-03-04T17:27:00)"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.action", "ResumeCluster"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.cluster_identifier", "tf-test-identifier"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"active",
				},
			},
		},
	})
}

func TestAccAWSRedshiftScheduledActionResizeCluster_basic(t *testing.T) {
	var v redshift.ScheduledAction

	rName := acctest.RandString(8)
	resourceName := "aws_redshift_scheduled_action.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftScheduledActionConfigResizeClusterBasic(rName, "cron(00 23 * * ? *)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftScheduledActionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-test-%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "name", resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "schedule", "cron(00 23 * * ? *)"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.action", "ResizeCluster"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.cluster_identifier", "tf-test-identifier"),
				),
			},
			{
				Config: testAccAWSRedshiftScheduledActionConfigResizeClusterBasic(rName, "at(2060-03-04T17:27:00)"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftScheduledActionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-test-%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "name", resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "schedule", "at(2060-03-04T17:27:00)"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.action", "ResizeCluster"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.cluster_identifier", "tf-test-identifier"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"active",
				},
			},
		},
	})
}

func TestAccAWSRedshiftScheduledActionResizeCluster_WithOptions(t *testing.T) {
	var v redshift.ScheduledAction

	rName := acctest.RandString(8)
	resourceName := "aws_redshift_scheduled_action.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftScheduledActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftScheduledActionConfigResizeClusterWithFullOption(rName, "cron(00 23 * * ? *)", "true", "multi-node", "dc1.large", "2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftScheduledActionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-test-%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "name", resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.classic", "true"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.cluster_type", "multi-node"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.node_type", "dc1.large"),
					resource.TestCheckResourceAttr(resourceName, "target_action.0.number_of_nodes", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"active",
				},
			},
		},
	})
}

func testAccCheckAWSRedshiftScheduledActionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_scheduled_action" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).redshiftconn
		resp, err := conn.DescribeScheduledActions(&redshift.DescribeScheduledActionsInput{
			ScheduledActionName: aws.String(rs.Primary.ID),
		})

		if isAWSErr(err, "ScheduledActionNotFound", "was not found.") {
			continue
		}

		if err == nil {
			if len(resp.ScheduledActions) != 0 {
				for _, s := range resp.ScheduledActions {
					if *s.ScheduledActionName == rs.Primary.ID {
						return fmt.Errorf("Redshift Cluster Scheduled Action %s still exists", rs.Primary.ID)
					}
				}
			}
		}

		return err
	}

	return nil
}

func testAccCheckAWSRedshiftScheduledActionExists(n string, v *redshift.ScheduledAction) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Cluster Scheduled Action ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).redshiftconn
		resp, err := conn.DescribeScheduledActions(&redshift.DescribeScheduledActionsInput{
			ScheduledActionName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		for _, s := range resp.ScheduledActions {
			if *s.ScheduledActionName == rs.Primary.ID {
				*v = *s
				return nil
			}
		}

		return fmt.Errorf("Redshift Scheduled Action (%s) not found", rs.Primary.ID)
	}
}

func testAccAWSRedshiftScheduledActionConfigDependentResource(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "default" {
  name = "tf-test-%s"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": [
          "scheduler.redshift.amazonaws.com",
          "redshift.amazonaws.com"
        ]
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "default" {
  name = "tf-test-%s"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
      {
          "Sid": "VisualEditor0",
          "Effect": "Allow",
          "Action": [
              "redshift:PauseCluster",
              "redshift:ResumeCluster",
              "redshift:ResizeCluster"
          ],
          "Resource": "*"
      }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "default" {
  policy_arn = aws_iam_policy.default.arn
  role = aws_iam_role.default.name
}
	`, rName, rName)
}

func testAccAWSRedshiftScheduledActionConfigPauseCluster(rName, schedule string) string {
	return fmt.Sprintf(`

%s

resource "aws_redshift_scheduled_action" "default" {
	name = "tf-test-%s"
    schedule = "%s"
    iam_role = aws_iam_role.default.arn
    target_action {
      action = "PauseCluster"
      cluster_identifier = "tf-test-identifier"
    }
}
	`, testAccAWSRedshiftScheduledActionConfigDependentResource(rName), rName, schedule)
}

func testAccAWSRedshiftScheduledActionConfigPauseClusterWithFullOption(rName, schedule, description, active, startTime, endTime string) string {
	return fmt.Sprintf(`

%s

resource "aws_redshift_scheduled_action" "default" {
	name = "tf-test-%s"
    description = "%s"
    active = %s
    start_time = "%s"
    end_time = "%s"
    schedule = "%s"
    iam_role = aws_iam_role.default.arn
    target_action {
      action = "PauseCluster"
      cluster_identifier = "tf-test-identifier"
    }
}
	`, testAccAWSRedshiftScheduledActionConfigDependentResource(rName), rName, description, active, startTime, endTime, schedule)
}

func testAccAWSRedshiftScheduledActionConfigResumeCluster(rName, schedule string) string {
	return fmt.Sprintf(`

%s

resource "aws_redshift_scheduled_action" "default" {
	name = "tf-test-%s"
    schedule = "%s"
    iam_role = aws_iam_role.default.arn
    target_action {
      action = "ResumeCluster"
      cluster_identifier = "tf-test-identifier"
    }
}
	`, testAccAWSRedshiftScheduledActionConfigDependentResource(rName), rName, schedule)
}

func testAccAWSRedshiftScheduledActionConfigResizeClusterBasic(rName, schedule string) string {
	return fmt.Sprintf(`

%s

resource "aws_redshift_scheduled_action" "default" {
	name = "tf-test-%s"
    schedule = "%s"
    iam_role = aws_iam_role.default.arn
    target_action {
      action = "ResizeCluster"
      cluster_identifier = "tf-test-identifier"
    }
}
	`, testAccAWSRedshiftScheduledActionConfigDependentResource(rName), rName, schedule)
}

func testAccAWSRedshiftScheduledActionConfigResizeClusterWithFullOption(rName, schedule, classic, clusterType, nodeType, numberOfNodes string) string {
	return fmt.Sprintf(`

%s

resource "aws_redshift_scheduled_action" "default" {
	name = "tf-test-%s"
    schedule = "%s"
    iam_role = aws_iam_role.default.arn
    target_action {
      action = "ResizeCluster"
      cluster_identifier = "tf-test-identifier"
	  classic = %s 
	  cluster_type = "%s"
	  node_type = "%s"
	  number_of_nodes = %s
    }
}
	`, testAccAWSRedshiftScheduledActionConfigDependentResource(rName), rName, schedule, classic, clusterType, nodeType, numberOfNodes)
}
