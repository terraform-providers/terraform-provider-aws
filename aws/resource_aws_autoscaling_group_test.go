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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
			log.Printf("[WARN] Skipping Auto Scaling Group sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Auto Scaling Groups in Sweeper: %s", err)
	}

	if len(resp.AutoScalingGroups) == 0 {
		log.Print("[DEBUG] No Auto Scaling Groups to sweep")
		return nil
	}

	for _, asg := range resp.AutoScalingGroups {
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

func TestAccAWSAutoScalingGroup_basic(t *testing.T) {
	var group autoscaling.Group
	var lc autoscaling.LaunchConfiguration

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
					testAccCheckAWSAutoScalingGroupHealthyCapacity(&group, 2),
					testAccCheckAWSAutoScalingGroupAttributes(&group, randName),
					testAccMatchResourceAttrRegionalARN("aws_autoscaling_group.bar", "arn", "autoscaling", regexp.MustCompile(`autoScalingGroup:.+`)),
					resource.TestCheckTypeSetElemAttrPair("aws_autoscaling_group.bar", "availability_zones.*", "data.aws_availability_zones.available", "names.0"),
					testAccCheckAutoScalingInstanceRefreshCount(&group, 0),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "default_cooldown", "300"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "desired_capacity", "4"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "enabled_metrics.#", "0"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "force_delete", "true"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "health_check_grace_period", "300"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "health_check_type", "ELB"),
					resource.TestCheckNoResourceAttr("aws_autoscaling_group.bar", "initial_lifecycle_hook.#"),
					resource.TestCheckResourceAttrPair("aws_autoscaling_group.bar", "launch_configuration", "aws_launch_configuration.foobar", "name"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "launch_template.#", "0"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "load_balancers.#", "0"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "max_size", "5"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "metrics_granularity", "1Minute"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "min_size", "2"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "mixed_instances_policy.#", "0"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "name", randName),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "placement_group", ""),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "protect_from_scale_in", "false"),
					testAccCheckResourceAttrGlobalARN("aws_autoscaling_group.bar", "service_linked_role_arn", "iam", "role/aws-service-role/autoscaling.amazonaws.com/AWSServiceRoleForAutoScaling"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "suspended_processes.#", "0"),
					resource.TestCheckNoResourceAttr("aws_autoscaling_group.bar", "tag.#"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "tags.#", "3"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "target_group_arns.#", "0"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "termination_policies.#", "2"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "termination_policies.0", "OldestInstance"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "termination_policies.1", "ClosestToNextInstanceHour"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "vpc_zone_identifier.#", "0"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "max_instance_lifetime", "0"),
					resource.TestCheckNoResourceAttr("aws_autoscaling_group.bar", "instance_refresh.#"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
			{
				Config: testAccAWSAutoScalingGroupConfigUpdate(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckAWSLaunchConfigurationExists("aws_launch_configuration.new", &lc),
					testAccCheckAutoScalingInstanceRefreshCount(&group, 0),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "desired_capacity", "5"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "termination_policies.0", "ClosestToNextInstanceHour"),
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "protect_from_scale_in", "true"),
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
				Config: testAccAWSAutoScalingGroupConfig_namePrefix(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"aws_autoscaling_group.test", "name", nameRegexp),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.test", "arn"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
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
				Config: testAccAWSAutoScalingGroupConfig_autoGeneratedName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"aws_autoscaling_group.bar", "name", asgNameRegexp),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.bar", "arn"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
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
				Config: testAccAWSAutoScalingGroupConfig_terminationPoliciesEmpty(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "termination_policies.#", "0"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
			{
				Config: testAccAWSAutoScalingGroupConfig_terminationPoliciesUpdate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "termination_policies.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "termination_policies.0", "OldestInstance"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
			{
				Config: testAccAWSAutoScalingGroupConfig_terminationPoliciesExplicitDefault(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "termination_policies.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "termination_policies.0", "Default"),
				),
			},
			{
				Config: testAccAWSAutoScalingGroupConfig_terminationPoliciesEmpty(),
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
				ResourceName:      "aws_autoscaling_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
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
				Config: testAccAWSAutoScalingGroupConfigWithAZ(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "availability_zones.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("aws_autoscaling_group.bar", "availability_zones.*", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "vpc_zone_identifier.#", "0"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
			{
				Config: testAccAWSAutoScalingGroupConfigWithVPCIdent(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckAWSAutoScalingGroupAttributesVPCZoneIdentifier(&group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "availability_zones.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("aws_autoscaling_group.bar", "availability_zones.*", "data.aws_availability_zones.available", "names.0"),
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
				Config: testAccAWSAutoScalingGroupConfigWithLoadBalancer(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckAWSAutoScalingGroupAttributesLoadBalancer(&group),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_WithLoadBalancer_ToTargetGroup(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfigWithLoadBalancer(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
			{
				Config: testAccAWSAutoScalingGroupConfigWithTargetGroup(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
			{
				Config: testAccAWSAutoScalingGroupConfigWithLoadBalancer(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
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
					resource.TestCheckResourceAttr("aws_autoscaling_group.bar", "placement_group", randName),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
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
				ResourceName:      "aws_autoscaling_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
			{
				Config: testAccAWSAutoscalingMetricsCollectionConfig_updatingMetricsCollected(),
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
				ResourceName:      "aws_autoscaling_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
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
				Config: testAccAWSAutoscalingMetricsCollectionConfig_allMetricsCollected(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "enabled_metrics.#", "13"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
			{
				Config: testAccAWSAutoscalingMetricsCollectionConfig_updatingMetricsCollected(),
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
				Config: testAccAWSAutoScalingGroupConfig_withServiceLinkedRoleARN(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.bar", "service_linked_role_arn"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_MaxInstanceLifetime(t *testing.T) {
	var group autoscaling.Group

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_withMaxInstanceLifetime(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "max_instance_lifetime", "864000"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
			{
				Config: testAccAWSAutoScalingGroupConfig_withMaxInstanceLifetime_update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "max_instance_lifetime", "604800"),
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
				Config: testAccAWSAutoScalingGroupConfig_ALB_TargetGroup_pre(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckAWSLBTargetGroupExists("aws_lb_target_group.test", &tg),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "target_group_arns.#", "0"),
				),
			},

			{
				Config: testAccAWSAutoScalingGroupConfig_ALB_TargetGroup_post_duo(),
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
				ResourceName:      "aws_autoscaling_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
			{
				Config: testAccAWSAutoScalingGroupConfig_ALB_TargetGroup_post(),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/256
func TestAccAWSAutoScalingGroup_TargetGroupArns(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_TargetGroupArns(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "11"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
			{
				Config: testAccAWSAutoScalingGroupConfig_TargetGroupArns(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "0"),
				),
			},
			{
				Config: testAccAWSAutoScalingGroupConfig_TargetGroupArns(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "11"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_initialLifecycleHook(t *testing.T) {
	var group autoscaling.Group

	randName := fmt.Sprintf("terraform-test-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupWithHookConfig(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					testAccCheckAWSAutoScalingGroupHealthyCapacity(&group, 2),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "initial_lifecycle_hook.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("aws_autoscaling_group.bar", "initial_lifecycle_hook.*", map[string]string{
						"default_result": "CONTINUE",
						"name":           "launching",
					}),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
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
			{
				ResourceName:      "aws_autoscaling_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_InstanceRefresh_Basic(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAutoScalingGroupConfig_InstanceRefresh_Basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"wait_for_capacity_timeout",
					"instance_refresh",
				},
			},
			{
				Config: testAccAwsAutoScalingGroupConfig_InstanceRefresh_Full(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.instance_warmup", "10"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.0.min_healthy_percentage", "50"),
				),
			},
			{
				Config: testAccAwsAutoScalingGroupConfig_InstanceRefresh_Disabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckNoResourceAttr(resourceName, "instance_refresh.#"),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_InstanceRefresh_Start(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	launchConfigurationName := "aws_launch_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAutoScalingGroupConfig_InstanceRefresh_Start("one"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", launchConfigurationName, "name"),
					testAccCheckAutoScalingInstanceRefreshCount(&group, 0),
				),
			},
			{
				Config: testAccAwsAutoScalingGroupConfig_InstanceRefresh_Start("two"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", launchConfigurationName, "name"),
					testAccCheckAutoScalingInstanceRefreshCount(&group, 1),
					testAccCheckAutoScalingInstanceRefreshStatus(&group, 0, autoscaling.InstanceRefreshStatusPending, autoscaling.InstanceRefreshStatusInProgress),
				),
			},
			{
				Config: testAccAwsAutoScalingGroupConfig_InstanceRefresh_Start("three"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttrPair(resourceName, "launch_configuration", launchConfigurationName, "name"),
					testAccCheckAutoScalingInstanceRefreshCount(&group, 2),
					testAccCheckAutoScalingInstanceRefreshStatus(&group, 0, autoscaling.InstanceRefreshStatusPending, autoscaling.InstanceRefreshStatusInProgress),
					testAccCheckAutoScalingInstanceRefreshStatus(&group, 1, autoscaling.InstanceRefreshStatusCancelled),
				),
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_InstanceRefresh_Triggers(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAutoScalingGroupConfig_InstanceRefresh_Basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.strategy", "Rolling"),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.preferences.#", "0"),
				),
			},
			{
				Config: testAccAwsAutoScalingGroupConfig_InstanceRefresh_Triggers(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_refresh.0.triggers.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "instance_refresh.0.triggers.*", "tags"),
					testAccCheckAutoScalingInstanceRefreshCount(&group, 1),
					testAccCheckAutoScalingInstanceRefreshStatus(&group, 0, autoscaling.InstanceRefreshStatusPending, autoscaling.InstanceRefreshStatusInProgress),
				),
			},
		},
	})
}

func testAccCheckAWSAutoScalingGroupExists(n string, group *autoscaling.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Auto Scaling Group ID is set")
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
			return fmt.Errorf("Auto Scaling Group not found")
		}

		*group = *describeGroups.AutoScalingGroups[0]

		return nil
	}
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
				return fmt.Errorf("Auto Scaling Group still exists")
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
		if *group.AutoScalingGroupName != name {
			return fmt.Errorf("Bad Auto Scaling Group name, expected (%s), got (%s)", name, *group.AutoScalingGroupName)
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

func testAccCheckAWSAutoScalingGroupAttributesVPCZoneIdentifier(group *autoscaling.Group) resource.TestCheckFunc {
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

// testAccCheckTags can be used to check the tags on a resource.
func testAccCheckAutoscalingTags(
	ts *[]*autoscaling.TagDescription, key string, expected map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		m := autoscalingTagDescriptionsToMap(ts)
		v, ok := m[key]
		if !ok {
			return fmt.Errorf("Missing tag: %s", key)
		}

		if v["value"] != expected["value"].(string) ||
			v["propagate_at_launch"] != expected["propagate_at_launch"].(bool) {
			return fmt.Errorf("%s: bad value: %s", key, v)
		}

		return nil
	}
}

func testAccCheckAutoscalingTagNotExists(ts *[]*autoscaling.TagDescription, key string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		m := autoscalingTagDescriptionsToMap(ts)
		if _, ok := m[key]; ok {
			return fmt.Errorf("Tag exists when it should not: %s", key)
		}

		return nil
	}
}

// autoscalingTagDescriptionsToMap turns the list of tags into a map.
func autoscalingTagDescriptionsToMap(ts *[]*autoscaling.TagDescription) map[string]map[string]interface{} {
	tags := make(map[string]map[string]interface{})
	for _, t := range *ts {
		tag := map[string]interface{}{
			"key":                 aws.StringValue(t.Key),
			"value":               aws.StringValue(t.Value),
			"propagate_at_launch": aws.BoolValue(t.PropagateAtLaunch),
		}
		tags[aws.StringValue(t.Key)] = tag
	}

	return tags
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_classicVpcZoneIdentifier(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.test", &group),
					resource.TestCheckResourceAttr("aws_autoscaling_group.test", "vpc_zone_identifier.#", "0"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
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
				Config: testAccAWSAutoScalingGroupConfig_withLaunchTemplate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.bar", "launch_template.0.id"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
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
				Config: testAccAWSAutoScalingGroupConfig_withLaunchTemplate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.bar", "launch_template.0.name"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.bar",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
			{
				Config: testAccAWSAutoScalingGroupConfig_withLaunchTemplate_toLaunchConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttrSet(
						"aws_autoscaling_group.bar", "launch_configuration"),
					resource.TestCheckNoResourceAttr(
						"aws_autoscaling_group.bar", "launch_template"),
				),
			},

			{
				Config: testAccAWSAutoScalingGroupConfig_withLaunchTemplate_toLaunchTemplateName(),
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
				Config: testAccAWSAutoScalingGroupConfig_withLaunchTemplate_toLaunchTemplateVersion(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.bar", &group),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "launch_template.0.version", "$Latest"),
				),
			},

			{
				Config: testAccAWSAutoScalingGroupConfig_withLaunchTemplate(),
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
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/256
func TestAccAWSAutoScalingGroup_LoadBalancers(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_LoadBalancers(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "11"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
			{
				Config: testAccAWSAutoScalingGroupConfig_LoadBalancers(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "0"),
				),
			},
			{
				Config: testAccAWSAutoScalingGroupConfig_LoadBalancers(rName, 11),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "11"),
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
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.weighted_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t3.small"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.weighted_capacity", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_MixedInstancesPolicy_CapacityRebalance(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_CapacityRebalance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "capacity_rebalance", "true"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.version", "$Default"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.weighted_capacity", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t3.small"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.weighted_capacity", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
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
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
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
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
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
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandBaseCapacity(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity", "0"),
				),
			},
		},
	})
}

// Test to verify fix for behavior in GH-ISSUE 7368
func TestAccAWSAutoScalingGroup_MixedInstancesPolicy_InstancesDistribution_UpdateToZeroOnDemandBaseCapacity(t *testing.T) {
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
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandBaseCapacity(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.on_demand_base_capacity", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
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
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
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
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
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
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
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
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
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
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_SpotMaxPrice(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.instances_distribution.0.spot_max_price", ""),
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
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
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
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_LaunchTemplate_LaunchTemplateSpecification_Version(rName, "$Latest"),
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
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
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

func TestAccAWSAutoScalingGroup_MixedInstancesPolicy_LaunchTemplate_Override_WeightedCapacity(t *testing.T) {
	var group autoscaling.Group
	resourceName := "aws_autoscaling_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_WeightedCapacity(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists(resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.0.weighted_capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.instance_type", "t3.small"),
					resource.TestCheckResourceAttr(resourceName, "mixed_instances_policy.0.launch_template.0.override.1.weighted_capacity", "4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
		},
	})
}

func TestAccAWSAutoScalingGroup_launchTempPartitionNum(t *testing.T) {
	var group autoscaling.Group

	randName := fmt.Sprintf("terraform-test-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAutoScalingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAutoScalingGroupPartitionConfig(randName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAutoScalingGroupExists("aws_autoscaling_group.test", &group),
				),
			},
			{
				ResourceName:      "aws_autoscaling_group.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_delete",
					"initial_lifecycle_hook",
					"name_prefix",
					"tag",
					"tags",
					"wait_for_capacity_timeout",
					"wait_for_elb_capacity",
				},
			},
		},
	})
}

func testAccAWSAutoScalingGroupConfig_autoGeneratedName() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(), `
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  desired_capacity     = 0
  max_size             = 0
  min_size             = 0
  launch_configuration = aws_launch_configuration.foobar.name
}
`)
}

func testAccAWSAutoScalingGroupConfig_namePrefix() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  desired_capacity     = 0
  max_size             = 0
  min_size             = 0
  name_prefix          = "tf-test-"
  launch_configuration = aws_launch_configuration.test.name
}
`)
}

func testAccAWSAutoScalingGroupConfig_terminationPoliciesEmpty() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  max_size           = 0
  min_size           = 0
  desired_capacity   = 0

  launch_configuration = aws_launch_configuration.foobar.name
}
`)
}

func testAccAWSAutoScalingGroupConfig_terminationPoliciesExplicitDefault() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  desired_capacity     = 0
  termination_policies = ["Default"]

  launch_configuration = aws_launch_configuration.foobar.name
}
`)
}

func testAccAWSAutoScalingGroupConfig_terminationPoliciesUpdate() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  max_size             = 0
  min_size             = 0
  desired_capacity     = 0
  termination_policies = ["OldestInstance"]

  launch_configuration = aws_launch_configuration.foobar.name
}
`)
}

func testAccAWSAutoScalingGroupConfig(name string) string {
	return testAccAvailableAZsNoOptInDefaultExcludeConfig() +
		fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

# TODO: Unused?
resource "aws_placement_group" "test" {
  name     = "asg_pg_%s"
  strategy = "cluster"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = "%s"
  max_size             = 5
  min_size             = 2
  health_check_type    = "ELB"
  desired_capacity     = 4
  force_delete         = true
  termination_policies = ["OldestInstance", "ClosestToNextInstanceHour"]

  launch_configuration = aws_launch_configuration.foobar.name

  tags = [
    {
      key                 = "FromTags1"
      value               = "value1"
      propagate_at_launch = true
    },
    {
      key                 = "FromTags3"
      value               = "value3"
      propagate_at_launch = true
    },
    {
      key                 = "FromTags2"
      value               = "value2"
      propagate_at_launch = true
    },
  ]
}
`, name, name)
}

func testAccAWSAutoScalingGroupConfigUpdate(name string) string {
	return testAccAvailableAZsNoOptInDefaultExcludeConfig() +
		fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_launch_configuration" "new" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones        = [data.aws_availability_zones.available.names[0]]
  name                      = "%s"
  max_size                  = 5
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 5
  force_delete              = true
  termination_policies      = ["ClosestToNextInstanceHour"]
  protect_from_scale_in     = true

  launch_configuration = aws_launch_configuration.new.name

  tags = [
    {
      key                 = "FromTags1Changed"
      value               = "value1changed"
      propagate_at_launch = true
    },
    {
      key                 = "FromTags2"
      value               = "value2changed"
      propagate_at_launch = true
    },
    {
      key                 = "FromTags3"
      value               = "value3"
      propagate_at_launch = true
    },
  ]
}
`, name)
}

