package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEbsVolumeDataSource_basic(t *testing.T) {
	resourceName := "aws_ebs_volume.test"
	dataSourceName := "data.aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEbsVolumeDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEbsVolumeDataSourceID(dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "size", resourceName, "size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags", resourceName, "tags"),
					resource.TestCheckResourceAttrPair(dataSourceName, "outpost_arn", resourceName, "outpost_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "multi_attach_enabled", resourceName, "multi_attach_enabled"),
				),
			},
		},
	})
}

func TestAccAWSEbsVolumeDataSource_multipleFilters(t *testing.T) {
	resourceName := "aws_ebs_volume.test"
	dataSourceName := "data.aws_ebs_volume.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEbsVolumeDataSourceConfigWithMultipleFilters,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEbsVolumeDataSourceID(dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "size", resourceName, "size"),
					resource.TestCheckResourceAttr(dataSourceName, "volume_type", "gp2"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags", resourceName, "tags"),
				),
			},
		},
	})
}

func testAccCheckAwsEbsVolumeDataSourceID(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find Volume data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Volume data source ID not set")
		}
		return nil
	}
}

var testAccCheckAwsEbsVolumeDataSourceConfig = testAccAvailableAZsNoOptInConfig() + `
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "gp2"
  size              = 40

  tags = {
    Name = "External Volume"
  }
}

data "aws_ebs_volume" "test" {
  most_recent = true

  filter {
    name   = "tag:Name"
    values = ["External Volume"]
  }

  filter {
    name   = "volume-type"
    values = [aws_ebs_volume.test.type]
  }
}
`

var testAccCheckAwsEbsVolumeDataSourceConfigWithMultipleFilters = testAccAvailableAZsNoOptInConfig() + `
resource "aws_ebs_volume" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  type              = "gp2"
  size              = 10

  tags = {
    Name = "External Volume 1"
  }
}

data "aws_ebs_volume" "test" {
  most_recent = true

  filter {
    name   = "tag:Name"
    values = ["External Volume 1"]
  }

  filter {
    name   = "size"
    values = [aws_ebs_volume.test.size]
  }

  filter {
    name   = "volume-type"
    values = [aws_ebs_volume.test.type]
  }
}
`
