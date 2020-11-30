package aws

import (
	"fmt"
	"log"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscalingplans"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/autoscalingplans/finder"
)

func init() {
	resource.AddTestSweepers("aws_autoscalingplans_scaling_plan", &resource.Sweeper{
		Name: "aws_autoscalingplans_scaling_plan",
		F:    testSweepAutoScalingPlansScalingPlans,
	})
}

func testSweepAutoScalingPlansScalingPlans(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).autoscalingplansconn
	input := &autoscalingplans.DescribeScalingPlansInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.DescribeScalingPlans(input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Auto Scaling Scaling Plans sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Auto Scaling Scaling Plans: %w", err))
			return sweeperErrs.ErrorOrNil()
		}

		for _, scalingPlan := range output.ScalingPlans {
			scalingPlanName := aws.StringValue(scalingPlan.ScalingPlanName)
			scalingPlanVersion := int(aws.Int64Value(scalingPlan.ScalingPlanVersion))

			r := resourceAwsAutoScalingPlansScalingPlan()
			d := r.Data(nil)
			d.SetId("????????????????") // ID not used in Delete.
			d.Set("name", scalingPlanName)
			d.Set("scaling_plan_version", scalingPlanVersion)
			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsAutoScalingPlansScalingPlan_basicDynamicScaling(t *testing.T) {
	var scalingPlan autoscalingplans.ScalingPlan
	resourceName := "aws_autoscalingplans_scaling_plan.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAutoScalingPlansScalingPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingPlansScalingPlanConfigBasicDynamicScaling(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingPlansScalingPlanExists(resourceName, &scalingPlan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "scaling_plan_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.0.cloudformation_stack_arn", ""),
					testAccCheckAutoScalingPlansApplicationSourceTags(&scalingPlan, map[string][]string{
						rName: {rName},
					}),
					resource.TestCheckResourceAttr(resourceName, "scaling_instruction.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scaling_instruction.*", map[string]string{
						"customized_load_metric_specification.#": "0",
						"disable_dynamic_scaling":                "false",
						"max_capacity":                           "3",
						"min_capacity":                           "0",
						"predefined_load_metric_specification.#": "0",
						"resource_id":                            fmt.Sprintf("autoScalingGroup/%s", rName),
						"scalable_dimension":                     "autoscaling:autoScalingGroup:DesiredCapacity",
						"scaling_policy_update_behavior":         "KeepExternalPolicies",
						"service_namespace":                      "autoscaling",
						"target_tracking_configuration.#":        "1",
					}),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return rName, nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsAutoScalingPlansScalingPlan_basicPredictiveScaling(t *testing.T) {
	var scalingPlan autoscalingplans.ScalingPlan
	resourceName := "aws_autoscalingplans_scaling_plan.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckIamServiceLinkedRole(t, "/aws-service-role/autoscaling-plans")
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAutoScalingPlansScalingPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingPlansScalingPlanConfigBasicPredictiveScaling(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingPlansScalingPlanExists(resourceName, &scalingPlan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "scaling_plan_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.0.cloudformation_stack_arn", ""),
					testAccCheckAutoScalingPlansApplicationSourceTags(&scalingPlan, map[string][]string{
						rName: {rName},
					}),
					resource.TestCheckResourceAttr(resourceName, "scaling_instruction.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scaling_instruction.*", map[string]string{
						"customized_load_metric_specification.#": "0",
						"disable_dynamic_scaling":                "true",
						"max_capacity":                           "3",
						"min_capacity":                           "0",
						"predefined_load_metric_specification.#": "1",
						"predefined_load_metric_specification.0.predefined_load_metric_type": "ASGTotalCPUUtilization",
						"predictive_scaling_max_capacity_behavior":                           "SetForecastCapacityToMaxCapacity",
						"predictive_scaling_mode":                                            "ForecastOnly",
						"resource_id":                                                        fmt.Sprintf("autoScalingGroup/%s", rName),
						"scalable_dimension":                                                 "autoscaling:autoScalingGroup:DesiredCapacity",
						"scaling_policy_update_behavior":                                     "KeepExternalPolicies",
						"service_namespace":                                                  "autoscaling",
						"target_tracking_configuration.#":                                    "1",
					}),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return rName, nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsAutoScalingPlansScalingPlan_basicUpdate(t *testing.T) {
	var scalingPlan autoscalingplans.ScalingPlan
	resourceName := "aws_autoscalingplans_scaling_plan.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckIamServiceLinkedRole(t, "/aws-service-role/autoscaling-plans")
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAutoScalingPlansScalingPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingPlansScalingPlanConfigBasicDynamicScaling(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingPlansScalingPlanExists(resourceName, &scalingPlan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "scaling_plan_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.0.cloudformation_stack_arn", ""),
					testAccCheckAutoScalingPlansApplicationSourceTags(&scalingPlan, map[string][]string{
						rName: {rName},
					}),
					resource.TestCheckResourceAttr(resourceName, "scaling_instruction.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scaling_instruction.*", map[string]string{
						"customized_load_metric_specification.#": "0",
						"disable_dynamic_scaling":                "false",
						"max_capacity":                           "3",
						"min_capacity":                           "0",
						"predefined_load_metric_specification.#": "0",
						"resource_id":                            fmt.Sprintf("autoScalingGroup/%s", rName),
						"scalable_dimension":                     "autoscaling:autoScalingGroup:DesiredCapacity",
						"scaling_policy_update_behavior":         "KeepExternalPolicies",
						"service_namespace":                      "autoscaling",
						"target_tracking_configuration.#":        "1",
					}),
				),
			},
			{
				Config: testAccAutoScalingPlansScalingPlanConfigBasicPredictiveScaling(rName, rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingPlansScalingPlanExists(resourceName, &scalingPlan),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "scaling_plan_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_source.0.cloudformation_stack_arn", ""),
					testAccCheckAutoScalingPlansApplicationSourceTags(&scalingPlan, map[string][]string{
						rNameUpdated: {rNameUpdated},
					}),
					resource.TestCheckResourceAttr(resourceName, "scaling_instruction.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "scaling_instruction.*", map[string]string{
						"customized_load_metric_specification.#": "0",
						"disable_dynamic_scaling":                "true",
						"max_capacity":                           "3",
						"min_capacity":                           "0",
						"predefined_load_metric_specification.#": "1",
						"predefined_load_metric_specification.0.predefined_load_metric_type": "ASGTotalCPUUtilization",
						"predictive_scaling_max_capacity_behavior":                           "SetForecastCapacityToMaxCapacity",
						"predictive_scaling_mode":                                            "ForecastOnly",
						"resource_id":                                                        fmt.Sprintf("autoScalingGroup/%s", rName),
						"scalable_dimension":                                                 "autoscaling:autoScalingGroup:DesiredCapacity",
						"scaling_policy_update_behavior":                                     "KeepExternalPolicies",
						"service_namespace":                                                  "autoscaling",
						"target_tracking_configuration.#":                                    "1",
					}),
				),
			},
			{
				ResourceName: resourceName,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return rName, nil
				},
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsAutoScalingPlansScalingPlan_disappears(t *testing.T) {
	var scalingPlan autoscalingplans.ScalingPlan
	resourceName := "aws_autoscalingplans_scaling_plan.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAutoScalingPlansScalingPlanDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAutoScalingPlansScalingPlanConfigBasicDynamicScaling(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAutoScalingPlansScalingPlanExists(resourceName, &scalingPlan),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAutoScalingPlansScalingPlan(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAutoScalingPlansScalingPlanDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).autoscalingplansconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_autoscalingplans_scaling_plan" {
			continue
		}

		scalingPlanVersion, err := strconv.Atoi(rs.Primary.Attributes["scaling_plan_version"])
		if err != nil {
			return err
		}

		scalingPlan, err := finder.ScalingPlan(conn, rs.Primary.Attributes["name"], scalingPlanVersion)
		if err != nil {
			return err
		}
		if scalingPlan == nil {
			continue
		}
		return fmt.Errorf("Auto Scaling Scaling Plan %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAutoScalingPlansScalingPlanExists(name string, v *autoscalingplans.ScalingPlan) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).autoscalingplansconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No Auto Scaling Scaling Plan ID is set")
		}

		scalingPlanVersion, err := strconv.Atoi(rs.Primary.Attributes["scaling_plan_version"])
		if err != nil {
			return err
		}

		scalingPlan, err := finder.ScalingPlan(conn, rs.Primary.Attributes["name"], scalingPlanVersion)
		if err != nil {
			return err
		}
		if scalingPlan == nil {
			return fmt.Errorf("Auto Scaling Scaling Plan %s not found", rs.Primary.ID)
		}

		*v = *scalingPlan

		return nil
	}
}

