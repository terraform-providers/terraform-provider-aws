package aws

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_autoscaling_group", &resource.Sweeper{
		Name: "aws_autoscaling_group",
		F:    testSweepAutoscalingGroups,
	})
}

func testSweepAutoscalingGroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).autoscalingconn

	resp, err := conn.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping AutoScaling Group sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving AutoScaling Groups in Sweeper: %s", err)
	}

	if len(resp.AutoScalingGroups) == 0 {
		log.Print("[DEBUG] No aws autoscaling groups to sweep")
		return nil
	}

	for _, asg := range resp.AutoScalingGroups {
		var testOptGroup bool
		for _, testName := range []string{"foobar", "terraform-", "tf-test", "tf-asg-"} {
			if strings.HasPrefix(*asg.AutoScalingGroupName, testName) {
				testOptGroup = true
				break
			}
		}

		if !testOptGroup {
			continue
		}

		deleteopts := autoscaling.DeleteAutoScalingGroupInput{
			AutoScalingGroupName: asg.AutoScalingGroupName,
			ForceDelete:          aws.Bool(true),
		}

		err = resource.Retry(5*time.Minute, func() *resource.RetryError {
			if _, err := conn.DeleteAutoScalingGroup(&deleteopts); err != nil {
				if awserr, ok := err.(awserr.Error); ok {
					switch awserr.Code() {
					case "InvalidGroup.NotFound":
						return nil
					case "ResourceInUse", "ScalingActivityInProgress":
						return resource.RetryableError(awserr)
					}
				}

				// Didn't recognize the error, so shouldn't retry.
				return resource.NonRetryableError(err)
			}
			// Successful delete
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func TestAccAWSAutoScalingGroup_importBasic(t *testing.T) {
	resourceName := "aws_autoscaling_group.bar"
	randName := fmt.Sprintf("terraform-test-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupImport(randName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete", "metrics_granularity", "wait_for_capacity_timeout"},
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_basic(t *testing.T) {
	var group autoscaling.Group
	var lc autoscaling.LaunchConfiguration

	randName := fmt.Sprintf("terraform-test-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   "aws_autoscaling_group.bar",
		IDRefreshIgnore: []string{"force_delete", "metrics_granularity", "wait_for_capacity_timeout"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckAWSAutoScalingGroupHealthyCapacity(&group, 2),
					testAccCheckAWSAutoScalingGroupAttributes(&group, randName),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "availability_zones.2487133097", "us-west-2a"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "name", randName),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "max_size", "5"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "min_size", "2"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "health_check_grace_period", "300"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "health_check_type", "ELB"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "desired_capacity", "4"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "force_delete", "true"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "termination_policies.0", "OldestInstance"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "termination_policies.1", "ClosestToNextInstanceHour"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "protect_from_scale_in", "false"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "enabled_metrics.#", "0"),
				),
			},

			{
				Config: testAccAWSAutoScalingGroupConfigUpdate(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckAWSLaunchConfigurationExists("aws_launch_configuration.new", &lc),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "desired_capacity", "5"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "termination_policies.0", "ClosestToNextInstanceHour"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "protect_from_scale_in", "true"),
					testLaunchConfigurationName("aws_autoscaling_group.bar", &lc),
					testAccCheckAutoscalingTags(&group.Tags, "FromTags1Changed", map[string]interface{}{
						"value":               "value1changed",
						"propagate_at_launch": true,
					}),
					testAccCheckAutoscalingTags(&group.Tags, "FromTags2", map[string]interface{}{
						"value":               "value2changed",
						"propagate_at_launch": true,
					}),
					testAccCheckAutoscalingTags(&group.Tags, "FromTags3", map[string]interface{}{
						"value":               "value3",
						"propagate_at_launch": true,
					}),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_namePrefix(t *testing.T) {
	nameRegexp := regexp.MustCompile("^tf-test-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"aws_autoscaling_group.test", "name", nameRegexp),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.test", "arn"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_autoGeneratedName(t *testing.T) {
	asgNameRegexp := regexp.MustCompile("^tf-asg-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_autoGeneratedName,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"aws_autoscaling_group.bar", "name", asgNameRegexp),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.bar", "arn"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_terminationPolicies(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_terminationPoliciesEmpty,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "termination_policies.#", "0"),
				),
			},

			{
				Config: testAccAWSAutoScalingGroupConfig_terminationPoliciesUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "termination_policies.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "termination_policies.0", "OldestInstance"),
				),
			},

			{
				Config: testAccAWSAutoScalingGroupConfig_terminationPoliciesExplicitDefault,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "termination_policies.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "termination_policies.0", "Default"),
				),
			},

			{
				Config: testAccAWSAutoScalingGroupConfig_terminationPoliciesEmpty,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "termination_policies.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_tags(t *testing.T) {
	var group autoscaling.Group

	randName := fmt.Sprintf("tf-test-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckAutoscalingTags(&group.Tags, "FromTags1", map[string]interface{}{
						"value":               "value1",
						"propagate_at_launch": true,
					}),
					testAccCheckAutoscalingTags(&group.Tags, "FromTags2", map[string]interface{}{
						"value":               "value2",
						"propagate_at_launch": true,
					}),
					testAccCheckAutoscalingTags(&group.Tags, "FromTags3", map[string]interface{}{
						"value":               "value3",
						"propagate_at_launch": true,
					}),
				),
			},

			{
				Config: testAccAWSAutoScalingGroupConfigUpdate(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckAutoscalingTagNotExists(&group.Tags, "Foo"),
					testAccCheckAutoscalingTags(&group.Tags, "FromTags1Changed", map[string]interface{}{
						"value":               "value1changed",
						"propagate_at_launch": true,
					}),
					testAccCheckAutoscalingTags(&group.Tags, "FromTags2", map[string]interface{}{
						"value":               "value2changed",
						"propagate_at_launch": true,
					}),
					testAccCheckAutoscalingTags(&group.Tags, "FromTags3", map[string]interface{}{
						"value":               "value3",
						"propagate_at_launch": true,
					}),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_VpcUpdates(t *testing.T) {
	var group autoscaling.Group

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfigWithAZ,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "availability_zones.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "availability_zones.2487133097", "us-west-2a"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "vpc_zone_identifier.#", "0"),
				),
			},

			{
				Config: testAccAWSAutoScalingGroupConfigWithVPCIdent,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckAWSAutoScalingGroupAttributesVPCZoneIdentifer(&group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "availability_zones.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "availability_zones.2487133097", "us-west-2a"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "vpc_zone_identifier.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_WithLoadBalancer(t *testing.T) {
	var group autoscaling.Group

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfigWithLoadBalancer,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckAWSAutoScalingGroupAttributesLoadBalancer(&group),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_withPlacementGroup(t *testing.T) {
	var group autoscaling.Group

	randName := fmt.Sprintf("tf-test-%s", acctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_withPlacementGroup(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "placement_group", randName),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_enablingMetrics(t *testing.T) {
	var group autoscaling.Group
	randName := fmt.Sprintf("terraform-test-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckNoResourceAttr(
						"aws_autoscaling_group.bar", "enabled_metrics"),
				),
			},

			{
				Config: testAccAWSAutoscalingMetricsCollectionConfig_updatingMetricsCollected,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "enabled_metrics.#", "5"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_suspendingProcesses(t *testing.T) {
	var group autoscaling.Group
	randName := fmt.Sprintf("terraform-test-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "suspended_processes.#", "0"),
				),
			},
			{
				Config: testAccAWSAutoScalingGroupConfigWithSuspendedProcesses(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "suspended_processes.#", "2"),
				),
			},
			{
				Config: testAccAWSAutoScalingGroupConfigWithSuspendedProcessesUpdated(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "suspended_processes.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_withMetrics(t *testing.T) {
	var group autoscaling.Group

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoscalingMetricsCollectionConfig_allMetricsCollected,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "enabled_metrics.#", "7"),
				),
			},

			{
				Config: testAccAWSAutoscalingMetricsCollectionConfig_updatingMetricsCollected,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "enabled_metrics.#", "5"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_serviceLinkedRoleARN(t *testing.T) {
	var group autoscaling.Group

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_withServiceLinkedRoleARN,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.bar", "service_linked_role_arn"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_ALB_TargetGroups(t *testing.T) {
	var group autoscaling.Group
	var tg elbv2.TargetGroup
	var tg2 elbv2.TargetGroup

	testCheck := func(targets []*elbv2.TargetGroup) resource.TestCheckFunc {
		return func(*terraform.State) error {
			var ts []string
			var gs []string
			for _, t := range targets {
				ts = append(ts, *t.TargetGroupArn)
			}

			for _, s := range group.TargetGroupARNs {
				gs = append(gs, *s)
			}

			sort.Strings(ts)
			sort.Strings(gs)

			if !reflect.DeepEqual(ts, gs) {
				return fmt.Errorf("Error: target group match not found!\nASG Target groups: %#v\nTarget Group: %#v", ts, gs)
			}
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_ALB_TargetGroup_pre,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &tg),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "target_group_arns.#", "0"),
				),
			},

			{
				Config: testAccAWSAutoScalingGroupConfig_ALB_TargetGroup_post_duo,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &tg),
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test_more", &tg2),
					testCheck([]*elbv2.TargetGroup{&tg, &tg2}),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "target_group_arns.#", "2"),
				),
			},

			{
				Config: testAccAWSAutoScalingGroupConfig_ALB_TargetGroup_post,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &tg),
					testCheck([]*elbv2.TargetGroup{&tg}),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "target_group_arns.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_initialLifecycleHook(t *testing.T) {
	var group autoscaling.Group

	randName := fmt.Sprintf("terraform-test-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:        func() { testAccPreCheck(t) },
		IDRefreshName:   "aws_autoscaling_group.bar",
		IDRefreshIgnore: []string{"force_delete", "metrics_granularity", "wait_for_capacity_timeout"},
		Providers:       testAccProviders,
		CheckDestroy:    testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupWithHookConfig(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckAWSAutoScalingGroupHealthyCapacity(&group, 2),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "initial_lifecycle_hook.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "initial_lifecycle_hook.391359060.default_result", "CONTINUE"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "initial_lifecycle_hook.391359060.name", "launching"),
					testAccCheckAWSAutoScalingGroupInitialLifecycleHookExists(
						"aws_autoscaling_group.bar", "initial_lifecycle_hook.391359060.name"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_ALB_TargetGroups_ELBCapacity(t *testing.T) {
	var group autoscaling.Group
	var tg elbv2.TargetGroup

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_ALB_TargetGroup_ELBCapacity(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &tg),
					testAccCheckAWSALBTargetGroupHealthy(&tg),
				),
			},
		},
	})
}

func testAccCheckAWSAutoScalingGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).autoscalingconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_autoscaling_group" {
			continue
		}

		// Try to find the Group
		describeGroups, err := conn.DescribeAutoScalingGroups(
			&autoscaling.DescribeAutoScalingGroupsInput{
				AutoScalingGroupNames: []*string{aws.String(rs.Primary.ID)},
			})

		if err == nil {
			if len(describeGroups.AutoScalingGroups) != 0 &&
				*describeGroups.AutoScalingGroups[0].AutoScalingGroupName == rs.Primary.ID {
				return fmt.Errorf("AutoScaling Group still exists")
			}
		}

		// Verify the error
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidGroup.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckAWSAutoScalingGroupAttributes(group *autoscaling.Group, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *group.AvailabilityZones[0] != "us-west-2a" {
			return fmt.Errorf("Bad availability_zones: %#v", group.AvailabilityZones[0])
		}

		if *group.AutoScalingGroupName != name {
			return fmt.Errorf("Bad Autoscaling Group name, expected (%s), got (%s)", name, *group.AutoScalingGroupName)
		}

		if *group.MaxSize != 5 {
			return fmt.Errorf("Bad max_size: %d", *group.MaxSize)
		}

		if *group.MinSize != 2 {
			return fmt.Errorf("Bad max_size: %d", *group.MinSize)
		}

		if *group.HealthCheckType != "ELB" {
			return fmt.Errorf("Bad health_check_type,\nexpected: %s\ngot: %s", "ELB", *group.HealthCheckType)
		}

		if *group.HealthCheckGracePeriod != 300 {
			return fmt.Errorf("Bad health_check_grace_period: %d", *group.HealthCheckGracePeriod)
		}

		if *group.DesiredCapacity != 4 {
			return fmt.Errorf("Bad desired_capacity: %d", *group.DesiredCapacity)
		}

		if *group.LaunchConfigurationName == "" {
			return fmt.Errorf("Bad launch configuration name: %s", *group.LaunchConfigurationName)
		}

		t := &autoscaling.TagDescription{
			Key:               aws.String("FromTags1"),
			Value:             aws.String("value1"),
			PropagateAtLaunch: aws.Bool(true),
			ResourceType:      aws.String("auto-scaling-group"),
			ResourceId:        group.AutoScalingGroupName,
		}

		if !reflect.DeepEqual(group.Tags[0], t) {
			return fmt.Errorf(
				"Got:\n\n%#v\n\nExpected:\n\n%#v\n",
				group.Tags[0],
				t)
		}

		return nil
	}
}

