package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ecs/waiter"
)

func init() {
	resource.AddTestSweepers("aws_ecs_capacity_provider", &resource.Sweeper{
		Name: "aws_ecs_capacity_provider",
		F:    testSweepEcsCapacityProviders,
		Dependencies: []string{
			"aws_ecs_cluster",
			"aws_ecs_service",
		},
	})
}

func testSweepEcsCapacityProviders(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).ecsconn
	input := &ecs.DescribeCapacityProvidersInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.DescribeCapacityProviders(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping ECS Capacity Provider sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving ECS Capacity Provider: %w", err))
			return sweeperErrs
		}

		for _, capacityProvider := range output.CapacityProviders {
			if capacityProvider == nil {
				continue
			}

			arn := aws.StringValue(capacityProvider.CapacityProviderArn)
			input := &ecs.DeleteCapacityProviderInput{
				CapacityProvider: capacityProvider.CapacityProviderArn,
			}

			if aws.StringValue(capacityProvider.Name) == "FARGATE" || aws.StringValue(capacityProvider.Name) == "FARGATE_SPOT" {
				log.Printf("[INFO] Skipping AWS managed ECS Capacity Provider: %s", arn)
				continue
			}

			if aws.StringValue(capacityProvider.Status) == ecs.CapacityProviderStatusInactive {
				log.Printf("[INFO] Skipping ECS Capacity Provider with INACTIVE status: %s", arn)
				continue
			}

			log.Printf("[INFO] Deleting ECS Capacity Provider: %s", arn)
			_, err := conn.DeleteCapacityProvider(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting ECS Capacity Provider (%s): %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			if _, err := waiter.CapacityProviderInactive(conn, arn); err != nil {
				sweeperErr := fmt.Errorf("error waiting for ECS Capacity Provider (%s) to delete: %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
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

func TestAccAWSEcsCapacityProvider_basic(t *testing.T) {
	var provider ecs.CapacityProvider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsCapacityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsCapacityProviderConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsCapacityProviderExists(resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccCheckResourceAttrRegionalARN(resourceName, "id", "ecs", fmt.Sprintf("capacity-provider/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "id", resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_termination_protection", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.maximum_scaling_step_size", "10000"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.target_capacity", "100"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEcsCapacityProvider_disappears(t *testing.T) {
	var provider ecs.CapacityProvider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsCapacityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsCapacityProviderConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsCapacityProviderExists(resourceName, &provider),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEcsCapacityProvider(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEcsCapacityProvider_ManagedScaling(t *testing.T) {
	var provider ecs.CapacityProvider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsCapacityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsCapacityProviderConfigManagedScaling(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsCapacityProviderExists(resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_termination_protection", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.maximum_scaling_step_size", "10"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.target_capacity", "50"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEcsCapacityProvider_ManagedScalingPartial(t *testing.T) {
	var provider ecs.CapacityProvider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsCapacityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsCapacityProviderConfigManagedScalingPartial(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsCapacityProviderExists(resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_termination_protection", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.maximum_scaling_step_size", "10000"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.target_capacity", "100"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEcsCapacityProvider_Tags(t *testing.T) {
	var provider ecs.CapacityProvider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcsCapacityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcsCapacityProviderConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsCapacityProviderExists(resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEcsCapacityProviderConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsCapacityProviderExists(resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEcsCapacityProviderConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcsCapacityProviderExists(resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

// TODO add an update test config - Reference: https://github.com/aws/containers-roadmap/issues/633

func testAccCheckAWSEcsCapacityProviderDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ecsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecs_capacity_provider" {
			continue
		}

		input := &ecs.DescribeCapacityProvidersInput{
			CapacityProviders: aws.StringSlice([]string{rs.Primary.ID}),
		}

		output, err := conn.DescribeCapacityProviders(input)

		if err != nil {
			return err
		}

		for _, capacityProvider := range output.CapacityProviders {
			if aws.StringValue(capacityProvider.CapacityProviderArn) == rs.Primary.ID && aws.StringValue(capacityProvider.Status) != ecs.CapacityProviderStatusInactive {
				return fmt.Errorf("ECS Capacity Provider (%s) still exists", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAWSEcsCapacityProviderExists(resourceName string, provider *ecs.CapacityProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).ecsconn

		input := &ecs.DescribeCapacityProvidersInput{
			CapacityProviders: []*string{aws.String(rs.Primary.ID)},
			Include:           []*string{aws.String(ecs.CapacityProviderFieldTags)},
		}

		output, err := conn.DescribeCapacityProviders(input)

		if err != nil {
			return fmt.Errorf("error reading ECS Capacity Provider (%s): %s", rs.Primary.ID, err)
		}

		for _, cp := range output.CapacityProviders {
			if aws.StringValue(cp.CapacityProviderArn) == rs.Primary.ID {
				*provider = *cp
				return nil
			}
		}

		return fmt.Errorf("ECS Capacity Provider (%s) not found", rs.Primary.ID)
	}
}

func testAccAWSEcsCapacityProviderConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.micro"
  name          = %[1]q
}

resource "aws_autoscaling_group" "test" {
  availability_zones = data.aws_availability_zones.available.names
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  launch_template {
    id = aws_launch_template.test.id
  }

  tags = [
    {
      key                 = "foo"
      value               = "bar"
      propagate_at_launch = true
    },
  ]
}
`, rName)
}

func testAccAWSEcsCapacityProviderConfig(rName string) string {
	return testAccAWSEcsCapacityProviderConfigBase(rName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %q

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn
  }
}
`, rName)
}

func testAccAWSEcsCapacityProviderConfigManagedScaling(rName string) string {
	return testAccAWSEcsCapacityProviderConfigBase(rName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %q

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn

    managed_scaling {
      maximum_scaling_step_size = 10
      minimum_scaling_step_size = 2
      status                    = "ENABLED"
      target_capacity           = 50
    }
  }
}
`, rName)
}

func testAccAWSEcsCapacityProviderConfigManagedScalingPartial(rName string) string {
	return testAccAWSEcsCapacityProviderConfigBase(rName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %q

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn

    managed_scaling {
      minimum_scaling_step_size = 2
      status                    = "ENABLED"
    }
  }
}
`, rName)
}

func testAccAWSEcsCapacityProviderConfigTags1(rName, tag1Key, tag1Value string) string {
	return testAccAWSEcsCapacityProviderConfigBase(rName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q,
  }

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccAWSEcsCapacityProviderConfigTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return testAccAWSEcsCapacityProviderConfigBase(rName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %q

  tags = {
    %q = %q,
    %q = %q,
  }

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
