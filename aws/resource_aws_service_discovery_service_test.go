package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSServiceDiscoveryService_private(t *testing.T) {
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryServiceConfig_private(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryServiceExists("aws_service_discovery_service.test"),
					resource.TestCheckResourceAttr("aws_service_discovery_service.test", "health_check_custom_config.0.failure_threshold", "5"),
					resource.TestCheckResourceAttr("aws_service_discovery_service.test", "dns_config.0.dns_records.#", "1"),
					resource.TestCheckResourceAttr("aws_service_discovery_service.test", "dns_config.0.dns_records.0.type", "A"),
					resource.TestCheckResourceAttr("aws_service_discovery_service.test", "dns_config.0.dns_records.0.ttl", "5"),
					resource.TestCheckResourceAttr("aws_service_discovery_service.test", "dns_config.0.routing_policy", "MULTIVALUE"),
					resource.TestCheckResourceAttrSet("aws_service_discovery_service.test", "arn"),
				),
			},
			{
				Config: testAccServiceDiscoveryServiceConfig_private_update(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryServiceExists("aws_service_discovery_service.test"),
					resource.TestCheckResourceAttr("aws_service_discovery_service.test", "dns_config.0.dns_records.#", "2"),
					resource.TestCheckResourceAttr("aws_service_discovery_service.test", "dns_config.0.dns_records.0.type", "A"),
					resource.TestCheckResourceAttr("aws_service_discovery_service.test", "dns_config.0.dns_records.0.ttl", "10"),
					resource.TestCheckResourceAttr("aws_service_discovery_service.test", "dns_config.0.dns_records.1.type", "AAAA"),
					resource.TestCheckResourceAttr("aws_service_discovery_service.test", "dns_config.0.dns_records.1.ttl", "5"),
					resource.TestCheckResourceAttrSet("aws_service_discovery_service.test", "arn"),
				),
			},
		},
	})
}

func TestAccAWSServiceDiscoveryService_public(t *testing.T) {
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryServiceConfig_public(rName, 5, "/path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryServiceExists("aws_service_discovery_service.test"),
					resource.TestCheckResourceAttr("aws_service_discovery_service.test", "health_check_config.0.type", "HTTP"),
					resource.TestCheckResourceAttr("aws_service_discovery_service.test", "health_check_config.0.failure_threshold", "5"),
					resource.TestCheckResourceAttr("aws_service_discovery_service.test", "health_check_config.0.resource_path", "/path"),
					resource.TestCheckResourceAttr("aws_service_discovery_service.test", "dns_config.0.routing_policy", "WEIGHTED"),
					resource.TestCheckResourceAttrSet("aws_service_discovery_service.test", "arn"),
				),
			},
			{
				Config: testAccServiceDiscoveryServiceConfig_public(rName, 3, "/updated-path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryServiceExists("aws_service_discovery_service.test"),
					resource.TestCheckResourceAttr("aws_service_discovery_service.test", "health_check_config.0.type", "HTTP"),
					resource.TestCheckResourceAttr("aws_service_discovery_service.test", "health_check_config.0.failure_threshold", "3"),
					resource.TestCheckResourceAttr("aws_service_discovery_service.test", "health_check_config.0.resource_path", "/updated-path"),
					resource.TestCheckResourceAttrSet("aws_service_discovery_service.test", "arn"),
				),
			},
		},
	})
}

func TestAccAWSServiceDiscoveryService_import(t *testing.T) {
	resourceName := "aws_service_discovery_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryServiceConfig_private(acctest.RandString(5), 5),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsServiceDiscoveryServiceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sdconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_service_discovery_service" {
			continue
		}

		input := &servicediscovery.GetServiceInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetService(input)
		if err != nil {
			if isAWSErr(err, servicediscovery.ErrCodeServiceNotFound, "") {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckAwsServiceDiscoveryServiceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).sdconn

		input := &servicediscovery.GetServiceInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetService(input)
		return err
	}
}

func testAccServiceDiscoveryServiceConfig_private(rName string, th int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-service-discovery-service-private"
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name = "tf-sd-%s.terraform.local"
  description = "test"
  vpc = "${aws_vpc.test.id}"
}

resource "aws_service_discovery_service" "test" {
  name = "tf-sd-%s"
  dns_config {
    namespace_id = "${aws_service_discovery_private_dns_namespace.test.id}"
    dns_records {
      ttl = 5
      type = "A"
    }
  }
  health_check_custom_config {
    failure_threshold = %d
  }
}
`, rName, rName, th)
}

func testAccServiceDiscoveryServiceConfig_private_update(rName string, th int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-service-discovery-service-private"
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name = "tf-sd-%s.terraform.local"
  description = "test"
  vpc = "${aws_vpc.test.id}"
}

resource "aws_service_discovery_service" "test" {
  name = "tf-sd-%s"
  dns_config {
    namespace_id = "${aws_service_discovery_private_dns_namespace.test.id}"
    dns_records {
      ttl = 10
      type = "A"
    }
    dns_records {
      ttl = 5
      type = "AAAA"
    }
    routing_policy = "MULTIVALUE"
  }
  health_check_custom_config {
    failure_threshold = %d
  }
}
`, rName, rName, th)
}

func testAccServiceDiscoveryServiceConfig_public(rName string, th int, path string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_public_dns_namespace" "test" {
  name = "tf-sd-%s.terraform.com"
  description = "test"
}

resource "aws_service_discovery_service" "test" {
  name = "tf-sd-%s"
  dns_config {
    namespace_id = "${aws_service_discovery_public_dns_namespace.test.id}"
    dns_records {
      ttl = 5
      type = "A"
    }
    routing_policy = "WEIGHTED"
  }
  health_check_config {
    failure_threshold = %d
    resource_path = "%s"
    type = "HTTP"
  }
}
`, rName, rName, th, path)
}