func testAccAWSAutoScalingGroupConfigWithLoadBalancer() string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-autoscaling-group-with-lb"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.foo.id
}

resource "aws_subnet" "foo" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.foo.id
  tags = {
    Name = "tf-acc-autoscaling-group-with-load-balancer"
  }
}

resource "aws_security_group" "foo" {
  vpc_id = aws_vpc.foo.id

  ingress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elb" "bar" {
  subnets         = [aws_subnet.foo.id]
  security_groups = [aws_security_group.foo.id]

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 2
    target              = "HTTP:80/"
    interval            = 5
    timeout             = 2
  }

  depends_on = [aws_internet_gateway.gw]
}

resource "aws_launch_configuration" "foobar" {
  image_id        = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type   = "t2.micro"
  security_groups = [aws_security_group.foo.id]

  # Need the instance to listen on port 80 at boot
  user_data = <<EOF
#!/bin/bash
echo "Terraform aws_autoscaling_group Testing" > index.html
nohup python -m SimpleHTTPServer 80 &
EOF
}

resource "aws_autoscaling_group" "bar" {
  vpc_zone_identifier       = [aws_subnet.foo.id]
  max_size                  = 2
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  wait_for_elb_capacity     = 2
  force_delete              = true

  launch_configuration = aws_launch_configuration.foobar.name
  load_balancers       = [aws_elb.bar.name]
}
`)
}

func testAccAWSAutoScalingGroupConfigWithTargetGroup() string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-autoscaling-group-with-lb"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = aws_vpc.foo.id
}

resource "aws_subnet" "foo" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.foo.id
  tags = {
    Name = "tf-acc-autoscaling-group-with-target-group"
  }
}

resource "aws_security_group" "foo" {
  vpc_id = aws_vpc.foo.id

  ingress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_lb_target_group" "foo" {
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.foo.id
}

resource "aws_elb" "bar" {
  subnets         = [aws_subnet.foo.id]
  security_groups = [aws_security_group.foo.id]

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 2
    target              = "HTTP:80/"
    interval            = 5
    timeout             = 2
  }

  depends_on = [aws_internet_gateway.gw]
}

resource "aws_launch_configuration" "foobar" {
  image_id        = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type   = "t2.micro"
  security_groups = [aws_security_group.foo.id]

  # Need the instance to listen on port 80 at boot
  user_data = <<EOF
#!/bin/bash
echo "Terraform aws_autoscaling_group Testing" > index.html
nohup python -m SimpleHTTPServer 80 &
EOF
}

resource "aws_autoscaling_group" "bar" {
  vpc_zone_identifier       = [aws_subnet.foo.id]
  max_size                  = 2
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  wait_for_elb_capacity     = 2
  force_delete              = true

  launch_configuration = aws_launch_configuration.foobar.name
  target_group_arns    = [aws_lb_target_group.foo.arn]
}
`)
}