func testAccCheckAWSAutoScalingGroupAttributesLoadBalancer(group *autoscaling.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(group.LoadBalancerNames) != 1 {
			return fmt.Errorf("Bad load_balancers: %v", group.LoadBalancerNames)
		}

		return nil
	}
}

func testAccCheckAWSAutoScalingGroupExists(n string, group *autoscaling.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No AutoScaling Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).autoscalingconn

		describeGroups, err := conn.DescribeAutoScalingGroups(
			&autoscaling.DescribeAutoScalingGroupsInput{
				AutoScalingGroupNames: []*string{aws.String(rs.Primary.ID)},
			})

		if err != nil {
			return err
		}

		if len(describeGroups.AutoScalingGroups) != 1 ||
			*describeGroups.AutoScalingGroups[0].AutoScalingGroupName != rs.Primary.ID {
			return fmt.Errorf("AutoScaling Group not found")
		}

		*group = *describeGroups.AutoScalingGroups[0]

		return nil
	}
}

func testAccCheckAWSAutoScalingGroupInitialLifecycleHookExists(asg, hookAttr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		asgResource, ok := s.RootModule().Resources[asg]
		if !ok {
			return fmt.Errorf("Not found: %s", asg)
		}

		if asgResource.Primary.ID == "" {
			return fmt.Errorf("No AutoScaling Group ID is set")
		}

		hookName := asgResource.Primary.Attributes[hookAttr]
		if hookName == "" {
			return fmt.Errorf("ASG %s has no hook name %s", asg, hookAttr)
		}

		return checkLifecycleHookExistsByName(asgResource.Primary.ID, hookName)
	}
}

