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
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicediscovery/waiter"
)

func init() {
	resource.AddTestSweepers("aws_service_discovery_private_dns_namespace", &resource.Sweeper{
		Name: "aws_service_discovery_private_dns_namespace",
		F:    testSweepServiceDiscoveryPrivateDnsNamespaces,
		Dependencies: []string{
			"aws_service_discovery_service",
		},
	})
}

func testSweepServiceDiscoveryPrivateDnsNamespaces(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).sdconn
	var sweeperErrs *multierror.Error

	input := &servicediscovery.ListNamespacesInput{
		Filters: []*servicediscovery.NamespaceFilter{
			{
				Condition: aws.String(servicediscovery.FilterConditionEq),
				Name:      aws.String(servicediscovery.NamespaceFilterNameType),
				Values:    aws.StringSlice([]string{servicediscovery.NamespaceTypeDnsPrivate}),
			},
		},
	}

	err = conn.ListNamespacesPages(input, func(page *servicediscovery.ListNamespacesOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, namespace := range page.Namespaces {
			if namespace == nil {
				continue
			}

			id := aws.StringValue(namespace.Id)
			input := &servicediscovery.DeleteNamespaceInput{
				Id: namespace.Id,
			}

			log.Printf("[INFO] Deleting Service Discovery Private DNS Namespace: %s", id)
			output, err := conn.DeleteNamespace(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Service Discovery Private DNS Namespace (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			if output != nil && output.OperationId != nil {
				if _, err := waiter.OperationSuccess(conn, aws.StringValue(output.OperationId)); err != nil {
					sweeperErr := fmt.Errorf("error waiting for Service Discovery Private DNS Namespace (%s) deletion: %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}
			}
		}

		return !isLast
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Service Discovery Private DNS Namespaces sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Service Discovery Private DNS Namespaces: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSServiceDiscoveryPrivateDnsNamespace_basic(t *testing.T) {
	resourceName := "aws_service_discovery_private_dns_namespace.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSServiceDiscovery(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryPrivateDnsNamespaceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "servicediscovery", regexp.MustCompile(`namespace/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "hosted_zone"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccServiceDiscoveryPrivateDnsNamespaceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSServiceDiscoveryPrivateDnsNamespace_disappears(t *testing.T) {
	resourceName := "aws_service_discovery_private_dns_namespace.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSServiceDiscovery(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryPrivateDnsNamespaceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServiceDiscoveryPrivateDnsNamespace(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSServiceDiscoveryPrivateDnsNamespace_Description(t *testing.T) {
	resourceName := "aws_service_discovery_private_dns_namespace.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSServiceDiscovery(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryPrivateDnsNamespaceConfigDescription(rName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
				),
			},
		},
	})
}

// This acceptance test ensures we properly send back error messaging. References:
//  * https://github.com/hashicorp/terraform-provider-aws/issues/2830
//  * https://github.com/hashicorp/terraform-provider-aws/issues/5532
func TestAccAWSServiceDiscoveryPrivateDnsNamespace_error_Overlap(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSServiceDiscovery(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceDiscoveryPrivateDnsNamespaceConfigOverlapping(rName),
				ExpectError: regexp.MustCompile(`ConflictingDomainExists`),
			},
		},
	})
}

func TestAccAWSServiceDiscoveryPrivateDnsNamespace_Tags(t *testing.T) {
	resourceName := "aws_service_discovery_private_dns_namespace.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSServiceDiscovery(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryPrivateDnsNamespaceConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccServiceDiscoveryPrivateDnsNamespaceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccServiceDiscoveryPrivateDnsNamespaceConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccServiceDiscoveryPrivateDnsNamespaceConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sdconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_service_discovery_private_dns_namespace" {
			continue
		}

		input := &servicediscovery.GetNamespaceInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetNamespace(input)
		if err != nil {
			if isAWSErr(err, servicediscovery.ErrCodeNamespaceNotFound, "") {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckAwsServiceDiscoveryPrivateDnsNamespaceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccPreCheckAWSServiceDiscovery(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).sdconn

	input := &servicediscovery.ListNamespacesInput{}

	_, err := conn.ListNamespaces(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccServiceDiscoveryPrivateDnsNamespaceConfig(rName string) string {
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
`, rName)
}

func testAccServiceDiscoveryPrivateDnsNamespaceConfigDescription(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[2]q
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  description = %[1]q
  name        = "%[2]s.tf"
  vpc         = aws_vpc.test.id
}
`, description, rName)
}

func testAccServiceDiscoveryPrivateDnsNamespaceConfigOverlapping(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_service_discovery_private_dns_namespace" "top" {
  name = "%[1]s.tf"
  vpc  = aws_vpc.test.id
}

# Ensure ordering after first namespace
resource "aws_service_discovery_private_dns_namespace" "subdomain" {
  name = aws_service_discovery_private_dns_namespace.top.name
  vpc  = aws_service_discovery_private_dns_namespace.top.vpc
}
`, rName)
}

func testAccServiceDiscoveryPrivateDnsNamespaceConfigTags1(rName, tagKey1, tagValue1 string) string {
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

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccServiceDiscoveryPrivateDnsNamespaceConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccServiceDiscoveryPrivateDnsNamespaceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return fmt.Sprintf("%s:%s", rs.Primary.ID, rs.Primary.Attributes["vpc"]), nil
	}
}