func testAccCheckAutoScalingPlansApplicationSourceTags(scalingPlan *autoscalingplans.ScalingPlan, expectedTagFilters map[string][]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, tagFilter := range scalingPlan.ApplicationSource.TagFilters {
			key := aws.StringValue(tagFilter.Key)
			values := aws.StringValueSlice(tagFilter.Values)

			expectedValues, ok := expectedTagFilters[key]
			if !ok {
				return fmt.Errorf("Scaling plan application source tag filter key %q not expected", key)
			}

			sort.Strings(values)
			sort.Strings(expectedValues)
			if !reflect.DeepEqual(values, expectedValues) {
				return fmt.Errorf("Scaling plan application source tag filter values %q, expected %q", values, expectedValues)
			}
		}

		return nil
	}
}

func testAccAutoScalingPlansScalingPlanConfigBase(rName, tagName string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
}

resource "aws_autoscaling_group" "test" {
  name = %[1]q

  launch_configuration = aws_launch_configuration.test.name
  availability_zones   = [data.aws_availability_zones.available.names[0]]

  min_size         = 0
  max_size         = 3
  desired_capacity = 0

  tags = [
    {
      key                 = %[2]q
      value               = %[2]q
      propagate_at_launch = true
    },
  ]
}
`, rName, tagName))
}

func testAccAutoScalingPlansScalingPlanConfigBasicDynamicScaling(rName, tagName string) string {
	return composeConfig(
		testAccAutoScalingPlansScalingPlanConfigBase(rName, tagName),
		fmt.Sprintf(`