func testLaunchConfigurationName(n string, lc *autoscaling.LaunchConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if *lc.LaunchConfigurationName != rs.Primary.Attributes["launch_configuration"] {
			return fmt.Errorf("Launch configuration names do not match")
		}

		return nil
	}
}

func testAccCheckAWSAutoScalingGroupHealthyCapacity(
	g *autoscaling.Group, exp int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		healthy := 0
		for _, i := range g.Instances {
			if i.HealthStatus == nil {
				continue
			}
			if strings.EqualFold(*i.HealthStatus, "Healthy") {
				healthy++
			}
		}
		if healthy < exp {
			return fmt.Errorf("Expected at least %d healthy, got %d.", exp, healthy)
		}
		return nil
	}
}

func testAccCheckAWSAutoScalingGroupAttributesVPCZoneIdentifer(group *autoscaling.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Grab Subnet Ids
		var subnets []string
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_subnet" {
				continue
			}
			subnets = append(subnets, rs.Primary.Attributes["id"])
		}

		if group.VPCZoneIdentifier == nil {
			return fmt.Errorf("Bad VPC Zone Identifier\nexpected: %s\ngot nil", subnets)
		}

		zones := strings.Split(*group.VPCZoneIdentifier, ",")

		remaining := len(zones)
		for _, z := range zones {
			for _, s := range subnets {
				if z == s {
					remaining--
				}
			}
		}

		if remaining != 0 {
			return fmt.Errorf("Bad VPC Zone Identifier match\nexpected: %s\ngot:%s", zones, subnets)
		}

		return nil
	}
}

// testAccCheckAWSALBTargetGroupHealthy checks an *elbv2.TargetGroup to make
// sure that all instances in it are healthy.
func testAccCheckAWSALBTargetGroupHealthy(res *elbv2.TargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).elbv2conn

		resp, err := conn.DescribeTargetHealth(&elbv2.DescribeTargetHealthInput{
			TargetGroupArn: res.TargetGroupArn,
		})

		if err != nil {
			return err
		}

		for _, target := range resp.TargetHealthDescriptions {
			if target.TargetHealth == nil || target.TargetHealth.State == nil || *target.TargetHealth.State != "healthy" {
				return errors.New("Not all instances in target group are healthy yet, but should be")
			}
		}

		return nil
	}
}