func testAccAWSAutoScalingGroupConfigWithAZ() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
resource "aws_vpc" "default" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-autoscaling-group-with-az"
  }
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = "tf-acc-autoscaling-group-with-az"
  }
}

data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = [
    data.aws_availability_zones.available.names[0]
  ]
  desired_capacity     = 0
  max_size             = 0
  min_size             = 0
  launch_configuration = aws_launch_configuration.foobar.name
}
`)
}

func testAccAWSAutoScalingGroupConfigWithVPCIdent() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
resource "aws_vpc" "default" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-autoscaling-group-with-vpc-id"
  }
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
  tags = {
    Name = "tf-acc-autoscaling-group-with-vpc-id"
  }
}

data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  vpc_zone_identifier = [
    aws_subnet.main.id,
  ]
  desired_capacity     = 0
  max_size             = 0
  min_size             = 0
  launch_configuration = aws_launch_configuration.foobar.name
}
`)
}

func testAccAWSAutoScalingGroupConfig_withPlacementGroup(name string) string {
	return testAccAvailableAZsNoOptInDefaultExcludeConfig() +
		fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "c3.large"
}

resource "aws_placement_group" "test" {
  name     = "%s"
  strategy = "cluster"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones        = [data.aws_availability_zones.available.names[0]]
  name                      = "%s"
  max_size                  = 1
  min_size                  = 1
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 1
  force_delete              = true
  termination_policies      = ["OldestInstance", "ClosestToNextInstanceHour"]
  placement_group           = aws_placement_group.test.name

  launch_configuration = aws_launch_configuration.foobar.name

  tag {
    key                 = "Foo"
    value               = "foo-bar"
    propagate_at_launch = true
  }
}
`, name, name)
}

func testAccAWSAutoScalingGroupConfig_withServiceLinkedRoleARN() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_iam_role" "autoscaling_service_linked_role" {
  name = "AWSServiceRoleForAutoScaling"
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones      = [data.aws_availability_zones.available.names[0]]
  desired_capacity        = 0
  max_size                = 0
  min_size                = 0
  launch_configuration    = aws_launch_configuration.foobar.name
  service_linked_role_arn = data.aws_iam_role.autoscaling_service_linked_role.arn
}
`)
}

