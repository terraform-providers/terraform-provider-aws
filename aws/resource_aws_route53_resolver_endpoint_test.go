package aws

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_route53_resolver_endpoint", &resource.Sweeper{
		Name: "aws_route53_resolver_endpoint",
		F:    testSweepRoute53ResolverEndpoints,
		Dependencies: []string{
			"aws_route53_resolver_rule",
		},
	})
}

func testSweepRoute53ResolverEndpoints(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).route53resolverconn

	var errors error
	err = conn.ListResolverEndpointsPages(&route53resolver.ListResolverEndpointsInput{}, func(page *route53resolver.ListResolverEndpointsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, resolverEndpoint := range page.ResolverEndpoints {
			id := aws.StringValue(resolverEndpoint.Id)

			log.Printf("[INFO] Deleting Route53 Resolver endpoint: %s", id)
			_, err := conn.DeleteResolverEndpoint(&route53resolver.DeleteResolverEndpointInput{
				ResolverEndpointId: aws.String(id),
			})
			if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
				continue
			}
			if err != nil {
				errors = multierror.Append(errors, fmt.Errorf("error deleting Route53 Resolver endpoint (%s): %w", id, err))
				continue
			}

			err = route53ResolverEndpointWaitUntilTargetState(conn, id, 10*time.Minute,
				[]string{route53resolver.ResolverEndpointStatusDeleting},
				[]string{route53ResolverEndpointStatusDeleted})
			if err != nil {
				errors = multierror.Append(errors, err)
				continue
			}
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Resolver endpoint sweep for %s: %s", region, err)
			return nil
		}
		errors = multierror.Append(errors, fmt.Errorf("error retrievingRoute53 Resolver endpoints: %w", err))
	}

	return errors
}

func TestAccAWSRoute53ResolverEndpoint_basicInbound(t *testing.T) {
	var ep route53resolver.ResolverEndpoint
	resourceName := "aws_route53_resolver_endpoint.foo"
	rInt := acctest.RandInt()
	name := fmt.Sprintf("terraform-testacc-r53-resolver-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		ErrorCheck:   testAccErrorCheckSkipRoute53(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverEndpointConfig_initial(rInt, "INBOUND", name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverEndpointExists(resourceName, &ep),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "ip_address.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
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

func TestAccAWSRoute53ResolverEndpoint_updateOutbound(t *testing.T) {
	var ep route53resolver.ResolverEndpoint
	resourceName := "aws_route53_resolver_endpoint.foo"
	rInt := acctest.RandInt()
	initialName := fmt.Sprintf("terraform-testacc-r53-resolver-%d", rInt)
	updatedName := fmt.Sprintf("terraform-testacc-r53-rupdated-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSRoute53Resolver(t) },
		ErrorCheck:   testAccErrorCheckSkipRoute53(t),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ResolverEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverEndpointConfig_initial(rInt, "OUTBOUND", initialName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverEndpointExists(resourceName, &ep),
					resource.TestCheckResourceAttr(resourceName, "name", initialName),
					resource.TestCheckResourceAttr(resourceName, "ip_address.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
				),
			},
			{
				Config: testAccRoute53ResolverEndpointConfig_updated(rInt, "OUTBOUND", updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverEndpointExists(resourceName, &ep),
					resource.TestCheckResourceAttr(resourceName, "name", updatedName),
					resource.TestCheckResourceAttr(resourceName, "ip_address.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "changed"),
				),
			},
		},
	})
}

func testAccCheckRoute53ResolverEndpointDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).route53resolverconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_endpoint" {
			continue
		}

		// Try to find the resource
		_, err := conn.GetResolverEndpoint(&route53resolver.GetResolverEndpointInput{
			ResolverEndpointId: aws.String(rs.Primary.ID),
		})
		// Verify the error is what we want
		if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("Route 53 Resolver endpoint still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckRoute53ResolverEndpointExists(n string, ep *route53resolver.ResolverEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route 53 Resolver endpoint ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).route53resolverconn
		resp, err := conn.GetResolverEndpoint(&route53resolver.GetResolverEndpointInput{
			ResolverEndpointId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*ep = *resp.ResolverEndpoint

		return nil
	}
}

func testAccPreCheckAWSRoute53Resolver(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).route53resolverconn

	input := &route53resolver.ListResolverEndpointsInput{}

	_, err := conn.ListResolverEndpoints(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccRoute53ResolverEndpointConfig_base(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-r53-resolver-vpc-%[1]d"
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "sn1" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = cidrsubnet(aws_vpc.foo.cidr_block, 2, 0)
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-r53-resolver-sn1-%[1]d"
  }
}

resource "aws_subnet" "sn2" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = cidrsubnet(aws_vpc.foo.cidr_block, 2, 1)
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-r53-resolver-sn2-%[1]d"
  }
}

resource "aws_subnet" "sn3" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = cidrsubnet(aws_vpc.foo.cidr_block, 2, 2)
  availability_zone = data.aws_availability_zones.available.names[2]

  tags = {
    Name = "tf-acc-r53-resolver-sn3-%[1]d"
  }
}

resource "aws_security_group" "sg1" {
  vpc_id = aws_vpc.foo.id
  name   = "tf-acc-r53-resolver-sg1-%[1]d"

  tags = {
    Name = "tf-acc-r53-resolver-sg1-%[1]d"
  }
}

resource "aws_security_group" "sg2" {
  vpc_id = aws_vpc.foo.id
  name   = "tf-acc-r53-resolver-sg2-%[1]d"

  tags = {
    Name = "tf-acc-r53-resolver-sg2-%[1]d"
  }
}
`, rInt)
}

func testAccRoute53ResolverEndpointConfig_initial(rInt int, direction, name string) string {
	return fmt.Sprintf(`
%s

resource "aws_route53_resolver_endpoint" "foo" {
  direction = "%s"
  name      = "%s"

  security_group_ids = [
    aws_security_group.sg1.id,
    aws_security_group.sg2.id,
  ]

  ip_address {
    subnet_id = aws_subnet.sn1.id
  }

  ip_address {
    subnet_id = aws_subnet.sn2.id
    ip        = cidrhost(aws_subnet.sn2.cidr_block, 8)
  }

  tags = {
    Environment = "production"
    Usage       = "original"
  }
}
`, testAccRoute53ResolverEndpointConfig_base(rInt), direction, name)
}

func testAccRoute53ResolverEndpointConfig_updated(rInt int, direction, name string) string {
	return fmt.Sprintf(`
%s

resource "aws_route53_resolver_endpoint" "foo" {
  direction = "%s"
  name      = "%s"

  security_group_ids = [
    aws_security_group.sg1.id,
    aws_security_group.sg2.id,
  ]

  ip_address {
    subnet_id = aws_subnet.sn1.id
  }

  ip_address {
    subnet_id = aws_subnet.sn3.id
  }

  tags = {
    Usage = "changed"
  }
}
`, testAccRoute53ResolverEndpointConfig_base(rInt), direction, name)
}