func TestAccAWSAutoScalingGroup_classicVpcZoneIdentifier(t *testing.T) {
	var group autoscaling.Group

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_autoscaling_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_classicVpcZoneIdentifier,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.test", &group),
					resource.TestCheckResourceAttr("aws_autoscaling_group.test", "vpc_zone_identifier.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_emptyAvailabilityZones(t *testing.T) {
	var group autoscaling.Group

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_autoscaling_group.test",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_emptyAvailabilityZones,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.test", &group),
					resource.TestCheckResourceAttr("aws_autoscaling_group.test", "availability_zones.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_launchTemplate(t *testing.T) {
	var group autoscaling.Group

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_withLaunchTemplate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.bar", "launch_template.0.id"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_launchTemplate_update(t *testing.T) {
	var group autoscaling.Group

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_withLaunchTemplate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.bar", "launch_template.0.name"),
				),
			},

			{
				Config: testAccAWSAutoScalingGroupConfig_withLaunchTemplate_toLaunchConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.bar", "launch_configuration"),
					resource.TestCheckNoResourceAttr(
						"aws_autoscaling_group.bar", "launch_template"),
				),
			},

			{
				Config: testAccAWSAutoScalingGroupConfig_withLaunchTemplate_toLaunchTemplateName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "launch_configuration", ""),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "launch_template.0.name", "foobar2"),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.bar", "launch_template.0.id"),
				),
			},

			{
				Config: testAccAWSAutoScalingGroupConfig_withLaunchTemplate_toLaunchTemplateVersion,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "launch_template.0.version", "$Latest"),
				),
			},

			{
				Config: testAccAWSAutoScalingGroupConfig_withLaunchTemplate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.bar", "launch_template.0.name"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "launch_template.0.version", "1"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_LaunchTemplate_IAMInstanceProfile(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_LaunchTemplate_IAMInstanceProfile(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_MixedInstancesPolicy(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.version", "$Default"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t3.small"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"metrics_granularity",
					"wait_for_capacity_timeout",
				},
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_MixedInstancesPolicy_InstancesDistribution_OnDemandAllocationStrategy(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandAllocationStrategy(rName, "prioritized"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_allocation_strategy", "prioritized"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"metrics_granularity",
					"wait_for_capacity_timeout",
				},
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_MixedInstancesPolicy_InstancesDistribution_OnDemandBaseCapacity(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandBaseCapacity(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"metrics_granularity",
					"wait_for_capacity_timeout",
				},
			},
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandBaseCapacity(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity", "2"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_MixedInstancesPolicy_InstancesDistribution_OnDemandPercentageAboveBaseCapacity(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandPercentageAboveBaseCapacity(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_percentage_above_base_capacity", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"metrics_granularity",
					"wait_for_capacity_timeout",
				},
			},
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandPercentageAboveBaseCapacity(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_percentage_above_base_capacity", "2"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_MixedInstancesPolicy_InstancesDistribution_SpotAllocationStrategy(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_SpotAllocationStrategy(rName, "lowest-price"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_allocation_strategy", "lowest-price"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"metrics_granularity",
					"wait_for_capacity_timeout",
				},
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_MixedInstancesPolicy_InstancesDistribution_SpotInstancePools(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_SpotInstancePools(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_instance_pools", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"metrics_granularity",
					"wait_for_capacity_timeout",
				},
			},
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_SpotInstancePools(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_instance_pools", "3"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_MixedInstancesPolicy_InstancesDistribution_SpotMaxPrice(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_SpotMaxPrice(rName, "0.50"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_max_price", "0.50"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"metrics_granularity",
					"wait_for_capacity_timeout",
				},
			},
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_SpotMaxPrice(rName, "0.51"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_max_price", "0.51"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_MixedInstancesPolicy_LaunchTemplate_LaunchTemplateSpecification_LaunchTemplateName(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_LaunchTemplate_LaunchTemplateSpecification_LaunchTemplateName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.launch_template_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"metrics_granularity",
					"wait_for_capacity_timeout",
				},
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_MixedInstancesPolicy_LaunchTemplate_LaunchTemplateSpecification_Version(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_LaunchTemplate_LaunchTemplateSpecification_Version(rName, "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"metrics_granularity",
					"wait_for_capacity_timeout",
				},
			},
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_LaunchTemplate_LaunchTemplateSpecification_Version(rName, "$$Latest"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.version", "$Latest"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_MixedInstancesPolicy_LaunchTemplate_Override_InstanceType(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_InstanceType(rName, "t3.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t3.small"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"metrics_granularity",
					"wait_for_capacity_timeout",
				},
			},
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_InstanceType(rName, "t3.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t3.medium"),
				),
			},
		},
	})
}

const testAccAWSAutoScalingGroupConfig_autoGeneratedName = `
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["us-west-2a"]
  desired_capacity = 0
  max_size = 0
  min_size = 0
  launch_configuration = "${aws_launch_configuration.foobar.name}"
}
`

const testAccAWSAutoScalingGroupConfig_namePrefix = `
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "test" {
  availability_zones = ["us-west-2a"]
  desired_capacity = 0
  max_size = 0
  min_size = 0
  name_prefix = "tf-test-"
  launch_configuration = "${aws_launch_configuration.test.name}"
}
`

const testAccAWSAutoScalingGroupConfig_terminationPoliciesEmpty = `
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["us-west-2a"]
  max_size = 0
  min_size = 0
  desired_capacity = 0

  launch_configuration = "${aws_launch_configuration.foobar.name}"
}
`

const testAccAWSAutoScalingGroupConfig_terminationPoliciesExplicitDefault = `
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["us-west-2a"]
  max_size = 0
  min_size = 0
  desired_capacity = 0
  termination_policies = ["Default"]

  launch_configuration = "${aws_launch_configuration.foobar.name}"
}
`

const testAccAWSAutoScalingGroupConfig_terminationPoliciesUpdate = `
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["us-west-2a"]
  max_size = 0
  min_size = 0
  desired_capacity = 0
  termination_policies = ["OldestInstance"]

  launch_configuration = "${aws_launch_configuration.foobar.name}"
}
`

func testAccAWSAutoScalingGroupConfig(name string) string {
	return fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_placement_group" "test" {
  name = "asg_pg_%s"
  strategy = "cluster"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["us-west-2a"]
  name = "%s"
  max_size = 5
  min_size = 2
  health_check_type = "ELB"
  desired_capacity = 4
  force_delete = true
  termination_policies = ["OldestInstance","ClosestToNextInstanceHour"]

  launch_configuration = "${aws_launch_configuration.foobar.name}"

  tags = [
    {
      key = "FromTags1"
      value = "value1"
      propagate_at_launch = true
    },
    {
      key = "FromTags2"
      value = "value2"
      propagate_at_launch = true
    },
    {
      key = "FromTags3"
      value = "value3"
      propagate_at_launch = true
    },
  ]
}
`, name, name)
}

func testAccAWSAutoScalingGroupConfigUpdate(name string) string {
	return fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_launch_configuration" "new" {
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["us-west-2a"]
  name = "%s"
  max_size = 5
  min_size = 2
  health_check_grace_period = 300
  health_check_type = "ELB"
  desired_capacity = 5
  force_delete = true
  termination_policies = ["ClosestToNextInstanceHour"]
  protect_from_scale_in = true

  launch_configuration = "${aws_launch_configuration.new.name}"

  tags = [
    {
      key = "FromTags1Changed"
      value = "value1changed"
      propagate_at_launch = true
    },
    {
      key = "FromTags2"
      value = "value2changed"
      propagate_at_launch = true
    },
    {
      key = "FromTags3"
      value = "value3"
      propagate_at_launch = true
    },
  ]
}
`, name)
}