func testAccAWSAutoScalingGroupConfig_withMaxInstanceLifetime() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones    = [data.aws_availability_zones.available.names[0]]
  desired_capacity      = 0
  max_size              = 0
  min_size              = 0
  launch_configuration  = aws_launch_configuration.foobar.name
  max_instance_lifetime = "864000"
}
`)
}

func testAccAWSAutoScalingGroupConfig_withMaxInstanceLifetime_update() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones    = [data.aws_availability_zones.available.names[0]]
  desired_capacity      = 0
  max_size              = 0
  min_size              = 0
  launch_configuration  = aws_launch_configuration.foobar.name
  max_instance_lifetime = "604800"
}
`)
}

func testAccAWSAutoscalingMetricsCollectionConfig_allMetricsCollected() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones        = [data.aws_availability_zones.available.names[0]]
  max_size                  = 1
  min_size                  = 0
  health_check_grace_period = 300
  health_check_type         = "EC2"
  desired_capacity          = 0
  force_delete              = true
  termination_policies      = ["OldestInstance", "ClosestToNextInstanceHour"]
  launch_configuration      = aws_launch_configuration.foobar.name
  enabled_metrics = ["GroupTotalInstances",
    "GroupPendingInstances",
    "GroupTerminatingInstances",
    "GroupDesiredCapacity",
    "GroupInServiceInstances",
    "GroupMinSize",
    "GroupMaxSize",
    "GroupPendingCapacity",
    "GroupInServiceCapacity",
    "GroupStandbyCapacity",
    "GroupTotalCapacity",
    "GroupTerminatingCapacity",
    "GroupStandbyInstances"
  ]
  metrics_granularity = "1Minute"
}
`)
}

func testAccAWSAutoscalingMetricsCollectionConfig_updatingMetricsCollected() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones        = [data.aws_availability_zones.available.names[0]]
  max_size                  = 1
  min_size                  = 0
  health_check_grace_period = 300
  health_check_type         = "EC2"
  desired_capacity          = 0
  force_delete              = true
  termination_policies      = ["OldestInstance", "ClosestToNextInstanceHour"]
  launch_configuration      = aws_launch_configuration.foobar.name
  enabled_metrics = ["GroupTotalInstances",
    "GroupPendingInstances",
    "GroupTerminatingInstances",
    "GroupDesiredCapacity",
    "GroupMaxSize"
  ]
  metrics_granularity = "1Minute"
}
`)
}

