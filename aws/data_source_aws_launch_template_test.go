package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSLaunchTemplateDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_launch_template.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateDataSourceConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "default_version", dataSourceName, "default_version"),
					resource.TestCheckResourceAttrPair(resourceName, "latest_version", dataSourceName, "latest_version"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "hibernation_options", dataSourceName, "hibernation_options"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplateDataSource_filter_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_launch_template.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateDataSourceConfigBasicFilter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "default_version", dataSourceName, "default_version"),
					resource.TestCheckResourceAttrPair(resourceName, "latest_version", dataSourceName, "latest_version"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
					resource.TestCheckResourceAttrPair(resourceName, "placement", dataSourceName, "placement"),
					resource.TestCheckResourceAttrPair(resourceName, "license_specification", dataSourceName, "license_specification"),
					resource.TestCheckResourceAttrPair(resourceName, "monitoring", dataSourceName, "monitoring"),
					resource.TestCheckResourceAttrPair(resourceName, "network_interfaces", dataSourceName, "network_interfaces"),
					resource.TestCheckResourceAttrPair(resourceName, "ram_disk_id", dataSourceName, "ram_disk_id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_security_group_ids", dataSourceName, "vpc_security_group_ids"),
					resource.TestCheckResourceAttrPair(resourceName, "tag_specifications", dataSourceName, "tag_specifications"),
					resource.TestCheckResourceAttrPair(resourceName, "user_data", dataSourceName, "user_data"),
					resource.TestCheckResourceAttrPair(resourceName, "key_name", dataSourceName, "key_name"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_type", dataSourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_market_options", dataSourceName, "instance_market_options"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_initiated_shutdown_behavior", dataSourceName, "instance_initiated_shutdown_behavior"),
					resource.TestCheckResourceAttrPair(resourceName, "image_id", dataSourceName, "image_id"),
					resource.TestCheckResourceAttrPair(resourceName, "iam_instance_profile", dataSourceName, "iam_instance_profile"),
					resource.TestCheckResourceAttrPair(resourceName, "elastic_inference_accelerator", dataSourceName, "elastic_inference_accelerator"),
					resource.TestCheckResourceAttrPair(resourceName, "elastic_gpu_specifications", dataSourceName, "elastic_gpu_specifications"),
					resource.TestCheckResourceAttrPair(resourceName, "ebs_optimized", dataSourceName, "ebs_optimized"),
					resource.TestCheckResourceAttrPair(resourceName, "disable_api_termination", dataSourceName, "disable_api_termination"),
					resource.TestCheckResourceAttrPair(resourceName, "credit_specification", dataSourceName, "credit_specification"),
					resource.TestCheckResourceAttrPair(resourceName, "capacity_reservation_specification", dataSourceName, "capacity_reservation_specification"),
					resource.TestCheckResourceAttrPair(resourceName, "block_device_mappings", dataSourceName, "block_device_mappings")),
			},
		},
	})
}

func TestAccAWSLaunchTemplateDataSource_filter_tags(t *testing.T) {
	rInt := acctest.RandInt()
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_launch_template.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateDataSourceConfigFilterTags(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "default_version", dataSourceName, "default_version"),
					resource.TestCheckResourceAttrPair(resourceName, "latest_version", dataSourceName, "latest_version"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplateDataSource_metadataOptions(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_launch_template.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateDataSourceConfig_metadataOptions(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "metadata_options.#", resourceName, "metadata_options.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "metadata_options.0.http_endpoint", resourceName, "metadata_options.0.http_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "metadata_options.0.http_tokens", resourceName, "metadata_options.0.http_tokens"),
					resource.TestCheckResourceAttrPair(dataSourceName, "metadata_options.0.http_put_response_hop_limit", resourceName, "metadata_options.0.http_put_response_hop_limit"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplateDataSource_associatePublicIPAddress(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_launch_template.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateDataSourceConfig_associatePublicIpAddress(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.#", resourceName, "network_interfaces.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.0.associate_public_ip_address", resourceName, "network_interfaces.0.associate_public_ip_address"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateDataSourceConfig_associatePublicIpAddress(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.#", resourceName, "network_interfaces.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.0.associate_public_ip_address", resourceName, "network_interfaces.0.associate_public_ip_address"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateDataSourceConfig_associatePublicIpAddress(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.#", resourceName, "network_interfaces.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.0.associate_public_ip_address", resourceName, "network_interfaces.0.associate_public_ip_address"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplateDataSource_networkInterfaces_deleteOnTermination(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_launch_template.test"
	resourceName := "aws_launch_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateDataSourceConfigNetworkInterfacesDeleteOnTermination(rName, "true"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.#", resourceName, "network_interfaces.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.0.delete_on_termination", resourceName, "network_interfaces.0.delete_on_termination"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateDataSourceConfigNetworkInterfacesDeleteOnTermination(rName, "false"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.#", resourceName, "network_interfaces.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.0.delete_on_termination", resourceName, "network_interfaces.0.delete_on_termination"),
				),
			},
			{
				Config: testAccAWSLaunchTemplateDataSourceConfigNetworkInterfacesDeleteOnTermination(rName, "null"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.#", resourceName, "network_interfaces.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_interfaces.0.delete_on_termination", resourceName, "network_interfaces.0.delete_on_termination"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplateDataSource_NonExistent(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSLaunchTemplateDataSourceConfig_NonExistent,
				ExpectError: regexp.MustCompile(`not found`),
			},
		},
	})
}

func testAccAWSLaunchTemplateDataSourceConfig_Basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %q
}

data "aws_launch_template" "test" {
  name = aws_launch_template.test.name
}
`, rName)
}

func testAccAWSLaunchTemplateDataSourceConfigBasicFilter(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q
}

data "aws_launch_template" "test" {
  filter {
    name   = "launch-template-name"
    values = [aws_launch_template.test.name]
  }
}
`, rName)
}

func testAccAWSLaunchTemplateDataSourceConfigFilterTags(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  tags = {
    Name     = "key1"
    TestSeed = "%[2]d"
  }
}

data "aws_launch_template" "test" {
  tags = {
    Name     = aws_launch_template.test.tags["Name"]
    TestSeed = "%[2]d"
  }
}
`, rName, rInt)
}

func testAccAWSLaunchTemplateDataSourceConfig_metadataOptions(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 2
  }
}

data "aws_launch_template" "test" {
  name = aws_launch_template.test.name
}
`, rName)
}

func testAccAWSLaunchTemplateDataSourceConfig_associatePublicIpAddress(rName, associatePublicIPAddress string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    associate_public_ip_address = %[2]s
  }
}

data "aws_launch_template" "test" {
  name = aws_launch_template.test.name
}
`, rName, associatePublicIPAddress)
}

func testAccAWSLaunchTemplateDataSourceConfigNetworkInterfacesDeleteOnTermination(rName, deleteOnTermination string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %[1]q

  network_interfaces {
    delete_on_termination = %[2]s
  }
}

data "aws_launch_template" "test" {
  name = aws_launch_template.test.name
}
`, rName, deleteOnTermination)
}

const testAccAWSLaunchTemplateDataSourceConfig_NonExistent = `
data "aws_launch_template" "test" {
  name = "tf-acc-test-nonexistent"
}
`