func testAccAWSAutoScalingGroupImport(name string) string {
	return fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_placement_group" "test" {
  name = "asg_pg_%s"
  strategy = "cluster"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["us-west-2a"]
  name = "%s"
  max_size = 5
  min_size = 2
  health_check_type = "ELB"
  desired_capacity = 4
  force_delete = true
  termination_policies = ["OldestInstance","ClosestToNextInstanceHour"]

  launch_configuration = "${aws_launch_configuration.foobar.name}"

  tag {
    key = "FromTags1"
    value = "value1"
    propagate_at_launch = true
  }

  tag {
    key = "FromTags2"
    value = "value2"
    propagate_at_launch = true
  }

  tag {
    key = "FromTags3"
    value = "value3"
    propagate_at_launch = true
  }
}
`, name, name)
}

const testAccAWSAutoScalingGroupConfigWithLoadBalancer = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-autoscaling-group-with-lb"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	tags = {
		Name = "tf-acc-autoscaling-group-with-load-balancer"
	}
}

resource "aws_security_group" "foo" {
  vpc_id="${aws_vpc.foo.id}"

  ingress {
    protocol = "-1"
    from_port = 0
    to_port = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    protocol = "-1"
    from_port = 0
    to_port = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elb" "bar" {
  subnets = ["${aws_subnet.foo.id}"]
	security_groups = ["${aws_security_group.foo.id}"]

  listener {
    instance_port = 80
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }

  health_check {
    healthy_threshold = 2
    unhealthy_threshold = 2
    target = "HTTP:80/"
    interval = 5
    timeout = 2
  }

	depends_on = ["aws_internet_gateway.gw"]
}

// need an AMI that listens on :80 at boot, this is:
data "aws_ami" "test_ami" {
  most_recent = true

  owners     = ["979382823631"]

  filter {
    name   = "name"
    values = ["bitnami-nginxstack-*-linux-debian-9-x86_64-hvm-ebs"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
	security_groups = ["${aws_security_group.foo.id}"]
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["${aws_subnet.foo.availability_zone}"]
  vpc_zone_identifier = ["${aws_subnet.foo.id}"]
  max_size = 2
  min_size = 2
  health_check_grace_period = 300
  health_check_type = "ELB"
  wait_for_elb_capacity = 2
  force_delete = true

  launch_configuration = "${aws_launch_configuration.foobar.name}"
  load_balancers = ["${aws_elb.bar.name}"]
}
`

const testAccAWSAutoScalingGroupConfigWithAZ = `
resource "aws_vpc" "default" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-autoscaling-group-with-az"
  }
}

resource "aws_subnet" "main" {
  vpc_id = "${aws_vpc.default.id}"
  cidr_block = "10.0.1.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-autoscaling-group-with-az"
  }
}

data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = [
	  "us-west-2a"
  ]
  desired_capacity = 0
  max_size = 0
  min_size = 0
  launch_configuration = "${aws_launch_configuration.foobar.name}"
}
`

const testAccAWSAutoScalingGroupConfigWithVPCIdent = `
resource "aws_vpc" "default" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-autoscaling-group-with-vpc-id"
  }
}

resource "aws_subnet" "main" {
  vpc_id = "${aws_vpc.default.id}"
  cidr_block = "10.0.1.0/24"
  availability_zone = "us-west-2a"
  tags = {
    Name = "tf-acc-autoscaling-group-with-vpc-id"
  }
}

data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  vpc_zone_identifier = [
    "${aws_subnet.main.id}",
  ]
  desired_capacity = 0
  max_size = 0
  min_size = 0
  launch_configuration = "${aws_launch_configuration.foobar.name}"
}
`

func testAccAWSAutoScalingGroupConfig_withPlacementGroup(name string) string {
	return fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "c3.large"
}

resource "aws_placement_group" "test" {
  name = "%s"
  strategy = "cluster"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["us-west-2a"]
  name = "%s"
  max_size = 1
  min_size = 1
  health_check_grace_period = 300
  health_check_type = "ELB"
  desired_capacity = 1
  force_delete = true
  termination_policies = ["OldestInstance","ClosestToNextInstanceHour"]
  placement_group = "${aws_placement_group.test.name}"

  launch_configuration = "${aws_launch_configuration.foobar.name}"

  tag {
    key = "Foo"
    value = "foo-bar"
    propagate_at_launch = true
  }
}
`, name, name)
}

const testAccAWSAutoScalingGroupConfig_withServiceLinkedRoleARN = `
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_iam_role" "autoscaling_service_linked_role" {
  name = "AWSServiceRoleForAutoScaling"
}

resource "aws_launch_configuration" "foobar" {
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["us-west-2a"]
  desired_capacity = 0
  max_size = 0
  min_size = 0
  launch_configuration = "${aws_launch_configuration.foobar.name}"
  service_linked_role_arn = "${data.aws_iam_role.autoscaling_service_linked_role.arn}"
}
`

const testAccAWSAutoscalingMetricsCollectionConfig_allMetricsCollected = `
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["us-west-2a"]
  max_size = 1
  min_size = 0
  health_check_grace_period = 300
  health_check_type = "EC2"
  desired_capacity = 0
  force_delete = true
  termination_policies = ["OldestInstance","ClosestToNextInstanceHour"]
  launch_configuration = "${aws_launch_configuration.foobar.name}"
  enabled_metrics = ["GroupTotalInstances",
  	     "GroupPendingInstances",
  	     "GroupTerminatingInstances",
  	     "GroupDesiredCapacity",
  	     "GroupInServiceInstances",
  	     "GroupMinSize",
  	     "GroupMaxSize"
  ]
  metrics_granularity = "1Minute"
}
`

const testAccAWSAutoscalingMetricsCollectionConfig_updatingMetricsCollected = `
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["us-west-2a"]
  max_size = 1
  min_size = 0
  health_check_grace_period = 300
  health_check_type = "EC2"
  desired_capacity = 0
  force_delete = true
  termination_policies = ["OldestInstance","ClosestToNextInstanceHour"]
  launch_configuration = "${aws_launch_configuration.foobar.name}"
  enabled_metrics = ["GroupTotalInstances",
  	     "GroupPendingInstances",
  	     "GroupTerminatingInstances",
  	     "GroupDesiredCapacity",
  	     "GroupMaxSize"
  ]
  metrics_granularity = "1Minute"
}
`

const testAccAWSAutoScalingGroupConfig_ALB_TargetGroup_pre = `
resource "aws_vpc" "default" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-autoscaling-group-alb-target-group"
  }
}

