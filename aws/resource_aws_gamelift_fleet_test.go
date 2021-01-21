package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

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

	fleetName := acctest.RandomWithPrefix("tf-acc-fleet")
	uFleetName := acctest.RandomWithPrefix("tf-acc-fleet-upd")
	buildName := acctest.RandomWithPrefix("tf-acc-build")

	desc := fmt.Sprintf("Updated description %s", acctest.RandString(8))

	region := testAccGetRegion()
	g, err := testAccAWSGameliftSampleGame(region)

	if isResourceNotFoundError(err) {
		t.Skip(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	loc := g.Location
	bucketName := *loc.Bucket
	roleArn := *loc.RoleArn
	key := *loc.Key

	launchPath := g.LaunchPath
	params := g.Parameters(33435)
	resourceName := "aws_gamelift_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(gamelift.EndpointsID, t)
			testAccPreCheckAWSGamelift(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftFleetBasicConfig(fleetName, launchPath, params, buildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftFleetExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "build_id"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`fleet/fleet-.+`)),
					resource.TestCheckResourceAttr(resourceName, "ec2_instance_type", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "log_paths.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", fleetName),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.0", "default"),
					resource.TestCheckResourceAttr(resourceName, "new_game_session_protection_policy", "NoProtection"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.concurrent_executions", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.launch_path", launchPath),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSGameliftFleetBasicUpdatedConfig(desc, uFleetName, launchPath, params, buildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftFleetExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "build_id"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`fleet/fleet-.+`)), resource.TestCheckResourceAttr(resourceName, "ec2_instance_type", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "log_paths.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "name", uFleetName),
					resource.TestCheckResourceAttr(resourceName, "description", desc),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.0", "UpdatedGroup"),
					resource.TestCheckResourceAttr(resourceName, "new_game_session_protection_policy", "FullProtection"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.0.new_game_sessions_per_creator", "2"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.0.policy_period_in_minutes", "15"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.concurrent_executions", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.launch_path", launchPath),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSGameliftFleet_tags(t *testing.T) {
	var conf gamelift.FleetAttributes

	fleetName := acctest.RandomWithPrefix("tf-acc-fleet")
	buildName := acctest.RandomWithPrefix("tf-acc-build")

	region := testAccGetRegion()
	g, err := testAccAWSGameliftSampleGame(region)

	if isResourceNotFoundError(err) {
		t.Skip(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	loc := g.Location
	bucketName := *loc.Bucket
	roleArn := *loc.RoleArn
	key := *loc.Key
	launchPath := g.LaunchPath
	params := g.Parameters(33435)

	resourceName := "aws_gamelift_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(gamelift.EndpointsID, t)
			testAccPreCheckAWSGamelift(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftFleetBasicConfigTags1(fleetName, launchPath, params, buildName, bucketName, key, roleArn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftFleetExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAWSGameliftFleetBasicConfigTags2(fleetName, launchPath, params, buildName, bucketName, key, roleArn, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftFleetExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSGameliftFleetBasicConfigTags1(fleetName, launchPath, params, buildName, bucketName, key, roleArn, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftFleetExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSGameliftFleet_allFields(t *testing.T) {
	var conf gamelift.FleetAttributes

	fleetName := acctest.RandomWithPrefix("tf-acc-fleet")
	buildName := acctest.RandomWithPrefix("tf-acc-build")

	desc := fmt.Sprintf("Terraform Acceptance Test %s", acctest.RandString(8))

	region := testAccGetRegion()
	g, err := testAccAWSGameliftSampleGame(region)

	if isResourceNotFoundError(err) {
		t.Skip(err)
	}

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
	resourceName := "aws_gamelift_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(gamelift.EndpointsID, t)
			testAccPreCheckAWSGamelift(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftFleetAllFieldsConfig(fleetName, desc, launchPath, params[0], buildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftFleetExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "build_id"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`fleet/fleet-.+`)), resource.TestCheckResourceAttr(resourceName, "ec2_instance_type", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "fleet_type", "ON_DEMAND"),
					resource.TestCheckResourceAttr(resourceName, "name", fleetName),
					resource.TestCheckResourceAttr(resourceName, "description", desc),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.0.from_port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.0.ip_range", "8.8.8.8/32"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.0.to_port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.1.from_port", "8443"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.1.ip_range", "8.8.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.1.protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.1.to_port", "8443"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.2.from_port", "60000"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.2.ip_range", "8.8.8.8/32"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.2.protocol", "UDP"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.2.to_port", "60000"),
					resource.TestCheckResourceAttr(resourceName, "log_paths.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.0", "TerraformAccTest"),
					resource.TestCheckResourceAttr(resourceName, "new_game_session_protection_policy", "FullProtection"),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "WINDOWS_2012"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.0.new_game_sessions_per_creator", "4"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.0.policy_period_in_minutes", "25"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.game_session_activation_timeout_seconds", "35"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.max_concurrent_game_session_activations", "99"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.concurrent_executions", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.launch_path", launchPath),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.parameters", params[0]),
				),
			},
			{
				Config: testAccAWSGameliftFleetAllFieldsUpdatedConfig(fleetName, desc, launchPath, params[1], buildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftFleetExists(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "build_id"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`fleet/fleet-.+`)), resource.TestCheckResourceAttr(resourceName, "ec2_instance_type", "c4.large"),
					resource.TestCheckResourceAttr(resourceName, "fleet_type", "ON_DEMAND"),
					resource.TestCheckResourceAttr(resourceName, "name", fleetName),
					resource.TestCheckResourceAttr(resourceName, "description", desc),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.0.from_port", "8888"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.0.ip_range", "8.8.8.8/32"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.0.to_port", "8888"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.1.from_port", "8443"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.1.ip_range", "8.4.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.1.protocol", "TCP"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.1.to_port", "8443"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.2.from_port", "60000"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.2.ip_range", "8.8.8.8/32"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.2.protocol", "UDP"),
					resource.TestCheckResourceAttr(resourceName, "ec2_inbound_permission.2.to_port", "60000"),
					resource.TestCheckResourceAttr(resourceName, "log_paths.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_groups.0", "TerraformAccTest"),
					resource.TestCheckResourceAttr(resourceName, "new_game_session_protection_policy", "FullProtection"),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "WINDOWS_2012"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.0.new_game_sessions_per_creator", "4"),
					resource.TestCheckResourceAttr(resourceName, "resource_creation_limit_policy.0.policy_period_in_minutes", "25"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.game_session_activation_timeout_seconds", "35"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.max_concurrent_game_session_activations", "98"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.concurrent_executions", "1"),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.launch_path", launchPath),
					resource.TestCheckResourceAttr(resourceName, "runtime_configuration.0.server_process.0.parameters", params[1]),
				),
			},
		},
	})
}

func TestAccAWSGameliftFleet_disappears(t *testing.T) {
	var conf gamelift.FleetAttributes

	fleetName := acctest.RandomWithPrefix("tf-acc-fleet")
	buildName := acctest.RandomWithPrefix("tf-acc-build")

	region := testAccGetRegion()
	g, err := testAccAWSGameliftSampleGame(region)

	if isResourceNotFoundError(err) {
		t.Skip(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	loc := g.Location
	bucketName := *loc.Bucket
	roleArn := *loc.RoleArn
	key := *loc.Key

	launchPath := g.LaunchPath
	params := g.Parameters(33435)
	resourceName := "aws_gamelift_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(gamelift.EndpointsID, t)
			testAccPreCheckAWSGamelift(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftFleetBasicConfig(fleetName, launchPath, params, buildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftFleetExists(resourceName, &conf),
					testAccCheckAWSGameliftFleetDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
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

func testAccCheckAWSGameliftFleetDisappears(res *gamelift.FleetAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).gameliftconn

		input := &gamelift.DeleteFleetInput{FleetId: res.FleetId}
		err := resource.Retry(60*time.Minute, func() *resource.RetryError {
			_, err := conn.DeleteFleet(input)
			if err != nil {
				msg := fmt.Sprintf("Cannot delete fleet %s that is in status of ", *res.FleetId)
				if isAWSErr(err, gamelift.ErrCodeInvalidRequestException, msg) {
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if isResourceTimeoutError(err) {
			_, err = conn.DeleteFleet(input)
		}
		if err != nil {
			return fmt.Errorf("Error deleting Gamelift fleet: %s", err)
		}

		return waitForGameliftFleetToBeDeleted(conn, *res.FleetId, 15*time.Minute)
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
	return testAccAWSGameliftFleetBasicTemplate(buildName, bucketName, key, roleArn) + fmt.Sprintf(`
resource "aws_gamelift_fleet" "test" {
  build_id          = aws_gamelift_build.test.id
  ec2_instance_type = "c4.large"
  name              = %[1]q

  runtime_configuration {
    server_process {
      concurrent_executions = 1
      launch_path           = %[2]q
      parameters            = %[3]q
    }
  }
}
`, fleetName, launchPath, params)
}

func testAccAWSGameliftFleetBasicConfigTags1(fleetName, launchPath, params, buildName, bucketName, key, roleArn, tagKey1, tagValue1 string) string {
	return testAccAWSGameliftFleetBasicTemplate(buildName, bucketName, key, roleArn) + fmt.Sprintf(`
resource "aws_gamelift_fleet" "test" {
  build_id          = aws_gamelift_build.test.id
  ec2_instance_type = "c4.large"
  name              = %[1]q

  runtime_configuration {
    server_process {
      concurrent_executions = 1
      launch_path           = %[2]q
      parameters            = %[3]q
    }
  }

  tags = {
    %[4]q = %[5]q
  }
}
`, fleetName, launchPath, params, tagKey1, tagValue1)
}

func testAccAWSGameliftFleetBasicConfigTags2(fleetName, launchPath, params, buildName, bucketName, key, roleArn, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSGameliftFleetBasicTemplate(buildName, bucketName, key, roleArn) + fmt.Sprintf(`
resource "aws_gamelift_fleet" "test" {
  build_id          = aws_gamelift_build.test.id
  ec2_instance_type = "c4.large"
  name              = %[1]q

  runtime_configuration {
    server_process {
      concurrent_executions = 1
      launch_path           = %[2]q
      parameters            = %[3]q
    }
  }

  tags = {
    %[4]q = %[5]q
    %[6]q = %[7]q
  }
}
`, fleetName, launchPath, params, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSGameliftFleetBasicUpdatedConfig(desc, fleetName, launchPath, params, buildName, bucketName, key, roleArn string) string {
	return testAccAWSGameliftFleetBasicTemplate(buildName, bucketName, key, roleArn) + fmt.Sprintf(`
resource "aws_gamelift_fleet" "test" {
  build_id                           = aws_gamelift_build.test.id
  ec2_instance_type                  = "c4.large"
  description                        = "%s"
  name                               = "%s"
  metric_groups                      = ["UpdatedGroup"]
  new_game_session_protection_policy = "FullProtection"

  resource_creation_limit_policy {
    new_game_sessions_per_creator = 2
    policy_period_in_minutes      = 15
  }

  runtime_configuration {
    server_process {
      concurrent_executions = 1
      launch_path           = %q
      parameters            = "%s"
    }
  }
}
`, desc, fleetName, launchPath, params)
}

func testAccAWSGameliftFleetAllFieldsConfig(fleetName, desc, launchPath string, params string, buildName, bucketName, key, roleArn string) string {
	return testAccAWSGameliftFleetBasicTemplate(buildName, bucketName, key, roleArn) +
		testAccAWSGameLiftFleetIAMRole(buildName) + fmt.Sprintf(`
resource "aws_gamelift_fleet" "test" {
  build_id          = aws_gamelift_build.test.id
  ec2_instance_type = "c4.large"
  name              = "%s"
  description       = "%s"
  instance_role_arn = aws_iam_role.test.arn
  fleet_type        = "ON_DEMAND"

  ec2_inbound_permission {
    from_port = 8080
    ip_range  = "8.8.8.8/32"
    protocol  = "TCP"
    to_port   = 8080
  }

  ec2_inbound_permission {
    from_port = 8443
    ip_range  = "8.8.0.0/16"
    protocol  = "TCP"
    to_port   = 8443
  }

  ec2_inbound_permission {
    from_port = 60000
    ip_range  = "8.8.8.8/32"
    protocol  = "UDP"
    to_port   = 60000
  }

  metric_groups                      = ["TerraformAccTest"]
  new_game_session_protection_policy = "FullProtection"

  resource_creation_limit_policy {
    new_game_sessions_per_creator = 4
    policy_period_in_minutes      = 25
  }

  runtime_configuration {
    game_session_activation_timeout_seconds = 35
    max_concurrent_game_session_activations = 99

    server_process {
      concurrent_executions = 1
      launch_path           = %q
      parameters            = "%s"
    }
  }
}
`, fleetName, desc, launchPath, params)
}

func testAccAWSGameliftFleetAllFieldsUpdatedConfig(fleetName, desc, launchPath string, params string, buildName, bucketName, key, roleArn string) string {
	return testAccAWSGameliftFleetBasicTemplate(buildName, bucketName, key, roleArn) +
		testAccAWSGameLiftFleetIAMRole(buildName) + fmt.Sprintf(`
resource "aws_gamelift_fleet" "test" {
  build_id          = aws_gamelift_build.test.id
  ec2_instance_type = "c4.large"

  name              = "%s"
  description       = "%s"
  instance_role_arn = aws_iam_role.test.arn
  fleet_type        = "ON_DEMAND"

  ec2_inbound_permission {
    from_port = 8888
    ip_range  = "8.8.8.8/32"
    protocol  = "TCP"
    to_port   = 8888
  }

  ec2_inbound_permission {
    from_port = 8443
    ip_range  = "8.4.0.0/16"
    protocol  = "TCP"
    to_port   = 8443
  }

  ec2_inbound_permission {
    from_port = 60000
    ip_range  = "8.8.8.8/32"
    protocol  = "UDP"
    to_port   = 60000
  }

  metric_groups                      = ["TerraformAccTest"]
  new_game_session_protection_policy = "FullProtection"

  resource_creation_limit_policy {
    new_game_sessions_per_creator = 4
    policy_period_in_minutes      = 25
  }

  runtime_configuration {
    game_session_activation_timeout_seconds = 35
    max_concurrent_game_session_activations = 98

    server_process {
      concurrent_executions = 1
      launch_path           = %q
      parameters            = "%s"
    }
  }
}
`, fleetName, desc, launchPath, params)
}

func testAccAWSGameliftFleetBasicTemplate(buildName, bucketName, key, roleArn string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_build" "test" {
  name             = %[1]q
  operating_system = "WINDOWS_2012"

  storage_location {
    bucket   = %[2]q
    key      = %[3]q
    role_arn = %[4]q
  }
}
`, buildName, bucketName, key, roleArn)
}

func testAccAWSGameLiftFleetIAMRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = "test-role-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "gamelift.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test" {
  name        = "test-policy-%[1]s"
  path        = "/"
  description = "GameLift Fleet PassRole Policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iam:PassRole",
        "sts:AssumeRole"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_policy_attachment" "test-attach" {
  name       = "test-attachment-%[1]s"
  roles      = [aws_iam_role.test.name]
  policy_arn = aws_iam_policy.test.arn
}
`, rName)
}
