package aws

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

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
				ImportStateVerifyIgnore: []string{"instance_type", "subnet_id", "image_id"},
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
	rNameUpdated := acctest.RandomWithPrefix("tf-acc-test-updated")
	description := acctest.RandomWithPrefix("Tf Acc Test")
	uDescription := acctest.RandomWithPrefix("Tf Acc Test Updated")
	userName := acctest.RandomWithPrefix("tf_acc_cloud9_env")
	resourceName := "aws_cloud9_environment_ec2.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloud9.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloud9.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloud9EnvironmentEc2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloud9EnvironmentEc2AllFieldsConfig(rName, description, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloud9EnvironmentEc2Exists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "cloud9", regexp.MustCompile(`environment:.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "owner_arn", "aws_iam_user.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", "ec2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"instance_type", "automatic_stop_time_minutes", "subnet_id", "image_id"},
			},
			{
				Config: testAccAWSCloud9EnvironmentEc2AllFieldsConfig(rNameUpdated, uDescription, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloud9EnvironmentEc2Exists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "cloud9", regexp.MustCompile(`environment:.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "owner_arn", "aws_iam_user.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", "ec2"),
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
				ImportStateVerifyIgnore: []string{"instance_type", "subnet_id", "image_id"},
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
					testAccCheckAWSCloud9EnvironmentEc2Disappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloud9EnvironmentEc2_SSM(t *testing.T) {
	var conf cloud9.Environment

	rName := acctest.RandomWithPrefix("tf-acc-test")
	description := acctest.RandomWithPrefix("Tf Acc Test")
	userName := acctest.RandomWithPrefix("tf_acc_cloud9_env")
	resourceName := "aws_cloud9_environment_ec2.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloud9.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloud9.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloud9EnvironmentEc2Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloud9EnvironmentEc2SSMConfig(rName, description, userName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloud9EnvironmentEc2Exists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "cloud9", regexp.MustCompile(`environment:.+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "owner_arn", "aws_iam_user.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "type", "ec2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"instance_type", "automatic_stop_time_minutes", "subnet_id", "image_id"},
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

		out, err := conn.DescribeEnvironments(&cloud9.DescribeEnvironmentsInput{
			EnvironmentIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			if isAWSErr(err, cloud9.ErrCodeNotFoundException, "") {
				return fmt.Errorf("Cloud9 Environment EC2 (%q) not found", rs.Primary.ID)
			}
			return err
		}
		if len(out.Environments) == 0 {
			return fmt.Errorf("Cloud9 Environment EC2 (%q) not found", rs.Primary.ID)
		}
		env := out.Environments[0]

		*res = *env

		return nil
	}
}

func testAccCheckAWSCloud9EnvironmentEc2Disappears(res *cloud9.Environment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloud9conn

		_, err := conn.DeleteEnvironment(&cloud9.DeleteEnvironmentInput{
			EnvironmentId: res.Id,
		})

		if err != nil {
			return err
		}

		input := &cloud9.DescribeEnvironmentsInput{
			EnvironmentIds: []*string{res.Id},
		}
		var out *cloud9.DescribeEnvironmentsOutput
		err = resource.Retry(20*time.Minute, func() *resource.RetryError { // Deleting instances can take a long time
			out, err = conn.DescribeEnvironments(input)
			if err != nil {
				if isAWSErr(err, cloud9.ErrCodeNotFoundException, "") {
					return nil
				}
				if isAWSErr(err, "AccessDeniedException", "is not authorized to access this resource") {
					return nil
				}
				return resource.NonRetryableError(err)
			}
			if len(out.Environments) == 0 {
				return nil
			}
			return resource.RetryableError(fmt.Errorf("Cloud9 EC2 Environment %q still exists", aws.StringValue(res.Id)))
		})

		return err
	}
}

func testAccCheckAWSCloud9EnvironmentEc2Destroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloud9conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloud9_environment_ec2" {
			continue
		}

		out, err := conn.DescribeEnvironments(&cloud9.DescribeEnvironmentsInput{
			EnvironmentIds: []*string{aws.String(rs.Primary.ID)},
		})
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
		if len(out.Environments) == 0 {
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
  availability_zone       = data.aws_availability_zones.available.names[0]
  cidr_block              = "10.0.0.0/24"
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true
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
  connection_type             = "CONNECT_SSH"
  image_id                    = "amazonlinux-1-x86_64"
  name                        = %[1]q
  owner_arn                   = aws_iam_user.test.arn
  subnet_id                   = aws_subnet.test.id
}

resource "aws_iam_user" "test" {
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

func testAccAWSCloud9EnvironmentEc2SSMConfig(name, description, userName string) string {
	return testAccAWSCloud9EnvironmentEc2ConfigBase() + fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "cloud9_ssm_access" {
  statement {
    effect = "Allow"
    principals {
      identifiers = [
        "ec2.amazonaws.com",
        "cloud9.amazonaws.com"
      ]
      type = "Service"
    }
    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "cloud9_ssm_access" {
  name               = "AWSCloud9SSMAccessRole"
  path               = "/service-role/"
  assume_role_policy = data.aws_iam_policy_document.cloud9_ssm_access.json
}

resource "aws_iam_role_policy_attachment" "cloud9_ssm_access" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AWSCloud9SSMInstanceProfile"
  role       = aws_iam_role.cloud9_ssm_access.name
}

resource "aws_iam_instance_profile" "cloud9_ssm_access" {
  name = "AWSCloud9SSMInstanceProfile"
  path = "/cloud9/"
  role = aws_iam_role.cloud9_ssm_access.name
}

resource "aws_cloud9_environment_ec2" "test" {
  depends_on = [aws_route.test, aws_iam_instance_profile.cloud9_ssm_access]

  automatic_stop_time_minutes = 60
  description                 = %[2]q
  instance_type               = "t2.micro"
  connection_type             = "CONNECT_SSM"
  image_id                    = "amazonlinux-2-x86_64"
  name                        = %[1]q
  owner_arn                   = aws_iam_user.test.arn
  subnet_id                   = aws_subnet.test.id
}

resource "aws_iam_user" "test" {
  name = %[3]q
}
`, name, description, userName)
}