func testAccAWSAutoScalingGroupConfig_ALB_TargetGroup_pre() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
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
  vpc_id   = aws_vpc.default.id
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-autoscaling-group-alb-target-group-main"
  }
}

resource "aws_subnet" "alt" {
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-autoscaling-group-alb-target-group-alt"
  }
}

data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id          = data.aws_ami.test_ami.id
  instance_type     = "t2.micro"
  enable_monitoring = false
}

resource "aws_autoscaling_group" "bar" {
  vpc_zone_identifier = [
    aws_subnet.main.id,
    aws_subnet.alt.id,
  ]

  max_size                  = 2
  min_size                  = 0
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 0
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.foobar.name

}

resource "aws_security_group" "tf_test_self" {
  name        = "tf_test_alb_asg"
  description = "tf_test_alb_asg"
  vpc_id      = aws_vpc.default.id

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
`)
}

func testAccAWSAutoScalingGroupConfig_ALB_TargetGroup_post() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
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
  vpc_id   = aws_vpc.default.id
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-autoscaling-group-alb-target-group-main"
  }
}

resource "aws_subnet" "alt" {
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-autoscaling-group-alb-target-group-alt"
  }
}

data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id          = data.aws_ami.test_ami.id
  instance_type     = "t2.micro"
  enable_monitoring = false
}

resource "aws_autoscaling_group" "bar" {
  vpc_zone_identifier = [
    aws_subnet.main.id,
    aws_subnet.alt.id,
  ]

  target_group_arns = [aws_lb_target_group.test.arn]

  max_size                  = 2
  min_size                  = 0
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 0
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.foobar.name

}

resource "aws_security_group" "tf_test_self" {
  name        = "tf_test_alb_asg"
  description = "tf_test_alb_asg"
  vpc_id      = aws_vpc.default.id

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
`)
}

func testAccAWSAutoScalingGroupConfig_ALB_TargetGroup_post_duo() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
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
  vpc_id   = aws_vpc.default.id
}

resource "aws_lb_target_group" "test_more" {
  name     = "tf-example-alb-tg-more"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.default.id
}

resource "aws_subnet" "main" {
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-autoscaling-group-alb-target-group-main"
  }
}

resource "aws_subnet" "alt" {
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-autoscaling-group-alb-target-group-alt"
  }
}

data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id          = data.aws_ami.test_ami.id
  instance_type     = "t2.micro"
  enable_monitoring = false
}

resource "aws_autoscaling_group" "bar" {
  vpc_zone_identifier = [
    aws_subnet.main.id,
    aws_subnet.alt.id,
  ]

  target_group_arns = [
    aws_lb_target_group.test.arn,
    aws_lb_target_group.test_more.arn,
  ]

  max_size                  = 2
  min_size                  = 0
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 0
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.foobar.name

}

resource "aws_security_group" "tf_test_self" {
  name        = "tf_test_alb_asg"
  description = "tf_test_alb_asg"
  vpc_id      = aws_vpc.default.id

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
`)
}

func testAccAWSAutoScalingGroupConfig_TargetGroupArns(rName string, tgCount int) string {
	return testAccAvailableAZsNoOptInDefaultExcludeConfig() +
		fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.micro"
  name          = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb_target_group" "test" {
  count = %[2]d

  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_autoscaling_group" "test" {
  force_delete        = true
  max_size            = 0
  min_size            = 0
  target_group_arns   = length(aws_lb_target_group.test) > 0 ? aws_lb_target_group.test[*].arn : []
  vpc_zone_identifier = [aws_subnet.test.id]

  launch_template {
    id = aws_launch_template.test.id
  }
}
`, rName, tgCount)
}

