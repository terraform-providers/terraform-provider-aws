package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/rds/finder"
)

func TestAccAWSDBProxyEndpoint_basic(t *testing.T) {
	var v rds.DBProxyEndpoint
	resourceName := "aws_db_proxy_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyEndpointPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "db_proxy_endpoint_name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "db_proxy_name", "aws_db_proxy.test", "name"),
					resource.TestCheckResourceAttr(resourceName, "target_role", "READ_WRITE"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`db-proxy-endpoint:.+`)),
					resource.TestCheckResourceAttr(resourceName, "vpc_subnet_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_subnet_ids.*", "aws_subnet.test.0", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_subnet_ids.*", "aws_subnet.test.1", "id"),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", "aws_vpc.test", "id"),
					resource.TestMatchResourceAttr(resourceName, "endpoint", regexp.MustCompile(`^[\w\-\.]+\.rds\.amazonaws\.com$`)),
					resource.TestCheckResourceAttr(resourceName, "is_default", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
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

func TestAccAWSDBProxyEndpoint_targetRole(t *testing.T) {
	var v rds.DBProxyEndpoint
	resourceName := "aws_db_proxy_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyEndpointPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyEndpointConfigTargetRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyEndpointExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "target_role", "READ_ONLY"),
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

func TestAccAWSDBProxyEndpoint_vpcSecurityGroupIds(t *testing.T) {
	var dbProxy rds.DBProxyEndpoint
	resourceName := "aws_db_proxy_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyEndpointPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyEndpointConfigVpcSecurityGroupIds1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyEndpointExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_security_group_ids.*", "aws_security_group.test", "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDBProxyEndpointConfigVpcSecurityGroupIds2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyEndpointExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_security_group_ids.*", "aws_security_group.test", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_security_group_ids.*", "aws_security_group.test2", "id"),
				),
			},
		},
	})
}

