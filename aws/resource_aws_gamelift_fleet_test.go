package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const testAccGameliftFleetPrefix = "tf_acc_fleet_"

func init() {
	resource.AddTestSweepers("aws_gamelift_fleet", &resource.Sweeper{
		Name: "aws_gamelift_fleet",
		Dependencies: []string{
			"aws_gamelift_build",
		},
		F: testSweepGameliftFleets,
	})
}

func testSweepGameliftFleets(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).gameliftconn

	return testAccGameliftListFleets(conn, nil, region, func(fleetIds []*string) error {
		if len(fleetIds) == 0 {
			log.Print("[DEBUG] No Gamelift Fleets to sweep")
			return nil
		}

		out, err := conn.DescribeFleetAttributes(&gamelift.DescribeFleetAttributesInput{
			FleetIds: fleetIds,
		})
		if err != nil {
			return fmt.Errorf("Error describing Gamelift Fleet attributes: %s", err)
		}

		log.Printf("[INFO] Found %d Gamelift Fleets", len(out.FleetAttributes))

		for _, attr := range out.FleetAttributes {
			if !strings.HasPrefix(*attr.Name, testAccGameliftFleetPrefix) {
				continue
			}

			log.Printf("[INFO] Deleting Gamelift Fleet %q", *attr.FleetId)
			err := resource.Retry(60*time.Minute, func() *resource.RetryError {
				_, err := conn.DeleteFleet(&gamelift.DeleteFleetInput{
					FleetId: attr.FleetId,
				})
				if err != nil {
					msg := fmt.Sprintf("Cannot delete fleet %s that is in status of ", *attr.FleetId)
					if isAWSErr(err, gamelift.ErrCodeInvalidRequestException, msg) {
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("Error deleting Gamelift Fleet (%s): %s",
					*attr.FleetId, err)
			}

			err = waitForGameliftFleetToBeDeleted(conn, *attr.FleetId, 5*time.Minute)
			if err != nil {
				return fmt.Errorf("Error waiting for Gamelift Fleet (%s) to be deleted: %s",
					*attr.FleetId, err)
			}
		}
		return nil
	})
}

func testAccGameliftListFleets(conn *gamelift.GameLift, nextToken *string, region string, f func([]*string) error) error {
	resp, err := conn.ListFleets(&gamelift.ListFleetsInput{
		NextToken: nextToken,
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Gamelift Fleet sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing Gamelift Fleets: %s", err)
	}

	err = f(resp.FleetIds)
	if err != nil {
		return err
	}
	if nextToken != nil {
		return testAccGameliftListFleets(conn, nextToken, region, f)
	}
	return nil
}

func TestDiffGameliftPortSettings(t *testing.T) {
	testCases := []struct {
		Old           []interface{}
		New           []interface{}
		ExpectedAuths []*gamelift.IpPermission
		ExpectedRevs  []*gamelift.IpPermission
	}{
		{ // No change
			Old: []interface{}{
				map[string]interface{}{
					"from_port": 8443,
					"ip_range":  "192.168.0.0/24",
					"protocol":  "TCP",
					"to_port":   8443,
				},
			},
			New: []interface{}{
				map[string]interface{}{
					"from_port": 8443,
					"ip_range":  "192.168.0.0/24",
					"protocol":  "TCP",
					"to_port":   8443,
				},
			},
			ExpectedAuths: []*gamelift.IpPermission{},
			ExpectedRevs:  []*gamelift.IpPermission{},
		},
		{ // Addition
			Old: []interface{}{
				map[string]interface{}{
					"from_port": 8443,
					"ip_range":  "192.168.0.0/24",
					"protocol":  "TCP",
					"to_port":   8443,
				},
			},
			New: []interface{}{
				map[string]interface{}{
					"from_port": 8443,
					"ip_range":  "192.168.0.0/24",
					"protocol":  "TCP",
					"to_port":   8443,
				},
				map[string]interface{}{
					"from_port": 8888,
					"ip_range":  "192.168.0.0/24",
					"protocol":  "TCP",
					"to_port":   8888,
				},
			},
			ExpectedAuths: []*gamelift.IpPermission{
				{
					FromPort: aws.Int64(8888),
					IpRange:  aws.String("192.168.0.0/24"),
					Protocol: aws.String("TCP"),
					ToPort:   aws.Int64(8888),
				},
			},
			ExpectedRevs: []*gamelift.IpPermission{},
		},
		{ // Removal
			Old: []interface{}{
				map[string]interface{}{
					"from_port": 8443,
					"ip_range":  "192.168.0.0/24",
					"protocol":  "TCP",
					"to_port":   8443,
				},
			},
			New:           []interface{}{},
			ExpectedAuths: []*gamelift.IpPermission{},
			ExpectedRevs: []*gamelift.IpPermission{
				{
					FromPort: aws.Int64(8443),
					IpRange:  aws.String("192.168.0.0/24"),
					Protocol: aws.String("TCP"),
					ToPort:   aws.Int64(8443),
				},
			},
		},
		{ // Removal + Addition
			Old: []interface{}{
				map[string]interface{}{
					"from_port": 8443,
					"ip_range":  "192.168.0.0/24",
					"protocol":  "TCP",
					"to_port":   8443,
				},
			},
			New: []interface{}{
				map[string]interface{}{
					"from_port": 8443,
					"ip_range":  "192.168.0.0/24",
					"protocol":  "UDP",
					"to_port":   8443,
				},
			},
			ExpectedAuths: []*gamelift.IpPermission{
				{
					FromPort: aws.Int64(8443),
					IpRange:  aws.String("192.168.0.0/24"),
					Protocol: aws.String("UDP"),
					ToPort:   aws.Int64(8443),
				},
			},
			ExpectedRevs: []*gamelift.IpPermission{
				{
					FromPort: aws.Int64(8443),
					IpRange:  aws.String("192.168.0.0/24"),
					Protocol: aws.String("TCP"),
					ToPort:   aws.Int64(8443),
				},
			},
		},
	}

	for _, tc := range testCases {
		a, r := diffGameliftPortSettings(tc.Old, tc.New)

		authsString := fmt.Sprintf("%+v", a)
		expectedAuths := fmt.Sprintf("%+v", tc.ExpectedAuths)
		if authsString != expectedAuths {
			t.Fatalf("Expected authorizations: %+v\nGiven: %+v", tc.ExpectedAuths, a)
		}

		revString := fmt.Sprintf("%+v", r)
		expectedRevs := fmt.Sprintf("%+v", tc.ExpectedRevs)
		if revString != expectedRevs {
			t.Fatalf("Expected authorizations: %+v\nGiven: %+v", tc.ExpectedRevs, r)
		}
	}
}

func TestAccAWSGameliftFleet_basic(t *testing.T) {
	var conf gamelift.FleetAttributes

	rString := acctest.RandString(8)

	fleetName := fmt.Sprintf("%s%s", testAccGameliftFleetPrefix, rString)
	uFleetName := fmt.Sprintf("%supd_%s", testAccGameliftFleetPrefix, rString)
	buildName := fmt.Sprintf("%s%s", testAccGameliftBuildPrefix, rString)

	desc := fmt.Sprintf("Updated description %s", rString)

	region := testAccGetRegion()
	g, err := testAccAWSGameliftSampleGame(region)
	if err != nil {
		t.Fatal(err)
	}

	loc := g.Location
	bucketName := *loc.Bucket
	roleArn := *loc.RoleArn
	key := *loc.Key

	launchPath := g.LaunchPath
	params := g.Parameters(33435)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftFleetBasicConfig(fleetName, launchPath, params, buildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftFleetExists("aws_gamelift_fleet.test", &conf),
					resource.TestCheckResourceAttrSet("aws_gamelift_fleet.test", "build_id"),
					resource.TestMatchResourceAttr("aws_gamelift_fleet.test", "arn", regexp.MustCompile(`^arn:[^:]+:gamelift:[^:]+:[^:]+:.+$`)),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_instance_type", "c4.large"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "log_paths.#", "0"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "name", fleetName),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "metric_groups.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "metric_groups.0", "default"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "new_game_session_protection_policy", "NoProtection"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "resource_creation_limit_policy.#", "0"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.0.server_process.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.0.server_process.0.concurrent_executions", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.0.server_process.0.launch_path", launchPath),
				),
			},
			{
				Config: testAccAWSGameliftFleetBasicUpdatedConfig(desc, uFleetName, launchPath, params, buildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftFleetExists("aws_gamelift_fleet.test", &conf),
					resource.TestCheckResourceAttrSet("aws_gamelift_fleet.test", "build_id"),
					resource.TestMatchResourceAttr("aws_gamelift_fleet.test", "arn", regexp.MustCompile(`^arn:[^:]+:gamelift:[^:]+:[^:]+:.+$`)),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_instance_type", "c4.large"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "log_paths.#", "0"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "name", uFleetName),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "description", desc),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "metric_groups.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "metric_groups.0", "UpdatedGroup"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "new_game_session_protection_policy", "FullProtection"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "resource_creation_limit_policy.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "resource_creation_limit_policy.0.new_game_sessions_per_creator", "2"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "resource_creation_limit_policy.0.policy_period_in_minutes", "15"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.0.server_process.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.0.server_process.0.concurrent_executions", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.0.server_process.0.launch_path", launchPath),
				),
			},
		},
	})
}