func testAccAWSAutoScalingGroupWithHookConfig(name string) string {
	return testAccAvailableAZsNoOptInDefaultExcludeConfig() +
		fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = "%s"
  max_size             = 5
  min_size             = 2
  health_check_type    = "ELB"
  desired_capacity     = 4
  force_delete         = true
  termination_policies = ["OldestInstance", "ClosestToNextInstanceHour"]

  launch_configuration = aws_launch_configuration.foobar.name

  initial_lifecycle_hook {
    name                 = "launching"
    default_result       = "CONTINUE"
    heartbeat_timeout    = 30 # minimum value
    lifecycle_transition = "autoscaling:EC2_INSTANCE_LAUNCHING"
  }
}
`, name)
}

func testAccAWSAutoScalingGroupConfig_ALB_TargetGroup_ELBCapacity(rInt int) string {
	return testAccAvailableAZsNoOptInDefaultExcludeConfig() +
		fmt.Sprintf(`
resource "aws_vpc" "default" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = "true"
  enable_dns_support   = "true"

  tags = {
    Name = "terraform-testacc-autoscaling-group-alb-target-group-elb-capacity"
  }
}

resource "aws_lb" "test_lb" {
  subnets = [aws_subnet.main.id, aws_subnet.alt.id]

  tags = {
    Name = "testAccAWSAutoScalingGroupConfig_ALB_TargetGroup_ELBCapacity"
  }
}

resource "aws_lb_listener" "test_listener" {
  load_balancer_arn = aws_lb.test_lb.arn
  port              = "80"

  default_action {
    target_group_arn = aws_lb_target_group.test.arn
    type             = "forward"
  }
}

resource "aws_lb_target_group" "test" {
  name     = "tf-alb-test-%d"
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.default.id

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
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-autoscaling-group-alb-target-group-elb-capacity-main"
  }
}

resource "aws_subnet" "alt" {
  vpc_id            = aws_vpc.default.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-autoscaling-group-alb-target-group-elb-capacity-alt"
  }
}

resource "aws_internet_gateway" "internet_gateway" {
  vpc_id = aws_vpc.default.id
}

resource "aws_route_table" "route_table" {
  vpc_id = aws_vpc.default.id
}

resource "aws_route_table_association" "route_table_association_main" {
  subnet_id      = aws_subnet.main.id
  route_table_id = aws_route_table.route_table.id
}

resource "aws_route_table_association" "route_table_association_alt" {
  subnet_id      = aws_subnet.alt.id
  route_table_id = aws_route_table.route_table.id
}

resource "aws_route" "public_default_route" {
  route_table_id         = aws_route_table.route_table.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.internet_gateway.id
}

data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id                    = data.aws_ami.test_ami.id
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
    aws_subnet.main.id,
    aws_subnet.alt.id,
  ]

  target_group_arns = [aws_lb_target_group.test.arn]

  max_size                  = 2
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  desired_capacity          = 2
  wait_for_elb_capacity     = 2
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.foobar.name
}
`, rInt)
}

func testAccAWSAutoScalingGroupConfigWithSuspendedProcesses(name string) string {
	return testAccAvailableAZsNoOptInDefaultExcludeConfig() +
		fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_placement_group" "test" {
  name     = "asg_pg_%s"
  strategy = "cluster"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = "%s"
  max_size             = 5
  min_size             = 2
  health_check_type    = "ELB"
  desired_capacity     = 4
  force_delete         = true
  termination_policies = ["OldestInstance", "ClosestToNextInstanceHour"]

  launch_configuration = aws_launch_configuration.foobar.name

  suspended_processes = ["AlarmNotification", "ScheduledActions"]

  tag {
    key                 = "Foo"
    value               = "foo-bar"
    propagate_at_launch = true
  }
}
`, name, name)
}

func testAccAWSAutoScalingGroupConfigWithSuspendedProcessesUpdated(name string) string {
	return testAccAvailableAZsNoOptInDefaultExcludeConfig() +
		fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_placement_group" "test" {
  name     = "asg_pg_%s"
  strategy = "cluster"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  name                 = "%s"
  max_size             = 5
  min_size             = 2
  health_check_type    = "ELB"
  desired_capacity     = 4
  force_delete         = true
  termination_policies = ["OldestInstance", "ClosestToNextInstanceHour"]

  launch_configuration = aws_launch_configuration.foobar.name

  suspended_processes = ["AZRebalance", "ScheduledActions"]

  tag {
    key                 = "Foo"
    value               = "foo-bar"
    propagate_at_launch = true
  }
}
`, name, name)
}

func testAccAWSAutoScalingGroupConfig_classicVpcZoneIdentifier() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
resource "aws_autoscaling_group" "test" {
  min_size = 0
  max_size = 0

  availability_zones   = [data.aws_availability_zones.available.names[0]]
  launch_configuration = aws_launch_configuration.test.name
}

data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t1.micro"
}
`)
}

func testAccAWSAutoScalingGroupConfig_withLaunchTemplate() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "foobar" {
  name_prefix   = "foobar"
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  launch_template {
    id      = aws_launch_template.foobar.id
    version = aws_launch_template.foobar.default_version
  }
}
`)
}

func testAccAWSAutoScalingGroupConfig_withLaunchTemplate_toLaunchConfig() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "foobar" {
  name_prefix   = "foobar"
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones   = [data.aws_availability_zones.available.names[0]]
  desired_capacity     = 0
  max_size             = 0
  min_size             = 0
  launch_configuration = aws_launch_configuration.test.name
}
`)
}

