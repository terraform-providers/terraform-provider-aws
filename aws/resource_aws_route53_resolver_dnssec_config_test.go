package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/route53resolver/finder"
)

func init() {
	resource.AddTestSweepers("aws_route53_resolver_dnssec_config", &resource.Sweeper{
		Name: "aws_route53_resolver_dnssec_config",
		F:    testSweepRoute53ResolverDnssecConfig,
	})
}

func testSweepRoute53ResolverDnssecConfig(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).route53resolverconn

	var sweeperErrs *multierror.Error
	err = conn.ListResolverDnssecConfigsPages(&route53resolver.ListResolverDnssecConfigsInput{}, func(page *route53resolver.ListResolverDnssecConfigsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, resolverDnssecConfig := range page.ResolverDnssecConfigs {
			if resolverDnssecConfig == nil {
				continue
			}

			id := aws.StringValue(resolverDnssecConfig.Id)
			resourceId := aws.StringValue(resolverDnssecConfig.ResourceId)

			log.Printf("[INFO] Deleting Route 53 Resolver Dnssec config: %s", id)

			r := resourceAwsRoute53ResolverDnssecConfig()
			d := r.Data(nil)
			d.SetId(aws.StringValue(resolverDnssecConfig.Id))
			d.Set("resource_id", resourceId)

			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Route 53 Resolver Resolver Dnssec config (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !isLast
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Route 53 Resolver Resolver Dnssec config sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Route 53 Resolver Resolver Dnssec config: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSRoute53ResolverDnssecConfig_basic(t *testing.T) {
	resourceName := "aws_route53_resolver_dnssec_config.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipRoute53(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverDnssecConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverDnssecConfigConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverDnssecConfigExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "route53resolver", regexp.MustCompile(`resolver-dnssec-config/.+$`)),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "owner_id"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_id"),
					resource.TestCheckResourceAttr(resourceName, "validation_status", "ENABLED"),
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

func TestAccAWSRoute53ResolverDnssecConfig_disappear(t *testing.T) {
	resourceName := "aws_route53_resolver_dnssec_config.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheckSkipRoute53(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverDnssecConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverDnssecConfigConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverDnssecConfigExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRoute53ResolverDnssecConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRoute53ResolverDnssecConfigDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).route53resolverconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_dnssec_config" {
			continue
		}

		config, err := finder.ResolverDnssecConfigByID(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if config == nil {
			continue
		}

		return fmt.Errorf("Route 53 Resolver Dnssec config still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckRoute53ResolverDnssecConfigExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route 53 Resolver Dnssec config ID is set")
		}

		id := rs.Primary.ID
		conn := testAccProvider.Meta().(*AWSClient).route53resolverconn

		config, err := finder.ResolverDnssecConfigByID(conn, id)

		if err != nil {
			return err
		}

		if config == nil {
			return fmt.Errorf("Route 53 Resolver Dnssec config (%s) not found", id)
		}

		if aws.StringValue(config.ValidationStatus) != route53resolver.ResolverDNSSECValidationStatusEnabled {
			return fmt.Errorf("Route 53 Resolver Dnssec config (%s) is not enabled", aws.StringValue(config.Id))
		}

		return nil
	}
}

func testAccRoute53ResolverDnssecConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %q
  }
}
`, rName)
}

func testAccRoute53ResolverDnssecConfigConfigBasic(rName string) string {
	return testAccRoute53ResolverDnssecConfigBase(rName) + `
resource "aws_route53_resolver_dnssec_config" "test" {
  resource_id = aws_vpc.test.id
}
`
}
