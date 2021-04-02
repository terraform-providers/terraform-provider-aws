package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloud9/finder"
)

func init() {
	resource.AddTestSweepers("aws_cloud9_environment_ec2", &resource.Sweeper{
		Name: "aws_cloud9_environment_ec2",
		F:    testSweepCloud9EnvironmentEC2s,
	})
}

func testSweepCloud9EnvironmentEC2s(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).cloud9conn
	var sweeperErrs *multierror.Error

	input := &cloud9.ListEnvironmentsInput{}
	err = conn.ListEnvironmentsPages(input, func(page *cloud9.ListEnvironmentsOutput, lastPage bool) bool {
		if len(page.EnvironmentIds) == 0 {
			log.Printf("[INFO] No Cloud9 Environment EC2s to sweep")
			return false
		}
		for _, envID := range page.EnvironmentIds {
			id := aws.StringValue(envID)

			log.Printf("[INFO] Deleting Cloud9 Environment EC2: %s", id)
			r := resourceAwsCloud9EnvironmentEc2()
			d := r.Data(nil)
			d.SetId(id)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}
		return !lastPage
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Cloud9 Environment EC2s sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Cloud9 Environment EC2s: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSCloud9EnvironmentEc2_basic(t *testing.T) {
	var conf cloud9.Environment

	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := acctest.RandomWithPrefix("tf-acc-test-updated")
	resourceName := "aws_cloud9_environment_ec2.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloud9.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloud9.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloud9EnvironmentEc2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloud9EnvironmentEc2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloud9EnvironmentEc2Exists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "cloud9", regexp.MustCompile(`environment:.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "owner_arn", "data.aws_caller_identity.current", "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"instance_type", "subnet_id"},
			},
			{
				Config: testAccAWSCloud9EnvironmentEc2Config(rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloud9EnvironmentEc2Exists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "cloud9", regexp.MustCompile(`environment:.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "owner_arn", "data.aws_caller_identity.current", "arn"),
				),
			},
		},
	})
}

func TestAccAWSCloud9EnvironmentEc2_allFields(t *testing.T) {
	var conf cloud9.Environment

	rName := acctest.RandomWithPrefix("tf-acc-test")
	description := acctest.RandomWithPrefix("Tf Acc Test")
	uDescription := acctest.RandomWithPrefix("Tf Acc Test Updated")
	resourceName := "aws_cloud9_environment_ec2.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloud9.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloud9.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloud9EnvironmentEc2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloud9EnvironmentEc2AllFieldsConfig(rName, description, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloud9EnvironmentEc2Exists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "cloud9", regexp.MustCompile(`environment:.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "owner_arn", "aws_cloud9_user.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", "ec2"),
					resource.TestCheckResourceAttr(resourceName, "description", description),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"instance_type", "automatic_stop_time_minutes", "subnet_id"},
			},
			{
				Config: testAccAWSCloud9EnvironmentEc2AllFieldsConfig(rName, uDescription, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloud9EnvironmentEc2Exists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "cloud9", regexp.MustCompile(`environment:.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "owner_arn", "aws_cloud9_user.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", "ec2"),
					resource.TestCheckResourceAttr(resourceName, "description", uDescription),
				),
			},
		},
	})
}

func TestAccAWSCloud9EnvironmentEc2_tags(t *testing.T) {
	var conf cloud9.Environment

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloud9_environment_ec2.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloud9.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloud9.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloud9EnvironmentEc2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloud9EnvironmentEc2ConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloud9EnvironmentEc2Exists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"instance_type", "subnet_id"},
			},
			{
				Config: testAccAWSCloud9EnvironmentEc2ConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloud9EnvironmentEc2Exists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSCloud9EnvironmentEc2ConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloud9EnvironmentEc2Exists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSCloud9EnvironmentEc2_disappears(t *testing.T) {
	var conf cloud9.Environment

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloud9_environment_ec2.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloud9.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloud9.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloud9EnvironmentEc2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloud9EnvironmentEc2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloud9EnvironmentEc2Exists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloud9EnvironmentEc2(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSCloud9EnvironmentEc2Exists(n string, res *cloud9.Environment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Cloud9 Environment EC2 ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cloud9conn

		out, err := finder.EnvironmentByID(conn, rs.Primary.ID)
		if err != nil {
			if isAWSErr(err, cloud9.ErrCodeNotFoundException, "") {
				return fmt.Errorf("Cloud9 Environment EC2 (%q) not found", rs.Primary.ID)
			}
			return err
		}
		if out == nil {
			return fmt.Errorf("Cloud9 Environment EC2 (%q) not found", rs.Primary.ID)
		}

		*res = *out

		return nil
	}
}

func testAccCheckAWSCloud9EnvironmentEc2Destroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloud9conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloud9_environment_ec2" {
			continue
		}

		out, err := finder.EnvironmentByID(conn, rs.Primary.ID)
		if err != nil {
			if isAWSErr(err, cloud9.ErrCodeNotFoundException, "") {
				return nil
			}
			// :'-(
			if isAWSErr(err, "AccessDeniedException", "is not authorized to access this resource") {
				return nil
			}
			return err
		}
		if out == nil {
			return nil
		}

		return fmt.Errorf("Cloud9 Environment EC2 %q still exists.", rs.Primary.ID)
	}
	return nil
}

func testAccAWSCloud9EnvironmentEc2ConfigBase() string {
	return `
data "aws_availability_zones" "available" {
  # t2.micro instance type is not available in these Availability Zones
  exclude_zone_ids = ["usw2-az4"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-cloud9-environment-ec2"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-cloud9-environment-ec2"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-cloud9-environment-ec2"
  }
}

resource "aws_route" "test" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
  route_table_id         = aws_vpc.test.main_route_table_id
}
`
}

func testAccAWSCloud9EnvironmentEc2Config(name string) string {
	return testAccAWSCloud9EnvironmentEc2ConfigBase() + fmt.Sprintf(`
resource "aws_cloud9_environment_ec2" "test" {
  depends_on = [aws_route.test]

  instance_type = "t2.micro"
  name          = %[1]q
  subnet_id     = aws_subnet.test.id
}

# By default, the Cloud9 environment EC2 is owned by the creator
data "aws_caller_identity" "current" {}
`, name)
}

func testAccAWSCloud9EnvironmentEc2AllFieldsConfig(name, description, userName string) string {
	return testAccAWSCloud9EnvironmentEc2ConfigBase() + fmt.Sprintf(`
resource "aws_cloud9_environment_ec2" "test" {
  depends_on = [aws_route.test]

  automatic_stop_time_minutes = 60
  description                 = %[2]q
  instance_type               = "t2.micro"
  name                        = %[1]q
  owner_arn                   = aws_cloud9_user.test.arn
  subnet_id                   = aws_subnet.test.id
}

resource "aws_cloud9_user" "test" {
  name = %[3]q
}
`, name, description, userName)
}

func testAccAWSCloud9EnvironmentEc2ConfigTags1(name, tagKey1, tagValue1 string) string {
	return testAccAWSCloud9EnvironmentEc2ConfigBase() + fmt.Sprintf(`
resource "aws_cloud9_environment_ec2" "test" {
  depends_on = [aws_route.test]

  instance_type = "t2.micro"
  name          = %[1]q
  subnet_id     = aws_subnet.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tagKey1, tagValue1)
}

func testAccAWSCloud9EnvironmentEc2ConfigTags2(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSCloud9EnvironmentEc2ConfigBase() + fmt.Sprintf(`
resource "aws_cloud9_environment_ec2" "test" {
  depends_on = [aws_route.test]

  instance_type = "t2.micro"
  name          = %[1]q
  subnet_id     = aws_subnet.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tagKey1, tagValue1, tagKey2, tagValue2)
}