func testAccAWSAutoScalingGroupConfig_withLaunchTemplate_toLaunchTemplateName() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "foobar" {
  name_prefix   = "foobar"
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_launch_template" "foobar2" {
  name          = "foobar2"
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  launch_template {
    name = "foobar2"
  }
}
`)
}

func testAccAWSAutoScalingGroupConfig_withLaunchTemplate_toLaunchTemplateVersion() string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "foobar" {
  name_prefix   = "foobar"
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_launch_template" "foobar2" {
  name          = "foobar2"
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  launch_template {
    id      = aws_launch_template.foobar.id
    version = "$Latest"
  }
}
`)
}

func testAccAWSAutoScalingGroupConfig_LaunchTemplate_IAMInstanceProfile(rName string) string {
	return testAccAvailableAZsNoOptInDefaultExcludeConfig() +
		fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_iam_role" "test" {
  assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"ec2.amazonaws.com\"]},\"Action\":[\"sts:AssumeRole\"]}]}"
  name               = %q
}

resource "aws_iam_instance_profile" "test" {
  name = %q
  role = aws_iam_role.test.name
}

resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t2.micro"
  name          = %q

  iam_instance_profile {
    name = aws_iam_instance_profile.test.id
  }
}

resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  launch_template {
    id = aws_launch_template.test.id
  }
}
`, rName, rName, rName, rName)
}

func testAccAWSAutoScalingGroupConfig_LoadBalancers(rName string, elbCount int) string {
	return testAccAvailableAZsNoOptInDefaultExcludeConfig() +
		fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.micro"
  name          = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id
}

resource "aws_elb" "test" {
  count      = %[2]d
  depends_on = [aws_internet_gateway.test]

  subnets = [aws_subnet.test.id]

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }
}

resource "aws_autoscaling_group" "test" {
  force_delete        = true
  max_size            = 0
  min_size            = 0
  load_balancers      = length(aws_elb.test) > 0 ? aws_elb.test[*].name : []
  vpc_zone_identifier = [aws_subnet.test.id]

  launch_template {
    id = aws_launch_template.test.id
  }
}
`, rName, elbCount)
}

func testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName string) string {
	return testAccAvailableAZsNoOptInDefaultExcludeConfig() +
		fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.micro"
  name          = %q
}
`, rName)
}

func testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy(rName string) string {
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type     = "t2.micro"
        weighted_capacity = "1"
      }
      override {
        instance_type     = "t3.small"
        weighted_capacity = "2"
      }
    }
  }
}
`, rName)
}

func testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_CapacityRebalance(rName string) string {
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  capacity_rebalance = true
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type     = "t2.micro"
        weighted_capacity = "1"
      }
      override {
        instance_type     = "t3.small"
        weighted_capacity = "2"
      }
    }
  }
}
`, rName)
}

func testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_InstancesDistribution_OnDemandAllocationStrategy(rName, onDemandAllocationStrategy string) string {
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
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
        launch_template_id = aws_launch_template.test.id
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
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
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
        launch_template_id = aws_launch_template.test.id
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
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
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
        launch_template_id = aws_launch_template.test.id
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
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
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
        launch_template_id = aws_launch_template.test.id
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
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
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
        launch_template_id = aws_launch_template.test.id
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
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
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
        launch_template_id = aws_launch_template.test.id
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
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_name = aws_launch_template.test.name
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
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
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
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
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

func testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_LaunchTemplate_Override_WeightedCapacity(rName string) string {
	return testAccAWSAutoScalingGroupConfig_MixedInstancesPolicy_Base(rName) +
		fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 4
  max_size           = 6
  min_size           = 2
  name               = %q

  mixed_instances_policy {
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test.id
      }

      override {
        instance_type     = "t2.micro"
        weighted_capacity = "2"
      }
      override {
        instance_type     = "t3.small"
        weighted_capacity = "4"
      }
    }
  }
}
`, rName)
}

func testAccAWSAutoScalingGroupPartitionConfig(rName string) string {
	return testAccAvailableAZsNoOptInDefaultExcludeConfig() +
		fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "this" {
  name          = %[1]q
  image_id      = data.aws_ami.test_ami.id
  instance_type = "m5.large"

  placement {
    tenancy    = "default"
    group_name = aws_placement_group.test.id
  }
}

resource "aws_placement_group" "test" {
  name     = %[1]q
  strategy = "cluster"
}

resource "aws_autoscaling_group" "test" {
  name_prefix        = "test"
  availability_zones = [data.aws_availability_zones.available.names[0]]
  min_size           = 0
  max_size           = 0

  launch_template {
    id      = aws_launch_template.this.id
    version = "$Latest"
  }
}
`, rName)
}

func testAccAwsAutoScalingGroupConfig_InstanceRefresh_Basic() string {
	return `
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.current.names[0]]
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"
  }
}

data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_availability_zones" "current" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.nano"
}
`
}

func testAccAwsAutoScalingGroupConfig_InstanceRefresh_Full() string {
	return `
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.current.names[0]]
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"
    preferences {
      instance_warmup        = 10
      min_healthy_percentage = 50
    }
  }
}

data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_availability_zones" "current" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.nano"
}
`
}

func testAccAwsAutoScalingGroupConfig_InstanceRefresh_Disabled() string {
	return `
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.current.names[0]]
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name
}

data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_availability_zones" "current" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.nano"
}
`
}

func testAccAwsAutoScalingGroupConfig_InstanceRefresh_Start(launchConfigurationName string) string {
	return fmt.Sprintf(`
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.current.names[0]]
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"
  }
}

data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_availability_zones" "current" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_launch_configuration" "test" {
  name_prefix   = %[1]q
  image_id      = data.aws_ami.test.id
  instance_type = "t3.nano"

  lifecycle {
    create_before_destroy = true
  }
}
`, launchConfigurationName)
}

func testAccAwsAutoScalingGroupConfig_InstanceRefresh_Triggers() string {
	return `
resource "aws_autoscaling_group" "test" {
  availability_zones   = [data.aws_availability_zones.current.names[0]]
  max_size             = 2
  min_size             = 1
  desired_capacity     = 1
  launch_configuration = aws_launch_configuration.test.name

  instance_refresh {
    strategy = "Rolling"
    triggers = ["tags"]
  }

  tag {
    key                 = "Key"
    value               = "Value"
    propagate_at_launch = true
  }
}

