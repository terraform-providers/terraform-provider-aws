package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/networkfirewall/finder"
)

func init() {
	resource.AddTestSweepers("aws_networkfirewall_firewall", &resource.Sweeper{
		Name:         "aws_networkfirewall_firewall",
		F:            testSweepNetworkFirewallFirewalls,
		Dependencies: []string{"aws_networkfirewall_logging_configuration"},
	})
}

func testSweepNetworkFirewallFirewalls(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).networkfirewallconn
	ctx := context.TODO()
	input := &networkfirewall.ListFirewallsInput{MaxResults: aws.Int64(100)}
	var sweeperErrs *multierror.Error

	for {
		resp, err := conn.ListFirewallsWithContext(ctx, input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping NetworkFirewall Firewall sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving NetworkFirewall firewalls: %s", err)
		}

		for _, f := range resp.Firewalls {
			if f == nil {
				continue
			}

			arn := aws.StringValue(f.FirewallArn)

			log.Printf("[INFO] Deleting NetworkFirewall Firewall: %s", arn)

			r := resourceAwsNetworkFirewallFirewall()
			d := r.Data(nil)
			d.SetId(arn)
			diags := r.DeleteContext(ctx, d, client)
			for i := range diags {
				if diags[i].Severity == diag.Error {
					log.Printf("[ERROR] %s", diags[i].Summary)
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf(diags[i].Summary))
					continue
				}
			}
		}

		if aws.StringValue(resp.NextToken) == "" {
			break
		}
		input.NextToken = resp.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsNetworkFirewallFirewall_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall.test"
	policyResourceName := "aws_networkfirewall_firewall_policy.test"
	subnetResourceName := "aws_subnet.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallFirewall_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "network-firewall", fmt.Sprintf("firewall/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "delete_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_policy_arn", policyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.0.sync_states.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.availability_zone", subnetResourceName, "availability_zone"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "firewall_status.0.sync_states.*", map[string]*regexp.Regexp{
						"attachment.0.endpoint_id": regexp.MustCompile(`vpce-`),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
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

func TestAccAwsNetworkFirewallFirewall_description(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallFirewall_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			{
				Config: testAccNetworkFirewallFirewall_updateDescription(rName, "updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "updated"),
				),
			},
			{
				Config: testAccNetworkFirewallFirewall_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
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

func TestAccAwsNetworkFirewallFirewall_deleteProtection(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallDestroy,
		Steps: []resource.TestStep{

			{
				Config: testAccNetworkFirewallFirewall_deleteProtection(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "delete_protection", "false"),
				),
			},
			{
				Config: testAccNetworkFirewallFirewall_deleteProtection(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "delete_protection", "true"),
				),
			},
			{
				Config: testAccNetworkFirewallFirewall_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "delete_protection", "false"),
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

func TestAccAwsNetworkFirewallFirewall_subnetMappings_updateSubnet(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall.test"
	subnetResourceName := "aws_subnet.test"
	updateSubnetResourceName := "aws_subnet.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallDestroy,
		Steps: []resource.TestStep{

			{
				Config: testAccNetworkFirewallFirewall_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", subnetResourceName, "id"),
				),
			},
			{
				Config: testAccNetworkFirewallFirewall_updateSubnet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.0.sync_states.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.availability_zone", updateSubnetResourceName, "availability_zone"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "firewall_status.0.sync_states.*", map[string]*regexp.Regexp{
						"attachment.0.endpoint_id": regexp.MustCompile(`vpce-`),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", updateSubnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", updateSubnetResourceName, "id"),
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

func TestAccAwsNetworkFirewallFirewall_subnetMappings_updateMultipleSubnets(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall.test"
	subnetResourceName := "aws_subnet.test"
	updateSubnetResourceName := "aws_subnet.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallFirewall_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", subnetResourceName, "id"),
				),
			},
			{
				Config: testAccNetworkFirewallFirewall_updateMultipleSubnets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.0.sync_states.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.availability_zone", subnetResourceName, "availability_zone"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", subnetResourceName, "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.availability_zone", updateSubnetResourceName, "availability_zone"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", updateSubnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", subnetResourceName, "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", updateSubnetResourceName, "id"),
				),
			},
			{
				Config: testAccNetworkFirewallFirewall_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_status.0.sync_states.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.availability_zone", subnetResourceName, "availability_zone"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "firewall_status.0.sync_states.*", map[string]*regexp.Regexp{
						"attachment.0.endpoint_id": regexp.MustCompile(`vpce-`),
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_status.0.sync_states.*.attachment.0.subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "subnet_mapping.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_mapping.*.subnet_id", subnetResourceName, "id"),
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

func TestAccAwsNetworkFirewallFirewall_tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallFirewall_oneTag(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccNetworkFirewallFirewall_twoTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Description", "updated"),
				),
			},
			{
				Config: testAccNetworkFirewallFirewall_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAwsNetworkFirewallFirewall_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_firewall.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallFirewallDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallFirewall_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallFirewallExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsNetworkFirewallFirewall(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsNetworkFirewallFirewallDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_networkfirewall_firewall" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).networkfirewallconn
		output, err := finder.Firewall(context.Background(), conn, rs.Primary.ID)
		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}
		if output != nil {
			return fmt.Errorf("NetworkFirewall Firewall still exists: %s", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsNetworkFirewallFirewallExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No NetworkFirewall Firewall ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).networkfirewallconn
		output, err := finder.Firewall(context.Background(), conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("NetworkFirewall Firewall (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPreCheckAwsNetworkFirewall(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).networkfirewallconn

	input := &networkfirewall.ListFirewallsInput{}

	_, err := conn.ListFirewalls(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccNetworkFirewallFirewallDependenciesConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}
`, rName)
}

func testAccNetworkFirewallFirewall_basic(rName string) string {
	return composeConfig(
		testAccNetworkFirewallFirewallDependenciesConfig(rName),
		fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}
`, rName))
}

func testAccNetworkFirewallFirewall_deleteProtection(rName string, deleteProtection bool) string {
	return composeConfig(
		testAccNetworkFirewallFirewallDependenciesConfig(rName),
		fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  delete_protection   = %t
  name                = %q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}
`, deleteProtection, rName))
}

func testAccNetworkFirewallFirewall_oneTag(rName string) string {
	return composeConfig(
		testAccNetworkFirewallFirewallDependenciesConfig(rName),
		fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id
  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccNetworkFirewallFirewall_twoTags(rName string) string {
	return composeConfig(
		testAccNetworkFirewallFirewallDependenciesConfig(rName),
		fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id
  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
  tags = {
    Name        = %[1]q
    Description = "updated"
  }
}
`, rName))
}

func testAccNetworkFirewallFirewall_updateDescription(rName, description string) string {
	return composeConfig(
		testAccNetworkFirewallFirewallDependenciesConfig(rName),
		fmt.Sprintf(`
resource "aws_networkfirewall_firewall" "test" {
  name                = %q
  description         = %q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id
  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}
`, rName, description))
}

func testAccNetworkFirewallFirewall_updateSubnet(rName string) string {
	return composeConfig(
		testAccNetworkFirewallFirewallDependenciesConfig(rName),
		fmt.Sprintf(`
resource "aws_subnet" "example" {
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 1)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.example.id
  }
}
`, rName))
}

func testAccNetworkFirewallFirewall_updateMultipleSubnets(rName string) string {
	return composeConfig(
		testAccNetworkFirewallFirewallDependenciesConfig(rName),
		fmt.Sprintf(`
resource "aws_subnet" "example" {
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 1)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }

  subnet_mapping {
    subnet_id = aws_subnet.example.id
  }
}
`, rName))
}
