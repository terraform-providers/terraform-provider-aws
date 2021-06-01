package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_launch_template", &resource.Sweeper{
		Name: "aws_launch_template",
		Dependencies: []string{
			"aws_autoscaling_group",
			"aws_batch_compute_environment",
		},
		F: testSweepLaunchTemplates,
	})
}

func testSweepLaunchTemplates(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).ec2conn
	input := &ec2.DescribeLaunchTemplatesInput{}
	var sweeperErrs *multierror.Error

	err = conn.DescribeLaunchTemplatesPages(input, func(page *ec2.DescribeLaunchTemplatesOutput, lastPage bool) bool {
		for _, launchTemplate := range page.LaunchTemplates {
			id := aws.StringValue(launchTemplate.LaunchTemplateId)
			input := &ec2.DeleteLaunchTemplateInput{
				LaunchTemplateId: launchTemplate.LaunchTemplateId,
			}

			log.Printf("[INFO] Deleting EC2 Launch Template: %s", id)
			_, err := conn.DeleteLaunchTemplate(input)

			if isAWSErr(err, "InvalidLaunchTemplateId.NotFound", "") {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting EC2 Launch Template (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Launch Template sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error describing EC2 Launch Templates: %w", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSLaunchTemplate_basic(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "default_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "1"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`launch-template/.+`)),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", ""),
					resource.TestCheckResourceAttr(resourceName, "elastic_inference_accelerator.#", "0"),
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

func TestAccAWSLaunchTemplate_disappears(t *testing.T) {
	var launchTemplate ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &launchTemplate),
					testAccCheckAWSLaunchTemplateDisappears(&launchTemplate),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLaunchTemplate_BlockDeviceMappings_EBS(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_BlockDeviceMappings_EBS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.device_name", "/dev/xvda"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.throughput", "0"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.volume_size", "15"),
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

func TestAccAWSLaunchTemplate_BlockDeviceMappings_EBS_DeleteOnTermination(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_BlockDeviceMappings_EBS_DeleteOnTermination(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.device_name", "/dev/xvda"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.delete_on_termination", "true"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.volume_size", "15"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateConfig_BlockDeviceMappings_EBS_DeleteOnTermination(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.device_name", "/dev/xvda"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.delete_on_termination", "false"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.volume_size", "15"),
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

func TestAccAWSLaunchTemplate_BlockDeviceMappings_EBS_Gp3(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_BlockDeviceMappings_EBS_Gp3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.device_name", "/dev/xvda"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.iops", "4000"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.throughput", "500"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.volume_size", "15"),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.0.ebs.0.volume_type", "gp3"),
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

func TestAccAWSLaunchTemplate_EbsOptimized(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_EbsOptimized(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLaunchTemplateConfig_EbsOptimized(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "false"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateConfig_EbsOptimized(rName, "\"true\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "true"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateConfig_EbsOptimized(rName, "\"false\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "false"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateConfig_EbsOptimized(rName, "\"\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", ""),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_ElasticInferenceAccelerator(t *testing.T) {
	var template1 ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfigElasticInferenceAccelerator(rName, "eia1.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template1),
					resource.TestCheckResourceAttr(resourceName, "elastic_inference_accelerator.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "elastic_inference_accelerator.0.type", "eia1.medium"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLaunchTemplateConfigElasticInferenceAccelerator(rName, "eia1.large"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template1),
					resource.TestCheckResourceAttr(resourceName, "elastic_inference_accelerator.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "elastic_inference_accelerator.0.type", "eia1.large"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_NetworkInterfaces_DeleteOnTermination(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_NetworkInterfaces_DeleteOnTermination(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.delete_on_termination", "true"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateConfig_NetworkInterfaces_DeleteOnTermination(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.delete_on_termination", "false"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateConfig_NetworkInterfaces_DeleteOnTermination(rName, "\"\""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.delete_on_termination", ""),
				),
			},
			{
				Config: testAccAWSLaunchTemplateConfig_NetworkInterfaces_DeleteOnTermination(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.delete_on_termination", ""),
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

func TestAccAWSLaunchTemplate_data(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_data(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "block_device_mappings.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "disable_api_termination"),
					resource.TestCheckResourceAttr(resourceName, "ebs_optimized", "false"),
					resource.TestCheckResourceAttr(resourceName, "elastic_gpu_specifications.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "iam_instance_profile.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "image_id"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_initiated_shutdown_behavior"),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_type"),
					resource.TestCheckResourceAttrSet(resourceName, "kernel_id"),
					resource.TestCheckResourceAttrSet(resourceName, "key_name"),
					resource.TestCheckResourceAttr(resourceName, "monitoring.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.security_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.delete_on_termination", ""),
					resource.TestCheckResourceAttr(resourceName, "placement.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "ram_disk_id"),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tag_specifications.#", "4"),
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

func TestAccAWSLaunchTemplate_description(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_description(rName, "Test Description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "description", "Test Description 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLaunchTemplateConfig_description(rName, "Test Description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "description", "Test Description 2"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_update(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_asg_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "default_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "1"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "launch_template.0.version", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
			{
				Config: testAccAWSLaunchTemplateConfig_asg_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "default_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_type"),
					resource.TestCheckResourceAttr(
						"aws_autoscaling_group.bar", "launch_template.0.version", "2"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_tags(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.test", "bar"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLaunchTemplateConfig_tagsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.bar", "baz"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_capacityReservation_preference(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_capacityReservation_preference(rName, "open"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
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

func TestAccAWSLaunchTemplate_capacityReservation_target(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_capacityReservation_target(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
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

func TestAccAWSLaunchTemplate_cpuOptions(t *testing.T) {
	var template ec2.LaunchTemplate
	resName := "aws_launch_template.foo"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_cpuOptions(rName, 4, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_creditSpecification_nonBurstable(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_creditSpecification(rName, "m1.small", "standard"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"credit_specification"},
			},
		},
	})
}

func TestAccAWSLaunchTemplate_creditSpecification_t2(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_creditSpecification(rName, "t2.micro", "unlimited"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
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

func TestAccAWSLaunchTemplate_creditSpecification_t3(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_creditSpecification(rName, "t3.micro", "unlimited"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credit_specification.0.cpu_credits", "unlimited"),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/6757
func TestAccAWSLaunchTemplate_IamInstanceProfile_EmptyConfigurationBlock(t *testing.T) {
	var template1 ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfigIamInstanceProfileEmptyConfigurationBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLaunchTemplate_networkInterface(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_networkInterface(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_public_ip_address", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.delete_on_termination", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", "2"),
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

func TestAccAWSLaunchTemplate_networkInterfaceAddresses(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_networkInterfaceAddresses(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_public_ip_address", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_addresses.#", "2"),
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

func TestAccAWSLaunchTemplate_associatePublicIPAddress(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_associatePublicIpAddress(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_public_ip_address", "true"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLaunchTemplateConfig_associatePublicIpAddress(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_public_ip_address", "false"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", "2"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateConfig_associatePublicIpAddress(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_public_ip_address", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", "2"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_associateCarrierIPAddress(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_associateCarrierIpAddress(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_carrier_ip_address", "true"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLaunchTemplateConfig_associateCarrierIpAddress(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_carrier_ip_address", "false"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", "2"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateConfig_associateCarrierIpAddress(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interfaces.0.network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.associate_carrier_ip_address", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv4_address_count", "2"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_placement_partitionNum(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfigPartition(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "placement.0.partition_number", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLaunchTemplateConfigPartition(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "placement.0.partition_number", "2"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_networkInterface_ipv6Addresses(t *testing.T) {
	var template ec2.LaunchTemplate
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_networkInterface_ipv6Addresses(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv6_addresses.#", "2"),
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

func TestAccAWSLaunchTemplate_networkInterface_ipv6AddressCount(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_ipv6_count(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_interfaces.0.ipv6_address_count", "1"),
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

func TestAccAWSLaunchTemplate_instanceMarketOptions(t *testing.T) {
	var template ec2.LaunchTemplate
	var group autoscaling.Group
	groupName := "aws_autoscaling_group.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_instanceMarketOptions_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					testAccCheckAWSAutoScalingGroupExists(groupName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.0.spot_options.#", "1"),
					resource.TestCheckResourceAttr(groupName, "launch_template.#", "1"),
					resource.TestCheckResourceAttr(groupName, "launch_template.0.version", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix", "instance_market_options"},
			},
			{
				Config: testAccAWSLaunchTemplateConfig_instanceMarketOptions_update,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					testAccCheckAWSAutoScalingGroupExists(groupName, &group),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_market_options.0.spot_options.#", "1"),
					resource.TestCheckResourceAttr(groupName, "launch_template.#", "1"),
					resource.TestCheckResourceAttr(groupName, "launch_template.0.version", "2"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_licenseSpecification(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.example"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_licenseSpecification(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "license_specification.#", "1"),
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

func TestAccAWSLaunchTemplate_metadataOptions(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_metadataOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_endpoint", "enabled"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_tokens", "required"),
					resource.TestCheckResourceAttr(resourceName, "metadata_options.0.http_put_response_hop_limit", "2"),
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

func TestAccAWSLaunchTemplate_enclaveOptions(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_enclaveOptions(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "enclave_options.0.enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLaunchTemplateConfig_enclaveOptions(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "enclave_options.0.enabled", "false"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateConfig_enclaveOptions(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "enclave_options.0.enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_hibernation(t *testing.T) {
	var template ec2.LaunchTemplate
	resourceName := "aws_launch_template.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfigHibernation(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "hibernation_options.0.configured", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLaunchTemplateConfigHibernation(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "hibernation_options.0.configured", "false"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateConfigHibernation(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resourceName, &template),
					resource.TestCheckResourceAttr(resourceName, "hibernation_options.0.configured", "true"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_defaultVersion(t *testing.T) {
	resourceName := "aws_launch_template.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	description := "Test Description 1"
	descriptionNew := "Test Description 2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_description(rName, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "1"),
				),
			},
			// An updated config should cause a new version to be created
			// but keep the default_version unchanged if unset
			{
				Config: testAccAWSLaunchTemplateConfig_description(rName, descriptionNew),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "2"),
				),
			},
			// An updated config to set the default_version to an existing version
			// should not cause a new version to be created
			{
				Config: testAccAWSLaunchTemplateConfig_descriptionDefaultVersion(rName, descriptionNew, 2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", "2"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "2"),
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

func TestAccAWSLaunchTemplate_updateDefaultVersion(t *testing.T) {
	resourceName := "aws_launch_template.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	description := "Test Description 1"
	descriptionNew := "Test Description 2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_description(rName, description),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "1"),
				),
			},
			// Updating a field should create a new version but not update the default_version
			// if update_default_version is set to false
			{
				Config: testAccAWSLaunchTemplateconfig_descriptionUpdateDefaultVersion(rName, descriptionNew, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", "1"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "2"),
				),
			},
			// Only updating the update_default_version to true should not create a new version
			// but update the template version to the latest available
			{
				Config: testAccAWSLaunchTemplateconfig_descriptionUpdateDefaultVersion(rName, descriptionNew, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", "2"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "2"),
				),
			},
			// Updating a field should create a new version and update the default_version
			// if update_default_version is set to true
			{
				Config: testAccAWSLaunchTemplateconfig_descriptionUpdateDefaultVersion(rName, description, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version", "3"),
					resource.TestCheckResourceAttr(resourceName, "latest_version", "3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"update_default_version",
				},
			},
		},
	})
}

func testAccCheckAWSLaunchTemplateExists(n string, t *ec2.LaunchTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Launch Template ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		resp, err := conn.DescribeLaunchTemplates(&ec2.DescribeLaunchTemplatesInput{
			LaunchTemplateIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}

		if len(resp.LaunchTemplates) != 1 || *resp.LaunchTemplates[0].LaunchTemplateId != rs.Primary.ID {
			return fmt.Errorf("Launch Template not found")
		}

		*t = *resp.LaunchTemplates[0]

		return nil
	}
}

func testAccCheckAWSLaunchTemplateDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_launch_template" {
			continue
		}

		resp, err := conn.DescribeLaunchTemplates(&ec2.DescribeLaunchTemplatesInput{
			LaunchTemplateIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err == nil {
			if len(resp.LaunchTemplates) != 0 && *resp.LaunchTemplates[0].LaunchTemplateId == rs.Primary.ID {
				return fmt.Errorf("Launch Template still exists")
			}
		}

		if isAWSErr(err, "InvalidLaunchTemplateId.NotFound", "") {
			log.Printf("[WARN] launch template (%s) not found.", rs.Primary.ID)
			continue
		}
		return err
	}

	return nil
}

func testAccCheckAWSLaunchTemplateDisappears(launchTemplate *ec2.LaunchTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		input := &ec2.DeleteLaunchTemplateInput{
			LaunchTemplateId: launchTemplate.LaunchTemplateId,
		}

		_, err := conn.DeleteLaunchTemplate(input)

		return err
	}
}

func testAccAWSLaunchTemplateConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %q

  tags = {
    test = "bar"
  }
}
`, rName)
}

func testAccAWSLaunchTemplateConfig_ipv6_count(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %q

  network_interfaces {
    ipv6_address_count = 1
  }
}
`, rName)
}

func testAccAWSLaunchTemplateConfig_BlockDeviceMappings_EBS(rName string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableAZsNoOptInConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  name          = %[1]q

  block_device_mappings {
    device_name = "/dev/xvda"

    ebs {
      volume_size = 15
    }
  }
}

# Creating an AutoScaling Group verifies the launch template
# ValidationError: You must use a valid fully-formed launch template. the encrypted flag cannot be specified since device /dev/sda1 has a snapshot specified.
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.default_version
  }
}
`, rName))
}

func testAccAWSLaunchTemplateConfig_BlockDeviceMappings_EBS_DeleteOnTermination(rName string, deleteOnTermination bool) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableAZsNoOptInConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  name          = %[1]q

  block_device_mappings {
    device_name = "/dev/xvda"

    ebs {
      delete_on_termination = %t
      volume_size           = 15
    }
  }
}

# Creating an AutoScaling Group verifies the launch template
# ValidationError: You must use a valid fully-formed launch template. the encrypted flag cannot be specified since device /dev/sda1 has a snapshot specified.
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.default_version
  }
}
`, rName, deleteOnTermination))
}

func testAccAWSLaunchTemplateConfig_BlockDeviceMappings_EBS_Gp3(rName string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableAZsNoOptInConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  name          = %[1]q

  block_device_mappings {
    device_name = "/dev/xvda"

    ebs {
      iops        = 4000
      throughput  = 500
      volume_size = 15
      volume_type = "gp3"
    }
  }
}

# Creating an AutoScaling Group verifies the launch template
# ValidationError: You must use a valid fully-formed launch template. the encrypted flag cannot be specified since device /dev/sda1 has a snapshot specified.
resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.default_version
  }
}
`, rName))
}

func testAccAWSLaunchTemplateConfig_NetworkInterfaces_DeleteOnTermination(rName string, deleteOnTermination string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %q

  network_interfaces {
    network_interface_id  = "eni-123456ab"
    security_groups       = ["sg-1a23bc45"]
    delete_on_termination = %s
  }
}
`, rName, deleteOnTermination)
}

func testAccAWSLaunchTemplateConfig_EbsOptimized(rName, ebsOptimized string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  ebs_optimized = %s
  name          = %q
}
`, ebsOptimized, rName)
}

func testAccAWSLaunchTemplateConfigElasticInferenceAccelerator(rName, elasticInferenceAcceleratorType string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  elastic_inference_accelerator {
    type = %[2]q
  }
}
`, rName, elasticInferenceAcceleratorType)
}

func testAccAWSLaunchTemplateConfig_data(rName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %q

  block_device_mappings {
    device_name = "test"
  }

  disable_api_termination = true
  ebs_optimized           = false

  elastic_gpu_specifications {
    type = "test"
  }

  iam_instance_profile {
    name = "test"
  }

  image_id                             = "ami-12a3b456"
  instance_initiated_shutdown_behavior = "terminate"

  instance_market_options {
    market_type = "spot"
  }

  instance_type = "t2.micro"
  kernel_id     = "aki-a12bc3de"
  key_name      = "test"

  monitoring {
    enabled = true
  }

  network_interfaces {
    network_interface_id = "eni-123456ab"
    security_groups      = ["sg-1a23bc45"]
  }

  placement {
    availability_zone = data.aws_availability_zones.available.names[0]
  }

  ram_disk_id            = "ari-a12bc3de"
  vpc_security_group_ids = ["sg-12a3b45c"]

  tag_specifications {
    resource_type = "instance"

    tags = {
      Name = "test"
    }
  }

  tag_specifications {
    resource_type = "volume"

    tags = {
      Name = "test"
    }
  }

  tag_specifications {
    resource_type = "spot-instances-request"

    tags = {
      Name = "test"
    }
  }

  tag_specifications {
    resource_type = "elastic-gpu"

    tags = {
      Name = "test"
    }
  }
}
`, rName)) //lintignore:AWSAT002
}

func testAccAWSLaunchTemplateConfig_tagsUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %q

  tags = {
    bar = "baz"
  }
}
`, rName)
}

func testAccAWSLaunchTemplateConfig_capacityReservation_preference(rName string, preference string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %q

  capacity_reservation_specification {
    capacity_reservation_preference = %q
  }
}
`, rName, preference)
}

func testAccAWSLaunchTemplateConfig_capacityReservation_target(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ec2_capacity_reservation" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  instance_count    = 1
  instance_platform = "Linux/UNIX"
  instance_type     = "t2.micro"
}

resource "aws_launch_template" "test" {
  name = %q

  capacity_reservation_specification {
    capacity_reservation_target {
      capacity_reservation_id = aws_ec2_capacity_reservation.test.id
    }
  }
}
`, rName)
}

func testAccAWSLaunchTemplateConfig_cpuOptions(rName string, coreCount, threadsPerCore int) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "foo" {
  name = %q

  cpu_options {
    core_count       = %d
    threads_per_core = %d
  }
}
`, rName, coreCount, threadsPerCore)
}

func testAccAWSLaunchTemplateConfig_creditSpecification(rName, instanceType, cpuCredits string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  instance_type = %q
  name          = %q

  credit_specification {
    cpu_credits = %q
  }
}
`, instanceType, rName, cpuCredits)
}

func testAccAWSLaunchTemplateConfigIamInstanceProfileEmptyConfigurationBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %q

  iam_instance_profile {}
}
`, rName)
}

func testAccAWSLaunchTemplateConfig_licenseSpecification(rName string) string {
	return fmt.Sprintf(`
resource "aws_licensemanager_license_configuration" "example" {
  name                  = "Example"
  license_counting_type = "vCPU"
}

resource "aws_launch_template" "example" {
  name = %q

  license_specification {
    license_configuration_arn = aws_licensemanager_license_configuration.example.id
  }
}
`, rName)
}

func testAccAWSLaunchTemplateConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name        = %q
  description = %q
}
`, rName, description)
}

func testAccAWSLaunchTemplateConfig_networkInterface(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.0.0/24"
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id
}

resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    network_interface_id = aws_network_interface.test.id
    ipv4_address_count   = 2
  }
}
`, rName)
}

func testAccAWSLaunchTemplateConfigPartition(rName string, partNum int) string {
	return fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name     = %[1]q
  strategy = "partition"
}

resource "aws_launch_template" "test" {
  name = %[1]q

  placement {
    group_name       = aws_placement_group.test.name
    partition_number = %[2]d
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, partNum)
}

func testAccAWSLaunchTemplateConfig_networkInterfaceAddresses(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.0.0/24"
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id
}

resource "aws_launch_template" "test" {
  name = %q

  network_interfaces {
    network_interface_id = aws_network_interface.test.id
    ipv4_addresses       = ["10.1.0.10", "10.1.0.11"]
  }
}
`, rName)

}

func testAccAWSLaunchTemplateConfig_associatePublicIpAddress(rName, associatePublicIPAddress string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.0.0/24"
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id
}

resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    network_interface_id        = aws_network_interface.test.id
    associate_public_ip_address = %[2]s
    ipv4_address_count          = 2
  }
}
`, rName, associatePublicIPAddress)
}

func testAccAWSLaunchTemplateConfig_associateCarrierIpAddress(rName, associateCarrierIPAddress string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "10.1.0.0/24"
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id
}

resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    network_interface_id         = aws_network_interface.test.id
    associate_carrier_ip_address = %[2]s
    ipv4_address_count           = 2
  }
}
`, rName, associateCarrierIPAddress)
}

func testAccAWSLaunchTemplateConfig_networkInterface_ipv6Addresses(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    ipv6_addresses = [
      "0:0:0:0:0:ffff:a01:5",
      "0:0:0:0:0:ffff:a01:6",
    ]
  }
}
`, rName)
}

const testAccAWSLaunchTemplateConfig_asg_basic = `
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  name_prefix   = "testbar"
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.micro"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.latest_version
  }
}
`

const testAccAWSLaunchTemplateConfig_asg_update = `
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  name_prefix   = "testbar"
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t2.nano"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.latest_version
  }
}
`

const testAccAWSLaunchTemplateConfig_instanceMarketOptions_basic = `
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  name_prefix   = "instance_market_options"
  image_id      = data.aws_ami.test.id
  instance_type = "t2.micro"

  instance_market_options {
    market_type = "spot"

    spot_options {
      spot_instance_type = "one-time"
    }
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  min_size           = 0
  max_size           = 0

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.latest_version
  }
}
`

const testAccAWSLaunchTemplateConfig_instanceMarketOptions_update = `
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_template" "test" {
  name_prefix   = "instance_market_options"
  image_id      = data.aws_ami.test.id
  instance_type = "t2.micro"

  instance_market_options {
    market_type = "spot"

    spot_options {
      max_price          = "0.5"
      spot_instance_type = "one-time"
    }
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_autoscaling_group" "test" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  desired_capacity   = 0
  min_size           = 0
  max_size           = 0

  launch_template {
    id      = aws_launch_template.test.id
    version = aws_launch_template.test.latest_version
  }
}
`

func testAccAWSLaunchTemplateConfig_metadataOptions(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
  }
}
`, rName)
}

func testAccAWSLaunchTemplateConfig_enclaveOptions(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  enclave_options {
    enabled = %[2]t
  }
}
`, rName, enabled)
}

func testAccAWSLaunchTemplateConfigHibernation(rName string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  hibernation_options {
    configured = %[2]t
  }
}
`, rName, enabled)
}

func testAccAWSLaunchTemplateConfig_descriptionDefaultVersion(rName, description string, version int) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name            = %q
  description     = %q
  default_version = %d

  tags = {
    test = "baz"
  }
}
`, rName, description, version)
}

func testAccAWSLaunchTemplateconfig_descriptionUpdateDefaultVersion(rName, description string, update bool) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name                   = %q
  description            = %q
  update_default_version = %t

  tags = {
    test = "baz"
  }
}
`, rName, description, update)
}
