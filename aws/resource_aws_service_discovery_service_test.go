package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_service_discovery_service", &resource.Sweeper{
		Name: "aws_service_discovery_service",
		F:    testSweepServiceDiscoveryServices,
	})
}

func testSweepServiceDiscoveryServices(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sdconn
	var sweeperErrs *multierror.Error

	input := &servicediscovery.ListServicesInput{}

	err = conn.ListServicesPages(input, func(page *servicediscovery.ListServicesOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, service := range page.Services {
			if service == nil {
				continue
			}

			serviceID := aws.StringValue(service.Id)
			input := &servicediscovery.DeleteServiceInput{
				Id: service.Id,
			}

			if aws.Int64Value(service.InstanceCount) > 0 {
				input := &servicediscovery.ListInstancesInput{}

				err := conn.ListInstancesPages(input, func(page *servicediscovery.ListInstancesOutput, isLast bool) bool {
					if page == nil {
						return !isLast
					}

					for _, instance := range page.Instances {
						if instance == nil {
							continue
						}

						instanceID := aws.StringValue(instance.Id)
						input := &servicediscovery.DeregisterInstanceInput{
							InstanceId: instance.Id,
							ServiceId:  service.Id,
						}

						log.Printf("[INFO] Deregistering Service Discovery Service (%s) Instance: %s", serviceID, instanceID)
						_, err := conn.DeregisterInstance(input)

						if err != nil {
							sweeperErr := fmt.Errorf("error deregistering Service Discovery Service (%s) Instance (%s): %w", serviceID, instanceID, err)
							log.Printf("[ERROR] %s", sweeperErr)
							sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
							continue
						}
					}

					return !isLast
				})

				if err != nil {
					sweeperErr := fmt.Errorf("error listing Service Discovery Service (%s) Instances: %w", serviceID, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}
			}

			log.Printf("[INFO] Deleting Service Discovery Service: %s", serviceID)
			_, err := conn.DeleteService(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Service Discovery Service (%s): %w", serviceID, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !isLast
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Service Discovery Services sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Service Discovery Services: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSServiceDiscoveryService_private(t *testing.T) {
	resourceName := "aws_service_discovery_service.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(servicediscovery.EndpointsID, t)
			testAccPreCheckAWSServiceDiscovery(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryServiceConfig_private(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "health_check_custom_config.0.failure_threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.dns_records.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.dns_records.0.type", "A"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.dns_records.0.ttl", "5"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.routing_policy", "MULTIVALUE"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "servicediscovery", regexp.MustCompile(`service/.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceDiscoveryServiceConfig_private_update(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.dns_records.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.dns_records.0.type", "A"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.dns_records.0.ttl", "10"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.dns_records.1.type", "AAAA"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.dns_records.1.ttl", "5"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "servicediscovery", regexp.MustCompile(`service/.+`)),
				),
			},
		},
	})
}

func TestAccAWSServiceDiscoveryService_public(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_service_discovery_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(servicediscovery.EndpointsID, t)
			testAccPreCheckAWSServiceDiscovery(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryServiceConfig_public(rName, 5, "/path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "health_check_config.0.type", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check_config.0.failure_threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "health_check_config.0.resource_path", "/path"),
					resource.TestCheckResourceAttr(resourceName, "dns_config.0.routing_policy", "WEIGHTED"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "servicediscovery", regexp.MustCompile(`service/.+`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceDiscoveryServiceConfig_public(rName, 3, "/updated-path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "health_check_config.0.type", "HTTP"),
					resource.TestCheckResourceAttr(resourceName, "health_check_config.0.failure_threshold", "3"),
					resource.TestCheckResourceAttr(resourceName, "health_check_config.0.resource_path", "/updated-path"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "servicediscovery", regexp.MustCompile(`service/.+`)),
				),
			},
			{
				Config: testAccServiceDiscoveryServiceConfig_public_update_noHealthCheckConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "health_check_config.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "servicediscovery", regexp.MustCompile(`service/.+`)),
				),
			},
		},
	})
}

func TestAccAWSServiceDiscoveryService_http(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	resourceName := "aws_service_discovery_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(servicediscovery.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryServiceConfig_http(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryServiceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "namespace_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "servicediscovery", regexp.MustCompile(`service/.+`)),
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

func TestAccAWSServiceDiscoveryService_disappears(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	resourceName := "aws_service_discovery_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(servicediscovery.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryServiceConfig_http(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryServiceExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceDiscoveryService(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceDiscoveryService_Tags(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(8, acctest.CharSetAlpha))
	resourceName := "aws_service_discovery_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(servicediscovery.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryServiceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryServiceConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceDiscoveryServiceConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccServiceDiscoveryServiceConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryServiceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
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
    Name = %[1]q
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name = "%[1]s.tf"
  vpc  = aws_vpc.test.id
}

resource "aws_service_discovery_service" "test" {
  name = %[1]q

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.test.id

    dns_records {
      ttl  = 5
      type = "A"
    }
  }

  health_check_custom_config {
    failure_threshold = %[2]d
  }
}
`, rName, th)
}

func testAccServiceDiscoveryServiceConfig_private_update(rName string, th int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name = "%[1]s.tf"
  vpc  = aws_vpc.test.id
}

resource "aws_service_discovery_service" "test" {
  name = %[1]q

  description = "test"

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.test.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    dns_records {
      ttl  = 5
      type = "AAAA"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = %[2]d
  }
}
`, rName, th)
}

func testAccServiceDiscoveryServiceConfig_public(rName string, th int, path string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_public_dns_namespace" "test" {
  name = "%[1]s.tf"
}

resource "aws_service_discovery_service" "test" {
  name = %[1]q

  description = "test"

  dns_config {
    namespace_id = aws_service_discovery_public_dns_namespace.test.id

    dns_records {
      ttl  = 5
      type = "A"
    }

    routing_policy = "WEIGHTED"
  }

  health_check_config {
    failure_threshold = %[2]d
    resource_path     = %[3]q
    type              = "HTTP"
  }
}
`, rName, th, path)
}

func testAccServiceDiscoveryServiceConfig_public_update_noHealthCheckConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_public_dns_namespace" "test" {
  name = "%[1]s.tf"
}

resource "aws_service_discovery_service" "test" {
  name = %[1]q

  dns_config {
    namespace_id = aws_service_discovery_public_dns_namespace.test.id

    dns_records {
      ttl  = 5
      type = "A"
    }

    routing_policy = "WEIGHTED"
  }
}
`, rName)
}

func testAccServiceDiscoveryServiceConfig_http(rName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_http_namespace" "test" {
  name = %[1]q
}

resource "aws_service_discovery_service" "test" {
  name         = %[1]q
  namespace_id = aws_service_discovery_http_namespace.test.id
}
`, rName)
}

func testAccServiceDiscoveryServiceConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_http_namespace" "test" {
  name = %[1]q
}

resource "aws_service_discovery_service" "test" {
  name         = %[1]q
  namespace_id = aws_service_discovery_http_namespace.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccServiceDiscoveryServiceConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_http_namespace" "test" {
  name = %[1]q
}

resource "aws_service_discovery_service" "test" {
  name         = %[1]q
  namespace_id = aws_service_discovery_http_namespace.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
