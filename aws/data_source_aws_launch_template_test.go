package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
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
				),
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

func testAccAWSLaunchTemplateDataSourceConfig_Basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name = %q
}

data "aws_launch_template" "test" {
  name = "${aws_launch_template.test.name}"
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
    values = ["${aws_launch_template.test.name}"]
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
    Name     = "${aws_launch_template.test.tags["Name"]}"
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