resource "aws_lb_target_group" "test" {
  name     = "tf-example-alb-tg"
  port     = 80
  protocol = "HTTP"
  vpc_id   = "${aws_vpc.default.id}"
}

resource "aws_subnet" "main" {
  vpc_id            = "${aws_vpc.default.id}"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-acc-autoscaling-group-alb-target-group-main"
  }
}

resource "aws_subnet" "alt" {
  vpc_id            = "${aws_vpc.default.id}"
  cidr_block        = "10.0.2.0/24"
  availability_zone = "us-west-2b"

  tags = {
    Name = "tf-acc-autoscaling-group-alb-target-group-alt"
  }
}

data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id          = "${data.aws_ami.test_ami.id}"
  instance_type     = "t2.micro"
  enable_monitoring = false
}

resource "aws_autoscaling_group" "bar" {
  vpc_zone_identifier = [
    "${aws_subnet.main.id}",
    "${aws_subnet.alt.id}",
  ]

  max_size                  = 2
  min_size                  = 0
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 0
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = "${aws_launch_configuration.foobar.name}"

}

resource "aws_security_group" "tf_test_self" {
  name        = "tf_test_alb_asg"
  description = "tf_test_alb_asg"
  vpc_id      = "${aws_vpc.default.id}"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "testAccAWSAutoScalingGroupConfig_ALB_TargetGroup"
  }
}
`

const testAccAWSAutoScalingGroupConfig_ALB_TargetGroup_post = `
resource "aws_vpc" "default" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-autoscaling-group-alb-target-group"
  }
}

resource "aws_lb_target_group" "test" {
  name     = "tf-example-alb-tg"
  port     = 80
  protocol = "HTTP"
  vpc_id   = "${aws_vpc.default.id}"
}

resource "aws_subnet" "main" {
  vpc_id            = "${aws_vpc.default.id}"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-acc-autoscaling-group-alb-target-group-main"
  }
}

resource "aws_subnet" "alt" {
  vpc_id            = "${aws_vpc.default.id}"
  cidr_block        = "10.0.2.0/24"
  availability_zone = "us-west-2b"

  tags = {
    Name = "tf-acc-autoscaling-group-alb-target-group-alt"
  }
}

data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id          = "${data.aws_ami.test_ami.id}"
  instance_type     = "t2.micro"
  enable_monitoring = false
}

resource "aws_autoscaling_group" "bar" {
  vpc_zone_identifier = [
    "${aws_subnet.main.id}",
    "${aws_subnet.alt.id}",
  ]

	target_group_arns = ["${aws_lb_target_group.test.arn}"]

  max_size                  = 2
  min_size                  = 0
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 0
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = "${aws_launch_configuration.foobar.name}"

}

resource "aws_security_group" "tf_test_self" {
  name        = "tf_test_alb_asg"
  description = "tf_test_alb_asg"
  vpc_id      = "${aws_vpc.default.id}"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "testAccAWSAutoScalingGroupConfig_ALB_TargetGroup"
  }
}
`

const testAccAWSAutoScalingGroupConfig_ALB_TargetGroup_post_duo = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_vpc" "default" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-autoscaling-group-alb-target-group"
  }
}

resource "aws_lb_target_group" "test" {
  name     = "tf-example-alb-tg"
  port     = 80
  protocol = "HTTP"
  vpc_id   = "${aws_vpc.default.id}"
}

resource "aws_lb_target_group" "test_more" {
  name     = "tf-example-alb-tg-more"
  port     = 80
  protocol = "HTTP"
  vpc_id   = "${aws_vpc.default.id}"
}

resource "aws_subnet" "main" {
  vpc_id            = "${aws_vpc.default.id}"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-acc-autoscaling-group-alb-target-group-main"
  }
}

resource "aws_subnet" "alt" {
  vpc_id            = "${aws_vpc.default.id}"
  cidr_block        = "10.0.2.0/24"
  availability_zone = "us-west-2b"

  tags = {
    Name = "tf-acc-autoscaling-group-alb-target-group-alt"
  }
}

data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id          = "${data.aws_ami.test_ami.id}"
  instance_type     = "t2.micro"
  enable_monitoring = false
}

resource "aws_autoscaling_group" "bar" {
  vpc_zone_identifier = [
    "${aws_subnet.main.id}",
    "${aws_subnet.alt.id}",
  ]

	target_group_arns = [
		"${aws_lb_target_group.test.arn}",
		"${aws_lb_target_group.test_more.arn}",
	]

  max_size                  = 2
  min_size                  = 0
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 0
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = "${aws_launch_configuration.foobar.name}"

}

resource "aws_security_group" "tf_test_self" {
  name        = "tf_test_alb_asg"
  description = "tf_test_alb_asg"
  vpc_id      = "${aws_vpc.default.id}"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "testAccAWSAutoScalingGroupConfig_ALB_TargetGroup"
  }
}
`

