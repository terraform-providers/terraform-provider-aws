package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func init() {
	resource.AddTestSweepers("aws_service_discovery_instance", &resource.Sweeper{
		Name: "aws_service_discovery_instance",
		Dependencies: []string{
			"aws_service_discovery_service",
		},
	})
}

func TestAccAWSServiceDiscoveryInstance_private(t *testing.T) {
	resourceName := "aws_service_discovery_instance.instance"
	serviceResourceName := "aws_service_discovery_service.sd_register_instance"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsServiceDiscoveryInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSServiceDiscoveryInstanceBaseConfig(rName),
					testAccAWSServiceDiscoveryInstancePrivateNamespaceConfig(rName),
					testAccAWSServiceDiscoveryInstanceConfig(rName, "AWS_INSTANCE_IPV4 = \"10.0.0.1\" \n    AWS_INSTANCE_IPV6 = \"2001:0db8:85a3:0000:0000:abcd:0001:2345\""),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryInstanceExists(resourceName, serviceResourceName),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", rName),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV4", "10.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV6", "2001:0db8:85a3:0000:0000:abcd:0001:2345"),
				),
			},
			{
				Config: composeConfig(
					testAccAWSServiceDiscoveryInstanceBaseConfig(rName),
					testAccAWSServiceDiscoveryInstancePrivateNamespaceConfig(rName),
					testAccAWSServiceDiscoveryInstanceConfig(rName, "AWS_INSTANCE_IPV4 = \"10.0.0.2\" \n    AWS_INSTANCE_IPV6 = \"2001:0db8:85a3:0000:0000:abcd:0001:2345\""),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryInstanceExists(resourceName, serviceResourceName),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", rName),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV4", "10.0.0.2"),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV6", "2001:0db8:85a3:0000:0000:abcd:0001:2345"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSServiceDiscoveryInstanceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSServiceDiscoveryInstance_public(t *testing.T) {
	resourceName := "aws_service_discovery_instance.instance"
	serviceResourceName := "aws_service_discovery_service.sd_register_instance"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsServiceDiscoveryInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSServiceDiscoveryInstanceBaseConfig(rName),
					testAccAWSServiceDiscoveryInstancePublicNamespaceConfig(rName),
					testAccAWSServiceDiscoveryInstanceConfig(rName, "AWS_INSTANCE_IPV4 = \"52.18.0.2\" \n    CUSTOM_KEY = \"this is a custom value\""),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryInstanceExists(resourceName, serviceResourceName),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", rName),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV4", "52.18.0.2"),
					resource.TestCheckResourceAttr(resourceName, "attributes.CUSTOM_KEY", "this is a custom value"),
				),
			},
			{
				Config: composeConfig(
					testAccAWSServiceDiscoveryInstanceBaseConfig(rName),
					testAccAWSServiceDiscoveryInstancePublicNamespaceConfig(rName),
					testAccAWSServiceDiscoveryInstanceConfig(rName, "AWS_INSTANCE_IPV4 = \"52.18.0.2\" \n    CUSTOM_KEY = \"this is a custom value updated\""),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryInstanceExists(resourceName, serviceResourceName),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", rName),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV4", "52.18.0.2"),
					resource.TestCheckResourceAttr(resourceName, "attributes.CUSTOM_KEY", "this is a custom value updated"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSServiceDiscoveryInstanceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSServiceDiscoveryInstance_http(t *testing.T) {
	resourceName := "aws_service_discovery_instance.instance"
	serviceResourceName := "aws_service_discovery_service.sd_register_instance"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAwsServiceDiscoveryInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSServiceDiscoveryInstanceBaseConfig(rName),
					testAccAWSServiceDiscoveryInstanceHttpNamespaceConfig(rName),
					testAccAWSServiceDiscoveryInstanceConfig(rName, "AWS_EC2_INSTANCE_ID = aws_instance.test_instance.id"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryInstanceExists(resourceName, serviceResourceName),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", rName),
					resource.TestCheckResourceAttrSet(resourceName, "attributes.AWS_EC2_INSTANCE_ID"),
				),
			},
			{
				Config: composeConfig(
					testAccAWSServiceDiscoveryInstanceBaseConfig(rName),
					testAccAWSServiceDiscoveryInstanceHttpNamespaceConfig(rName),
					testAccAWSServiceDiscoveryInstanceConfig(rName, "AWS_INSTANCE_IPV4 = \"172.18.0.12\""),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryInstanceExists(resourceName, serviceResourceName),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", rName),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV4", "172.18.0.12"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSServiceDiscoveryInstanceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAWSServiceDiscoveryInstanceBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "sd_register_instance" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %q
  }
}`, rName)
}

func testAccAWSServiceDiscoveryInstancePrivateNamespaceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_private_dns_namespace" "sd_register_instance" {
  name        = "sd-register-instance.local"
  description = "SD Register Instance - %[1]s"
  vpc         = aws_vpc.sd_register_instance.id
}

resource "aws_service_discovery_service" "sd_register_instance" {
  name = "%[1]s-service"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.sd_register_instance.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = 1
  }
} 
`, rName)
}

func testAccAWSServiceDiscoveryInstancePublicNamespaceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_public_dns_namespace" "sd_register_instance" {
  name        = "sd-register-instance.local"
}

resource "aws_service_discovery_service" "sd_register_instance" {
  name = "%[1]s-service"

  dns_config {
    namespace_id = aws_service_discovery_public_dns_namespace.sd_register_instance.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = 1
  }
} 
`, rName)
}

func testAccAWSServiceDiscoveryInstanceHttpNamespaceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_instance" "test_instance" {
  instance_type = "t3.micro"
  ami = data.aws_ami.amzn_linux_2018_03.id
  tags = {
    Name = "test instance"
  }
}

data "aws_ami" "amzn_linux_2018_03" {
  owners = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-2018.03.0.20180811-x86_64-gp2"]
  }
}

resource "aws_service_discovery_http_namespace" "sd_register_instance" {
  name        = "sd-register-instance.local"
}

resource "aws_service_discovery_service" "sd_register_instance" {
  name = "%[1]s-service"
  namespace_id = aws_service_discovery_http_namespace.sd_register_instance.id
} 
`, rName)
}

func testAccAWSServiceDiscoveryInstanceConfig(instanceID string, attributes string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_instance" "instance" {
  service_id = aws_service_discovery_service.sd_register_instance.id
  instance_id = "%s"
  attributes = {
    %s
  }
}
`, instanceID, attributes)
}

func testAccCheckAwsServiceDiscoveryInstanceExists(name, rServiceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rsInstance, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		rsService, ok := s.RootModule().Resources[rServiceName]
		if !ok {
			return fmt.Errorf("Not found: %s", rServiceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).sdconn

		input := &servicediscovery.GetInstanceInput{
			InstanceId: aws.String(rsInstance.Primary.ID),
			ServiceId:  aws.String(rsService.Primary.ID),
		}

		_, err := conn.GetInstance(input)
		return err
	}
}

func testAccAWSServiceDiscoveryInstanceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["service_id"], rs.Primary.Attributes["instance_id"]), nil
	}
}

func testAccCheckAwsServiceDiscoveryInstanceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sdconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_service_discovery_instance" {
			continue
		}

		input := &servicediscovery.GetInstanceInput{
			InstanceId: aws.String(rs.Primary.ID),
			ServiceId:  aws.String(rs.Primary.Attributes["service_id"]),
		}

		_, err := conn.GetInstance(input)
		if err != nil {
			if isAWSErr(err, servicediscovery.ErrCodeInstanceNotFound, "") {
				return nil
			}
			return err
		}
	}
	return nil
}