func TestAccAWSDBProxyEndpoint_tags(t *testing.T) {
	var dbProxy rds.DBProxyEndpoint
	resourceName := "aws_db_proxy_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyEndpointPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyEndpointConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyEndpointExists(resourceName, &dbProxy),
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
				Config: testAccAWSDBProxyEndpointConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyEndpointExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSDBProxyEndpointConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyEndpointExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSDBProxyEndpoint_disappears(t *testing.T) {
	var v rds.DBProxyEndpoint
	resourceName := "aws_db_proxy_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyEndpointPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyEndpointExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDbProxyEndpoint(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDBProxyEndpoint_disappears_proxy(t *testing.T) {
	var v rds.DBProxyEndpoint
	resourceName := "aws_db_proxy_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyEndpointPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, rds.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyEndpointConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyEndpointExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDbProxy(), "aws_db_proxy.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// testAccDBProxyEndpointPreCheck checks if a call to describe db proxies errors out meaning feature not supported
func testAccDBProxyEndpointPreCheck(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	_, err := conn.DescribeDBProxyEndpoints(&rds.DescribeDBProxyEndpointsInput{})

	if isAWSErr(err, "InvalidAction", "") {
		t.Skipf("skipping acceptance test, RDS Proxy not supported: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckAWSDBProxyEndpointDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_proxy_endpoint" {
			continue
		}

		dbProxyEndpoint, err := finder.DBProxyEndpoint(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyNotFoundFault) || tfawserr.ErrCodeEquals(err, rds.ErrCodeDBProxyEndpointNotFoundFault) {
			continue
		}

		if err != nil {
			return err
		}

		if dbProxyEndpoint != nil {
			return fmt.Errorf("RDS DB Proxy Endpoint (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSDBProxyEndpointExists(n string, v *rds.DBProxyEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Proxy ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		dbProxyEndpoint, err := finder.DBProxyEndpoint(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if dbProxyEndpoint == nil {
			return fmt.Errorf("RDS DB Proxy Endpoint (%s) still not found", rs.Primary.ID)
		}

		*v = *dbProxyEndpoint

		return nil
	}
}

func testAccAWSDBProxyEndpointConfigBase(rName string) string {
	return fmt.Sprintf(`
# Secrets Manager setup

resource "aws_secretsmanager_secret" "test" {
  name                    = "%[1]s"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = "{\"username\":\"db_user\",\"password\":\"db_user_password\"}"
}

# IAM setup

resource "aws_iam_role" "test" {
  name               = "%[1]s"
  assume_role_policy = data.aws_iam_policy_document.assume.json
}

data "aws_iam_policy_document" "assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["rds.amazonaws.com"]
    }
  }
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "secretsmanager:GetRandomPassword",
      "secretsmanager:CreateSecret",
      "secretsmanager:ListSecrets",
    ]
    resources = ["*"]
  }

  statement {
    actions   = ["secretsmanager:*"]
    resources = [aws_secretsmanager_secret.test.arn]
  }
}

# VPC setup

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "%[1]s"
  }
}

resource "aws_security_group" "test" {
  name   = "%[1]s"
  vpc_id = aws_vpc.test.id
}

resource "aws_subnet" "test" {
  count             = 2
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "%[1]s-${count.index}"
  }
}

resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = %[1]q
  debug_logging          = false
  engine_family          = "MYSQL"
  idle_client_timeout    = 1800
  require_tls            = true
  role_arn               = aws_iam_role.test.arn
  vpc_security_group_ids = [aws_security_group.test.id]
  vpc_subnet_ids         = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}
`, rName)
}

func testAccAWSDBProxyEndpointConfig(rName string) string {
	return testAccAWSDBProxyEndpointConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy_endpoint" "test" {
  db_proxy_name          = aws_db_proxy.test.name
  db_proxy_endpoint_name = %[1]q
  vpc_subnet_ids         = aws_subnet.test.*.id
}
`, rName)
}

func testAccAWSDBProxyEndpointConfigTargetRole(rName string) string {
	return testAccAWSDBProxyEndpointConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy_endpoint" "test" {
  db_proxy_name          = aws_db_proxy.test.name
  db_proxy_endpoint_name = %[1]q
  vpc_subnet_ids         = aws_subnet.test.*.id
  target_role            = "READ_ONLY"
}
`, rName)
}

func testAccAWSDBProxyEndpointConfigVpcSecurityGroupIds1(rName string) string {
	return testAccAWSDBProxyEndpointConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy_endpoint" "test" {
  db_proxy_name          = aws_db_proxy.test.name
  db_proxy_endpoint_name = %[1]q
  vpc_subnet_ids         = aws_subnet.test.*.id
  vpc_security_group_ids = [aws_security_group.test.id]
}
`, rName)
}

func testAccAWSDBProxyEndpointConfigVpcSecurityGroupIds2(rName string) string {
	return testAccAWSDBProxyEndpointConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy_endpoint" "test" {
  db_proxy_name          = aws_db_proxy.test.name
  db_proxy_endpoint_name = %[1]q
  vpc_subnet_ids         = aws_subnet.test.*.id
  vpc_security_group_ids = [aws_security_group.test.id, aws_security_group.test2.id]
}

resource "aws_security_group" "test2" {
  name   = "%[1]s-2"
  vpc_id = aws_vpc.test.id
}
`, rName)
}

func testAccAWSDBProxyEndpointConfigTags1(rName, key1, value1 string) string {
	return testAccAWSDBProxyEndpointConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy_endpoint" "test" {
  db_proxy_name          = aws_db_proxy.test.name
  db_proxy_endpoint_name = %[1]q
  vpc_subnet_ids         = aws_subnet.test.*.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, key1, value1)
}

func testAccAWSDBProxyEndpointConfigTags2(rName, key1, value1, key2, value2 string) string {
	return testAccAWSDBProxyEndpointConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy_endpoint" "test" {
  db_proxy_name          = aws_db_proxy.test.name
  db_proxy_endpoint_name = %[1]q
  vpc_subnet_ids         = aws_subnet.test.*.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, key1, value1, key2, value2)
}