func TestAccAWSGameliftFleet_allFields(t *testing.T) {
	var conf gamelift.FleetAttributes

	rString := acctest.RandString(8)

	fleetName := fmt.Sprintf("%s%s", testAccGameliftFleetPrefix, rString)
	buildName := fmt.Sprintf("%s%s", testAccGameliftBuildPrefix, rString)

	desc := fmt.Sprintf("Terraform Acceptance Test %s", rString)

	region := testAccGetRegion()
	g, err := testAccAWSGameliftSampleGame(region)
	if err != nil {
		t.Fatal(err)
	}

	loc := g.Location
	bucketName := *loc.Bucket
	roleArn := *loc.RoleArn
	key := *loc.Key

	launchPath := g.LaunchPath
	params := []string{
		g.Parameters(33435),
		g.Parameters(33436),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftFleetAllFieldsConfig(fleetName, desc, launchPath, params[0], buildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftFleetExists("aws_gamelift_fleet.test", &conf),
					resource.TestCheckResourceAttrSet("aws_gamelift_fleet.test", "build_id"),
					resource.TestMatchResourceAttr("aws_gamelift_fleet.test", "arn", regexp.MustCompile(`^arn:[^:]+:gamelift:[^:]+:[^:]+:.+$`)),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_instance_type", "c4.large"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "name", fleetName),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "description", desc),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.#", "3"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.0.from_port", "8080"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.0.ip_range", "8.8.8.8/32"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.0.to_port", "8080"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.1.from_port", "8443"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.1.ip_range", "8.8.0.0/16"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.1.protocol", "TCP"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.1.to_port", "8443"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.2.from_port", "60000"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.2.ip_range", "8.8.8.8/32"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.2.protocol", "UDP"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.2.to_port", "60000"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "log_paths.#", "0"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "metric_groups.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "metric_groups.0", "TerraformAccTest"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "new_game_session_protection_policy", "FullProtection"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "operating_system", "WINDOWS_2012"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "resource_creation_limit_policy.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "resource_creation_limit_policy.0.new_game_sessions_per_creator", "4"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "resource_creation_limit_policy.0.policy_period_in_minutes", "25"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.0.game_session_activation_timeout_seconds", "35"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.0.max_concurrent_game_session_activations", "99"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.0.server_process.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.0.server_process.0.concurrent_executions", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.0.server_process.0.launch_path", launchPath),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.0.server_process.0.parameters", params[0]),
				),
			},
			{
				Config: testAccAWSGameliftFleetAllFieldsUpdatedConfig(fleetName, desc, launchPath, params[1], buildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftFleetExists("aws_gamelift_fleet.test", &conf),
					resource.TestCheckResourceAttrSet("aws_gamelift_fleet.test", "build_id"),
					resource.TestMatchResourceAttr("aws_gamelift_fleet.test", "arn", regexp.MustCompile(`^arn:[^:]+:gamelift:[^:]+:[^:]+:.+$`)),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_instance_type", "c4.large"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "name", fleetName),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "description", desc),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.#", "3"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.0.from_port", "8888"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.0.ip_range", "8.8.8.8/32"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.0.protocol", "TCP"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.0.to_port", "8888"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.1.from_port", "8443"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.1.ip_range", "8.4.0.0/16"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.1.protocol", "TCP"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.1.to_port", "8443"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.2.from_port", "60000"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.2.ip_range", "8.8.8.8/32"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.2.protocol", "UDP"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "ec2_inbound_permission.2.to_port", "60000"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "log_paths.#", "0"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "metric_groups.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "metric_groups.0", "TerraformAccTest"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "new_game_session_protection_policy", "FullProtection"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "operating_system", "WINDOWS_2012"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "resource_creation_limit_policy.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "resource_creation_limit_policy.0.new_game_sessions_per_creator", "4"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "resource_creation_limit_policy.0.policy_period_in_minutes", "25"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.0.game_session_activation_timeout_seconds", "35"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.0.max_concurrent_game_session_activations", "98"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.0.server_process.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.0.server_process.0.concurrent_executions", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.0.server_process.0.launch_path", launchPath),
					resource.TestCheckResourceAttr("aws_gamelift_fleet.test", "runtime_configuration.0.server_process.0.parameters", params[1]),
				),
			},
		},
	})
}