func testAccAWSAutoScalingGroupWithHookConfig(name string) string {
	return fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["us-west-2a"]
  name = "%s"
  max_size = 5
  min_size = 2
  health_check_type = "ELB"
  desired_capacity = 4
  force_delete = true
  termination_policies = ["OldestInstance","ClosestToNextInstanceHour"]

  launch_configuration = "${aws_launch_configuration.foobar.name}"

  initial_lifecycle_hook {
    name = "launching"
    default_result = "CONTINUE"
    heartbeat_timeout = 30  # minimum value
    lifecycle_transition = "autoscaling:EC2_INSTANCE_LAUNCHING"
  }
}
`, name)
}

func testAccAWSAutoScalingGroupConfig_ALB_TargetGroup_ELBCapacity(rInt int) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-west-2"
}

resource "aws_vpc" "default" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = "true"
  enable_dns_support   = "true"

  tags = {
    Name = "terraform-testacc-autoscaling-group-alb-target-group-elb-capacity"
  }
}

resource "aws_lb" "test_lb" {
  subnets = ["${aws_subnet.main.id}", "${aws_subnet.alt.id}"]

  tags = {
    Name = "testAccAWSAutoScalingGroupConfig_ALB_TargetGroup_ELBCapacity"
  }
}

resource "aws_lb_listener" "test_listener" {
  load_balancer_arn = "${aws_lb.test_lb.arn}"
  port              = "80"

  default_action {
    target_group_arn = "${aws_lb_target_group.test.arn}"
    type             = "forward"
  }
}

resource "aws_lb_target_group" "test" {
  name     = "tf-alb-test-%d"
  port     = 80
  protocol = "HTTP"
  vpc_id   = "${aws_vpc.default.id}"

  health_check {
    path              = "/"
    healthy_threshold = "2"
    timeout           = "2"
    interval          = "5"
    matcher           = "200"
  }

  tags = {
    Name = "testAccAWSAutoScalingGroupConfig_ALB_TargetGroup_ELBCapacity"
  }
}

resource "aws_subnet" "main" {
  vpc_id            = "${aws_vpc.default.id}"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-west-2a"

  tags = {
    Name = "tf-acc-autoscaling-group-alb-target-group-elb-capacity-main"
  }
}

resource "aws_subnet" "alt" {
  vpc_id            = "${aws_vpc.default.id}"
  cidr_block        = "10.0.2.0/24"
  availability_zone = "us-west-2b"

  tags = {
    Name = "tf-acc-autoscaling-group-alb-target-group-elb-capacity-alt"
  }
}

resource "aws_internet_gateway" "internet_gateway" {
  vpc_id = "${aws_vpc.default.id}"
}

resource "aws_route_table" "route_table" {
  vpc_id = "${aws_vpc.default.id}"
}

resource "aws_route_table_association" "route_table_association_main" {
  subnet_id      = "${aws_subnet.main.id}"
  route_table_id = "${aws_route_table.route_table.id}"
}

resource "aws_route_table_association" "route_table_association_alt" {
  subnet_id      = "${aws_subnet.alt.id}"
  route_table_id = "${aws_route_table.route_table.id}"
}

resource "aws_route" "public_default_route" {
  route_table_id         = "${aws_route_table.route_table.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = "${aws_internet_gateway.internet_gateway.id}"
}

data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id                    = "${data.aws_ami.test_ami.id}"
  instance_type               = "t2.micro"
  associate_public_ip_address = "true"

  user_data = <<EOS
#!/bin/bash
yum -y install httpd
echo "hello world" > /var/www/html/index.html
chkconfig httpd on
service httpd start
EOS
}

resource "aws_autoscaling_group" "bar" {
  vpc_zone_identifier = [
    "${aws_subnet.main.id}",
    "${aws_subnet.alt.id}",
  ]

  target_group_arns = ["${aws_lb_target_group.test.arn}"]

  max_size                  = 2
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 2
  wait_for_elb_capacity     = 2
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = "${aws_launch_configuration.foobar.name}"
}`, rInt)
}

func testAccAWSAutoScalingGroupConfigWithSuspendedProcesses(name string) string {
	return fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_placement_group" "test" {
  name = "asg_pg_%s"
  strategy = "cluster"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["us-west-2a"]
  name = "%s"
  max_size = 5
  min_size = 2
  health_check_type = "ELB"
  desired_capacity = 4
  force_delete = true
  termination_policies = ["OldestInstance","ClosestToNextInstanceHour"]

  launch_configuration = "${aws_launch_configuration.foobar.name}"

  suspended_processes = ["AlarmNotification","ScheduledActions"]

  tag {
    key = "Foo"
    value = "foo-bar"
    propagate_at_launch = true
  }
}
`, name, name)
}

func testAccAWSAutoScalingGroupConfigWithSuspendedProcessesUpdated(name string) string {
	return fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_placement_group" "test" {
  name = "asg_pg_%s"
  strategy = "cluster"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["us-west-2a"]
  name = "%s"
  max_size = 5
  min_size = 2
  health_check_type = "ELB"
  desired_capacity = 4
  force_delete = true
  termination_policies = ["OldestInstance","ClosestToNextInstanceHour"]

  launch_configuration = "${aws_launch_configuration.foobar.name}"

  suspended_processes = ["AZRebalance","ScheduledActions"]

  tag {
    key = "Foo"
    value = "foo-bar"
    propagate_at_launch = true
  }
}
`, name, name)
}

const testAccAWSAutoScalingGroupConfig_classicVpcZoneIdentifier = `
resource "aws_autoscaling_group" "test" {
  min_size = 0
  max_size = 0

  availability_zones   = ["us-west-2a"]
  launch_configuration = "${aws_launch_configuration.test.name}"
  vpc_zone_identifier  = []
}

data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id      = "${data.aws_ami.test_ami.id}"
  instance_type = "t1.micro"
}
`

const testAccAWSAutoScalingGroupConfig_emptyAvailabilityZones = `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-autoscaling-group-empty-azs"
  }
}

resource "aws_subnet" "test" {
  vpc_id     = "${aws_vpc.test.id}"
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "tf-acc-autoscaling-group-empty-availability-zones"
  }
}

resource "aws_autoscaling_group" "test" {
  min_size = 0
  max_size = 0

  availability_zones   = []
  launch_configuration = "${aws_launch_configuration.test.name}"
  vpc_zone_identifier  = ["${aws_subnet.test.id}"]
}

data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id      = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}
`

const testAccAWSAutoScalingGroupConfig_withLaunchTemplate = `
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "foobar" {
  name_prefix = "foobar"
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

data "aws_availability_zones" "available" {}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity = 0
  max_size = 0
  min_size = 0
  launch_template {
    id = "${aws_launch_template.foobar.id}"
    version = "${aws_launch_template.foobar.default_version}"
  }
}
`

