package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsMqInstanceBrokerTypeOfferingsDataSource_InstanceType(t *testing.T) {
	dataSourceName := "data.aws_mq_broker_instance_type_offerings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsMqInstanceBrokerTypeOfferings(t) },
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMqInstanceBrokerTypeOfferingsDataSourceConfigHostInstanceType(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqInstanceBrokerTypeOfferingsInstanceTypes(dataSourceName),
				),
			},
		},
	})
}

func TestAccAwsMqInstanceBrokerTypeOfferingsDataSource_EngineType(t *testing.T) {
	dataSourceName := "data.aws_mq_broker_instance_type_offerings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsMqInstanceBrokerTypeOfferings(t) },
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMqInstanceBrokerTypeOfferingsDataSourceConfigEngineType(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqInstanceBrokerTypeOfferingsInstanceTypes(dataSourceName),
				),
			},
		},
	})
}

func TestAccAwsMqInstanceBrokerTypeOfferingsDataSource_StorageType(t *testing.T) {
	dataSourceName := "data.aws_mq_broker_instance_type_offerings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsMqInstanceBrokerTypeOfferings(t) },
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMqInstanceBrokerTypeOfferingsDataSourceConfigStorageType(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqInstanceBrokerTypeOfferingsInstanceTypes(dataSourceName),
				),
			},
		},
	})
}

func TestAccAwsMqInstanceBrokerTypeOfferingsDataSource_All(t *testing.T) {
	dataSourceName := "data.aws_mq_broker_instance_type_offerings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsMqInstanceBrokerTypeOfferings(t) },
		ErrorCheck:   testAccErrorCheck(t, mq.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsMqInstanceBrokerTypeOfferingsDataSourceConfigAllTypes(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMqInstanceBrokerTypeOfferingsInstanceTypes(dataSourceName),
				),
			},
		},
	})
}

func testAccCheckAwsMqInstanceBrokerTypeOfferingsInstanceTypes(dataSourceName string) resource.TestCheckFunc {
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

func testAccPreCheckAwsMqInstanceBrokerTypeOfferings(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).mqconn

	input := &mq.DescribeBrokerInstanceOptionsInput{
		MaxResults: aws.Int64(5),
	}

	_, err := conn.DescribeBrokerInstanceOptions(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAwsMqInstanceBrokerTypeOfferingsDataSourceConfigHostInstanceType() string {
	return `
data "aws_mq_broker_instance_type_offerings" "test" {
  host_instance_type = "mq.m5.large"
}
`
}

func testAccAwsMqInstanceBrokerTypeOfferingsDataSourceConfigEngineType() string {
	return `
data "aws_mq_broker_instance_type_offerings" "test" {
  engine_type = "RABBITMQ"
}
`
}

func testAccAwsMqInstanceBrokerTypeOfferingsDataSourceConfigStorageType() string {
	return `
data "aws_mq_broker_instance_type_offerings" "test" {
  storage_type = "EBS"
}
`
}

func testAccAwsMqInstanceBrokerTypeOfferingsDataSourceConfigAllTypes() string {
	return `
data "aws_mq_broker_instance_type_offerings" "test" {
  storage_type       = "EBS"
  engine_type        = "RABBITMQ"
  host_instance_type = "mq.m5.large"
}
`
}