func testAccCheckAWSGameliftFleetExists(n string, res *gamelift.FleetAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Gamelift Fleet ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).gameliftconn

		out, err := conn.DescribeFleetAttributes(&gamelift.DescribeFleetAttributesInput{
			FleetIds: aws.StringSlice([]string{rs.Primary.ID}),
		})
		if err != nil {
			return err
		}
		attributes := out.FleetAttributes
		if len(attributes) < 1 {
			return fmt.Errorf("Gamelift Fleet %q not found", rs.Primary.ID)
		}
		if len(attributes) != 1 {
			return fmt.Errorf("Expected exactly 1 Gamelift Fleet, found %d under %q",
				len(attributes), rs.Primary.ID)
		}
		fleet := attributes[0]

		if *fleet.FleetId != rs.Primary.ID {
			return fmt.Errorf("Gamelift Fleet not found")
		}

		*res = *fleet

		return nil
	}
}

func testAccCheckAWSGameliftFleetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).gameliftconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_gamelift_fleet" {
			continue
		}

		out, err := conn.DescribeFleetAttributes(&gamelift.DescribeFleetAttributesInput{
			FleetIds: aws.StringSlice([]string{rs.Primary.ID}),
		})
		if err != nil {
			return err
		}

		attributes := out.FleetAttributes

		if len(attributes) > 0 {
			return fmt.Errorf("Gamelift Fleet still exists")
		}

		return nil
	}

	return nil
}

