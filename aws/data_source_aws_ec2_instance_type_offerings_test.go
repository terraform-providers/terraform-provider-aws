package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEc2InstanceTypeOfferingsDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_instance_type_offerings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2InstanceTypeOfferings(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2InstanceTypeOfferingsDataSourceConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2InstanceTypeOfferingsInstanceTypes(dataSourceName),
				),
			},
		},
	})
}

func TestAccAWSEc2InstanceTypeOfferingsDataSource_LocationType(t *testing.T) {
	dataSourceName := "data.aws_ec2_instance_type_offerings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2InstanceTypeOfferings(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2InstanceTypeOfferingsDataSourceConfigLocationType(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEc2InstanceTypeOfferingsInstanceTypes(dataSourceName),
				),
			},
		},
	})
}

func testAccCheckEc2InstanceTypeOfferingsInstanceTypes(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", dataSourceName)
		}

		if v := rs.Primary.Attributes["instance_types.#"]; v == "0" {
			return fmt.Errorf("expected at least one instance_types result, got none")
		}

		return nil
	}
}

func testAccPreCheckAWSEc2InstanceTypeOfferings(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	input := &ec2.DescribeInstanceTypeOfferingsInput{
		MaxResults: aws.Int64(5),
	}

	_, err := conn.DescribeInstanceTypeOfferings(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSEc2InstanceTypeOfferingsDataSourceConfigFilter() string {
	return `
data "aws_ec2_instance_type_offerings" "test" {
  filter {
    name   = "instance-type"
    values = ["t2.micro", "t3.micro"]
  }
}
`
}

func testAccAWSEc2InstanceTypeOfferingsDataSourceConfigLocationType() string {
	return testAccAvailableAZsNoOptInConfig() + `
data "aws_ec2_instance_type_offerings" "test" {
  filter {
    name   = "location"
    values = [data.aws_availability_zones.available.names[0]]
  }

  location_type = "availability-zone"
}
`
}