const testAccAWSAutoScalingGroupConfig_withLaunchTemplate_toLaunchConfig = `
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "foobar" {
  name_prefix = "foobar"
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_launch_configuration" "test" {
  image_id      = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

data "aws_availability_zones" "available" {}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity = 0
  max_size = 0
  min_size = 0
  launch_configuration = "${aws_launch_configuration.test.name}"
}
`

const testAccAWSAutoScalingGroupConfig_withLaunchTemplate_toLaunchTemplateName = `
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "foobar" {
  name_prefix = "foobar"
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_launch_configuration" "test" {
  image_id      = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_launch_template" "foobar2" {
  name = "foobar2"
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

data "aws_availability_zones" "available" {}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity = 0
  max_size = 0
  min_size = 0
  launch_template {
    name = "foobar2"
  }
}
`

const testAccAWSAutoScalingGroupConfig_withLaunchTemplate_toLaunchTemplateVersion = `
data "aws_ami" "test_ami" {
  most_recent = true

  filter {
    name   = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "foobar" {
  name_prefix = "foobar"
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_launch_configuration" "test" {
  image_id      = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

resource "aws_launch_template" "foobar2" {
  name = "foobar2"
  image_id = "${data.aws_ami.test_ami.id}"
  instance_type = "t2.micro"
}

data "aws_availability_zones" "available" {}

resource "aws_autoscaling_group" "bar" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity = 0
  max_size = 0
  min_size = 0
  launch_template {
    id = "${aws_launch_template.foobar.id}"
    version = "$$Latest"
  }
}
`

func testAccAWSAutoScalingGroupConfig_LaunchTemplate_IAMInstanceProfile(rName string) string {
	return fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_availability_zones" "available" {}

resource "aws_iam_role" "test" {
  assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"ec2.amazonaws.com\"]},\"Action\":[\"sts:AssumeRole\"]}]}"
  name               = %q
}

resource "aws_iam_instance_profile" "test" {
  name  = %q
  roles = ["${aws_iam_role.test.name}"]
}

resource "aws_launch_template" "test" {
  image_id      = "${data.aws_ami.test.id}"
  instance_type = "t2.micro"
  name          = %q

  iam_instance_profile {
    name = "${aws_iam_instance_profile.test.id}"
  }
}

resource "aws_autoscaling_group" "test" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  launch_template {
    id = "${aws_launch_template.test.id}"
  }
}
`, rName, rName, rName, rName)
}

func testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName string) string {
	return fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_availability_zones" "available" {}

resource "aws_launch_template" "test" {
  image_id      = "${data.aws_ami.test.id}"
  instance_type = "t3.micro"
  name          = %q
}
`, rName)
}

func testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy(rName string) string {
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) + fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = "${aws_launch_template.test.id}"
      }

      override {
        instance_type = "t2.micro"
      }
      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName)
}

func testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandAllocationStrategy(rName, onDemandAllocationStrategy string) string {
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) + fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    instances_distribution {
      on_demand_allocation_strategy = %q
    }

    launch_template {
      launch_template_specification {
        launch_template_id = "${aws_launch_template.test.id}"
      }

      override {
        instance_type = "t2.micro"
      }
      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, onDemandAllocationStrategy)
}

func testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandBaseCapacity(rName string, onDemandBaseCapacity int) string {
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) + fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity   = 0
  max_size           = 2
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    instances_distribution {
      on_demand_base_capacity = %d
    }

    launch_template {
      launch_template_specification {
        launch_template_id = "${aws_launch_template.test.id}"
      }

      override {
        instance_type = "t2.micro"
      }
      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, onDemandBaseCapacity)
}

func testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandPercentageAboveBaseCapacity(rName string, onDemandPercentageAboveBaseCapacity int) string {
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) + fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    instances_distribution {
      on_demand_percentage_above_base_capacity = %d
    }

    launch_template {
      launch_template_specification {
        launch_template_id = "${aws_launch_template.test.id}"
      }

      override {
        instance_type = "t2.micro"
      }
      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, onDemandPercentageAboveBaseCapacity)
}

func testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_SpotAllocationStrategy(rName, spotAllocationStrategy string) string {
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) + fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    instances_distribution {
      spot_allocation_strategy = %q
    }

    launch_template {
      launch_template_specification {
        launch_template_id = "${aws_launch_template.test.id}"
      }

      override {
        instance_type = "t2.micro"
      }
      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, spotAllocationStrategy)
}

func testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_SpotInstancePools(rName string, spotInstancePools int) string {
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) + fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    instances_distribution {
      spot_instance_pools = %d
    }

    launch_template {
      launch_template_specification {
        launch_template_id = "${aws_launch_template.test.id}"
      }

      override {
        instance_type = "t2.micro"
      }
      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, spotInstancePools)
}

func testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_SpotMaxPrice(rName, spotMaxPrice string) string {
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) + fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    instances_distribution {
      spot_max_price = %q
    }

    launch_template {
      launch_template_specification {
        launch_template_id = "${aws_launch_template.test.id}"
      }

      override {
        instance_type = "t2.micro"
      }
      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, spotMaxPrice)
}

func testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_LaunchTemplate_LaunchTemplateSpecification_LaunchTemplateName(rName string) string {
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) + fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_name = "${aws_launch_template.test.name}"
      }

      override {
        instance_type = "t2.micro"
      }
      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName)
}

func testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_LaunchTemplate_LaunchTemplateSpecification_Version(rName, version string) string {
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) + fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = "${aws_launch_template.test.id}"
        version            = %q
      }

      override {
        instance_type = "t2.micro"
      }
      override {
        instance_type = "t3.small"
      }
    }
  }
}
`, rName, version)
}

func testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_InstanceType(rName, instanceType string) string {
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) + fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = ["${data.aws_availability_zones.available.names[0]}"]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = "${aws_launch_template.test.id}"
      }

      override {
        instance_type = "t2.micro"
      }
      override {
        instance_type = %q
      }
    }
  }
}
`, rName, instanceType)
}