data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_availability_zones" "current" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.nano"
}
`
}

func testAccCheckAutoScalingInstanceRefreshCount(group *autoscaling.Group, expected int) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).autoscalingconn

		input := autoscaling.DescribeInstanceRefreshesInput{
			AutoScalingGroupName: group.AutoScalingGroupName,
		}
		resp, err := conn.DescribeInstanceRefreshes(&input)
		if err != nil {
			return fmt.Errorf("error describing Auto Scaling Group (%s) Instance Refreshes: %w", aws.StringValue(group.AutoScalingGroupName), err)
		}

		if len(resp.InstanceRefreshes) != expected {
			return fmt.Errorf("expected %d Instance Refreshes, got %d", expected, len(resp.InstanceRefreshes))
		}
		return nil
	}
}

func testAccCheckAutoScalingInstanceRefreshStatus(group *autoscaling.Group, offset int, expected ...string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).autoscalingconn

		input := autoscaling.DescribeInstanceRefreshesInput{
			AutoScalingGroupName: group.AutoScalingGroupName,
		}
		resp, err := conn.DescribeInstanceRefreshes(&input)
		if err != nil {
			return fmt.Errorf("error describing Auto Scaling Group (%s) Instance Refreshes: %w", aws.StringValue(group.AutoScalingGroupName), err)
		}

		if len(resp.InstanceRefreshes) < offset {
			return fmt.Errorf("expected at least %d Instance Refreshes, got %d", offset+1, len(resp.InstanceRefreshes))
		}

		actual := aws.StringValue(resp.InstanceRefreshes[offset].Status)
		for _, s := range expected {
			if actual == s {
				return nil
			}
		}
		return fmt.Errorf("expected Instance Refresh at index %d to be in %q, got %q", offset, expected, actual)
	}
}

func TestCreateAutoScalingGroupInstanceRefreshInput(t *testing.T) {
	const asgName = "test-asg"
	testCases := []struct {
		name     string
		input    []interface{}
		expected *autoscaling.StartInstanceRefreshInput
	}{
		{
			name:     "empty list",
			input:    []interface{}{},
			expected: nil,
		},
		{
			name:     "nil",
			input:    []interface{}{nil},
			expected: nil,
		},
		{
			name: "defaults",
			input: []interface{}{map[string]interface{}{
				"strategy":    "Rolling",
				"preferences": []interface{}{},
			}},
			expected: &autoscaling.StartInstanceRefreshInput{
				AutoScalingGroupName: aws.String(asgName),
				Strategy:             aws.String("Rolling"),
				Preferences:          nil,
			},
		},
		{
			name: "instance_warmup only",
			input: []interface{}{map[string]interface{}{
				"strategy": "Rolling",
				"preferences": []interface{}{
					map[string]interface{}{
						"instance_warmup": "60",
					},
				},
			}},
			expected: &autoscaling.StartInstanceRefreshInput{
				AutoScalingGroupName: aws.String(asgName),
				Strategy:             aws.String("Rolling"),
				Preferences: &autoscaling.RefreshPreferences{
					InstanceWarmup:       aws.Int64(60),
					MinHealthyPercentage: nil,
				},
			},
		},
		{
			name: "instance_warmup zero",
			input: []interface{}{map[string]interface{}{
				"strategy": "Rolling",
				"preferences": []interface{}{
					map[string]interface{}{
						"instance_warmup": "0",
					},
				},
			}},
			expected: &autoscaling.StartInstanceRefreshInput{
				AutoScalingGroupName: aws.String(asgName),
				Strategy:             aws.String("Rolling"),
				Preferences: &autoscaling.RefreshPreferences{
					InstanceWarmup:       aws.Int64(0),
					MinHealthyPercentage: nil,
				},
			},
		},
		{
			name: "instance_warmup empty string",
			input: []interface{}{map[string]interface{}{
				"strategy": "Rolling",
				"preferences": []interface{}{
					map[string]interface{}{
						"instance_warmup":        "",
						"min_healthy_percentage": 80,
					},
				},
			}},
			expected: &autoscaling.StartInstanceRefreshInput{
				AutoScalingGroupName: aws.String(asgName),
				Strategy:             aws.String("Rolling"),
				Preferences: &autoscaling.RefreshPreferences{
					InstanceWarmup:       nil,
					MinHealthyPercentage: aws.Int64(80),
				},
			},
		},
		{
			name: "min_healthy_percentage only",
			input: []interface{}{map[string]interface{}{
				"strategy": "Rolling",
				"preferences": []interface{}{
					map[string]interface{}{
						"min_healthy_percentage": 80,
					},
				},
			}},
			expected: &autoscaling.StartInstanceRefreshInput{
				AutoScalingGroupName: aws.String(asgName),
				Strategy:             aws.String("Rolling"),
				Preferences: &autoscaling.RefreshPreferences{
					InstanceWarmup:       nil,
					MinHealthyPercentage: aws.Int64(80),
				},
			},
		},
		{
			name: "preferences",
			input: []interface{}{map[string]interface{}{
				"strategy": "Rolling",
				"preferences": []interface{}{
					map[string]interface{}{
						"instance_warmup":        "60",
						"min_healthy_percentage": 80,
					},
				},
			}},
			expected: &autoscaling.StartInstanceRefreshInput{
				AutoScalingGroupName: aws.String(asgName),
				Strategy:             aws.String("Rolling"),
				Preferences: &autoscaling.RefreshPreferences{
					InstanceWarmup:       aws.Int64(60),
					MinHealthyPercentage: aws.Int64(80),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := createAutoScalingGroupInstanceRefreshInput(asgName, testCase.input)

			if !reflect.DeepEqual(got, testCase.expected) {
				t.Errorf("got %s, expected %s", got, testCase.expected)
			}
		})
	}
}
