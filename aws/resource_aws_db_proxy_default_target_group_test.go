package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func TestAccAWSDBProxyDefaultTargetGroup_Basic(t *testing.T) {
	var dbProxyTargetGroup rds.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyDefaultTargetGroupConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`target-group:.+`)),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "connection_pool_config.*", map[string]string{
						"connection_borrow_timeout":    "120",
						"init_query":                   "",
						"max_connections_percent":      "100",
						"max_idle_connections_percent": "50",
					}),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.session_pinning_filters.#", "0"),
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

func TestAccAWSDBProxyDefaultTargetGroup_ConnectionBorrowTimeout(t *testing.T) {
	var dbProxyTargetGroup rds.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyDefaultTargetGroupConfig_ConnectionBorrowTimeout(rName, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.connection_borrow_timeout", "120"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDBProxyDefaultTargetGroupConfig_ConnectionBorrowTimeout(rName, 90),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.connection_borrow_timeout", "90"),
				),
			},
		},
	})
}

func TestAccAWSDBProxyDefaultTargetGroup_InitQuery(t *testing.T) {
	var dbProxyTargetGroup rds.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyDefaultTargetGroupConfig_InitQuery(rName, "SET x=1, y=2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.init_query", "SET x=1, y=2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDBProxyDefaultTargetGroupConfig_InitQuery(rName, "SET a=2, b=1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.init_query", "SET a=2, b=1"),
				),
			},
		},
	})
}

func TestAccAWSDBProxyDefaultTargetGroup_MaxConnectionsPercent(t *testing.T) {
	var dbProxyTargetGroup rds.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyDefaultTargetGroupConfig_MaxConnectionsPercent(rName, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.max_connections_percent", "100"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDBProxyDefaultTargetGroupConfig_MaxConnectionsPercent(rName, 75),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.max_connections_percent", "75"),
				),
			},
		},
	})
}

func TestAccAWSDBProxyDefaultTargetGroup_MaxIdleConnectionsPercent(t *testing.T) {
	var dbProxyTargetGroup rds.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyDefaultTargetGroupConfig_MaxIdleConnectionsPercent(rName, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.max_idle_connections_percent", "50"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDBProxyDefaultTargetGroupConfig_MaxIdleConnectionsPercent(rName, 33),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.max_idle_connections_percent", "33"),
				),
			},
		},
	})
}

func TestAccAWSDBProxyDefaultTargetGroup_SessionPinningFilters(t *testing.T) {
	var dbProxyTargetGroup rds.DBProxyTargetGroup
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	sessionPinningFilters := "EXCLUDE_VARIABLE_SETS"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyDefaultTargetGroupConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.session_pinning_filters.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDBProxyDefaultTargetGroupConfig_SessionPinningFilters(rName, sessionPinningFilters),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyTargetGroupExists(resourceName, &dbProxyTargetGroup),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.session_pinning_filters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connection_pool_config.0.session_pinning_filters.0", sessionPinningFilters),
				),
			},
		},
	})
}

func TestAccAWSDBProxyDefaultTargetGroup_disappears(t *testing.T) {
	var v rds.DBProxy
	dbProxyResourceName := "aws_db_proxy.test"
	resourceName := "aws_db_proxy_default_target_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyTargetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyDefaultTargetGroupConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &v),
					// DB Proxy default Target Group implicitly exists so it cannot be removed.
					// Verify disappearance handling for DB Proxy removal instead.
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDbProxy(), dbProxyResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSDBProxyTargetGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_proxy_default_target_group" {
			continue
		}

		// Try to find the Group
		resp, err := conn.DescribeDBProxyTargetGroups(
			&rds.DescribeDBProxyTargetGroupsInput{
				DBProxyName:     aws.String(rs.Primary.ID),
				TargetGroupName: aws.String("default"),
			})

		if err == nil {
			if len(resp.TargetGroups) != 0 &&
				*resp.TargetGroups[0].DBProxyName == rs.Primary.ID {
				return fmt.Errorf("DB Proxy Target Group still exists")
			}
		}

		if !isAWSErr(err, rds.ErrCodeDBProxyNotFoundFault, "") {
			return err
		}
	}

	return nil
}

func testAccCheckAWSDBProxyTargetGroupExists(n string, v *rds.DBProxyTargetGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Proxy ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		opts := rds.DescribeDBProxyTargetGroupsInput{
			DBProxyName:     aws.String(rs.Primary.ID),
			TargetGroupName: aws.String("default"),
		}

		resp, err := conn.DescribeDBProxyTargetGroups(&opts)

		if err != nil {
			return err
		}

		if len(resp.TargetGroups) != 1 ||
			*resp.TargetGroups[0].DBProxyName != rs.Primary.ID {
			return fmt.Errorf("DB Proxy Target Group not found")
		}

		*v = *resp.TargetGroups[0]

		return nil
	}
}

func testAccAWSDBProxyDefaultTargetGroupConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = "%[1]s"
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

  tags = {
    Name = "%[1]s"
  }
}

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
`, rName)
}

func testAccAWSDBProxyDefaultTargetGroupConfig_Basic(rName string) string {
	return testAccAWSDBProxyDefaultTargetGroupConfigBase(rName) + `
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name

  connection_pool_config {
  }
}
`
}

func testAccAWSDBProxyDefaultTargetGroupConfig_ConnectionBorrowTimeout(rName string, connectionBorrowTimeout int) string {
	return testAccAWSDBProxyDefaultTargetGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name

  connection_pool_config {
    connection_borrow_timeout = %[2]d
  }
}
`, rName, connectionBorrowTimeout)
}

func testAccAWSDBProxyDefaultTargetGroupConfig_InitQuery(rName, initQuery string) string {
	return testAccAWSDBProxyDefaultTargetGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name

  connection_pool_config {
    init_query = "%[2]s"
  }
}
`, rName, initQuery)
}

func testAccAWSDBProxyDefaultTargetGroupConfig_MaxConnectionsPercent(rName string, maxConnectionsPercent int) string {
	return testAccAWSDBProxyDefaultTargetGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name

  connection_pool_config {
    max_connections_percent = %[2]d
  }
}
`, rName, maxConnectionsPercent)
}

func testAccAWSDBProxyDefaultTargetGroupConfig_MaxIdleConnectionsPercent(rName string, maxIdleConnectionsPercent int) string {
	return testAccAWSDBProxyDefaultTargetGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name

  connection_pool_config {
    max_idle_connections_percent = %[2]d
  }
}
`, rName, maxIdleConnectionsPercent)
}

func testAccAWSDBProxyDefaultTargetGroupConfig_SessionPinningFilters(rName, sessionPinningFilters string) string {
	return testAccAWSDBProxyDefaultTargetGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy_default_target_group" "test" {
  db_proxy_name = aws_db_proxy.test.name

  connection_pool_config {
    session_pinning_filters = ["%[2]s"]
  }
}
`, rName, sessionPinningFilters)
}