func testAccAWSGameliftFleetBasicConfig(fleetName, launchPath, params, buildName, bucketName, key, roleArn string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_fleet" "test" {
  build_id = "${aws_gamelift_build.test.id}"
  ec2_instance_type = "c4.large"
  name = "%s"
  runtime_configuration {
    server_process {
      concurrent_executions = 1
      launch_path = %q
      parameters = "%s"
    }
  }
}
%s
`, fleetName, launchPath, params, testAccAWSGameliftFleetBasicTemplate(buildName, bucketName, key, roleArn))
}

func testAccAWSGameliftFleetBasicUpdatedConfig(desc, fleetName, launchPath, params, buildName, bucketName, key, roleArn string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_fleet" "test" {
  build_id = "${aws_gamelift_build.test.id}"
  ec2_instance_type = "c4.large"
  description = "%s"
  name = "%s"
  metric_groups = ["UpdatedGroup"]
  new_game_session_protection_policy = "FullProtection"
  resource_creation_limit_policy {
    new_game_sessions_per_creator = 2
    policy_period_in_minutes = 15
  }
  runtime_configuration {
    server_process {
      concurrent_executions = 1
      launch_path = %q
      parameters = "%s"
    }
  }
}
%s
`, desc, fleetName, launchPath, params, testAccAWSGameliftFleetBasicTemplate(buildName, bucketName, key, roleArn))
}