resource "aws_autoscalingplans_scaling_plan" "test" {
  name = %[1]q

  application_source {
    tag_filter {
      key    = %[2]q
      values = [%[2]q]
    }
  }

  scaling_instruction {
    max_capacity       = aws_autoscaling_group.test.max_size
    min_capacity       = aws_autoscaling_group.test.min_size
    resource_id        = format("autoScalingGroup/%%s", aws_autoscaling_group.test.name)
    scalable_dimension = "autoscaling:autoScalingGroup:DesiredCapacity"
    service_namespace  = "autoscaling"

    target_tracking_configuration {
      predefined_scaling_metric_specification {
        predefined_scaling_metric_type = "ASGAverageCPUUtilization"
      }

      target_value = 75
    }
  }
}
`, rName, tagName))
}

func testAccAutoScalingPlansScalingPlanConfigBasicPredictiveScaling(rName, tagName string) string {
	return composeConfig(
		testAccAutoScalingPlansScalingPlanConfigBase(rName, tagName),
		fmt.Sprintf(`
resource "aws_autoscalingplans_scaling_plan" "test" {
  name = %[1]q

  application_source {
    tag_filter {
      key    = %[2]q
      values = [%[2]q]
    }
  }

  scaling_instruction {
    disable_dynamic_scaling = true

    max_capacity       = aws_autoscaling_group.test.max_size
    min_capacity       = aws_autoscaling_group.test.min_size
    resource_id        = format("autoScalingGroup/%%s", aws_autoscaling_group.test.name)
    scalable_dimension = "autoscaling:autoScalingGroup:DesiredCapacity"
    service_namespace  = "autoscaling"

    target_tracking_configuration {
      predefined_scaling_metric_specification {
        predefined_scaling_metric_type = "ASGAverageCPUUtilization"
      }

      target_value = 75
    }

    predictive_scaling_max_capacity_behavior = "SetForecastCapacityToMaxCapacity"
    predictive_scaling_mode                  = "ForecastOnly"

    predefined_load_metric_specification {
      predefined_load_metric_type = "ASGTotalCPUUtilization"
    }
  }
}
`, rName, tagName))
}