func testAccAWSGameliftFleetAllFieldsConfig(fleetName, desc, launchPath string, params string, buildName, bucketName, key, roleArn string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_fleet" "test" {
  build_id = "${aws_gamelift_build.test.id}"
  ec2_instance_type = "c4.large"
  name = "%s"
  description = "%s"

  ec2_inbound_permission {
    from_port = 8080
    ip_range = "8.8.8.8/32"
    protocol = "TCP"
    to_port = 8080
  }
  ec2_inbound_permission {
    from_port = 8443
    ip_range = "8.8.0.0/16"
    protocol = "TCP"
    to_port = 8443
  }
  ec2_inbound_permission {
    from_port = 60000
    ip_range = "8.8.8.8/32"
    protocol = "UDP"
    to_port = 60000
  }

  metric_groups = ["TerraformAccTest"]
  new_game_session_protection_policy = "FullProtection"
  
  resource_creation_limit_policy {
    new_game_sessions_per_creator = 4
    policy_period_in_minutes = 25
  }

  runtime_configuration {
    game_session_activation_timeout_seconds = 35
    max_concurrent_game_session_activations = 99

    server_process {
      concurrent_executions = 1
      launch_path = %q
      parameters = "%s"
    }
  }
}
%s
`, fleetName, desc, launchPath, params,
		testAccAWSGameliftFleetBasicTemplate(buildName, bucketName, key, roleArn))
}

func testAccAWSGameliftFleetAllFieldsUpdatedConfig(fleetName, desc, launchPath string, params string, buildName, bucketName, key, roleArn string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_fleet" "test" {
  build_id = "${aws_gamelift_build.test.id}"
  ec2_instance_type = "c4.large"
  name = "%s"
  description = "%s"

  ec2_inbound_permission {
    from_port = 8888
    ip_range = "8.8.8.8/32"
    protocol = "TCP"
    to_port = 8888
  }
  ec2_inbound_permission {
    from_port = 8443
    ip_range = "8.4.0.0/16"
    protocol = "TCP"
    to_port = 8443
  }
  ec2_inbound_permission {
    from_port = 60000
    ip_range = "8.8.8.8/32"
    protocol = "UDP"
    to_port = 60000
  }

  metric_groups = ["TerraformAccTest"]
  new_game_session_protection_policy = "FullProtection"
  
  resource_creation_limit_policy {
    new_game_sessions_per_creator = 4
    policy_period_in_minutes = 25
  }

  runtime_configuration {
    game_session_activation_timeout_seconds = 35
    max_concurrent_game_session_activations = 98

    server_process {
      concurrent_executions = 1
      launch_path = %q
      parameters = "%s"
    }
  }
}
%s
`, fleetName, desc, launchPath, params,
		testAccAWSGameliftFleetBasicTemplate(buildName, bucketName, key, roleArn))
}

func testAccAWSGameliftFleetBasicTemplate(buildName, bucketName, key, roleArn string) string {
	return fmt.Sprintf(`resource "aws_gamelift_build" "test" {
  name = "%s"
  operating_system = "WINDOWS_2012"
  storage_location {
    bucket = "%s"
    key = "%s"
    role_arn = "%s"
  }
}
`, buildName, bucketName, key, roleArn)
}
