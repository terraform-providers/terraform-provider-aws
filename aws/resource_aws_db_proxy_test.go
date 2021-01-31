package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_db_proxy", &resource.Sweeper{
		Name: "aws_db_proxy",
		F:    testSweepRdsDbProxies,
	})
}

func testSweepRdsDbProxies(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*AWSClient).rdsconn

	err = conn.DescribeDBProxiesPages(&rds.DescribeDBProxiesInput{}, func(out *rds.DescribeDBProxiesOutput, lastPage bool) bool {
		for _, dbpg := range out.DBProxies {
			if dbpg == nil {
				continue
			}

			input := &rds.DeleteDBProxyInput{
				DBProxyName: dbpg.DBProxyName,
			}
			name := aws.StringValue(dbpg.DBProxyName)

			log.Printf("[INFO] Deleting DB Proxy: %s", name)

			_, err := conn.DeleteDBProxy(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DB Proxy %s: %s", name, err)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS DB Proxy sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving DB Proxies: %s", err)
	}

	return nil
}

func TestAccAWSDBProxy_basic(t *testing.T) {
	var v rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "engine_family", "MYSQL"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(`db-proxy:.+`)),
					resource.TestCheckResourceAttr(resourceName, "auth.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "auth.*", map[string]string{
						"auth_scheme": "SECRETS",
						"description": "test",
						"iam_auth":    "DISABLED",
					}),
					resource.TestCheckResourceAttr(resourceName, "debug_logging", "false"),
					resource.TestCheckResourceAttr(resourceName, "idle_client_timeout", "1800"),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "require_tls", "true"),
					resource.TestCheckResourceAttr(resourceName, "vpc_subnet_ids.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_subnet_ids.*", "aws_subnet.test.0", "id"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_subnet_ids.*", "aws_subnet.test.1", "id"),
					resource.TestMatchResourceAttr(resourceName, "endpoint", regexp.MustCompile(`^[\w\-\.]+\.rds\.amazonaws\.com$`))),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDBProxy_Name(t *testing.T) {
	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	nName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyConfigName(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDBProxyConfigName(rName, nName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "name", nName),
				),
			},
		},
	})
}

func TestAccAWSDBProxy_DebugLogging(t *testing.T) {
	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyConfigDebugLogging(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "debug_logging", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDBProxyConfigDebugLogging(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "debug_logging", "false"),
				),
			},
		},
	})
}

func TestAccAWSDBProxy_IdleClientTimeout(t *testing.T) {
	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyConfigIdleClientTimeout(rName, 900),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "idle_client_timeout", "900"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDBProxyConfigIdleClientTimeout(rName, 3600),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "idle_client_timeout", "3600"),
				),
			},
		},
	})
}

func TestAccAWSDBProxy_RequireTls(t *testing.T) {
	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyConfigRequireTls(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "require_tls", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDBProxyConfigRequireTls(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "require_tls", "false"),
				),
			},
		},
	})
}

func TestAccAWSDBProxy_RoleArn(t *testing.T) {
	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	nName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyConfigName(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDBProxyConfigRoleArn(rName, nName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.test2", "arn"),
				),
			},
		},
	})
}

func TestAccAWSDBProxy_VpcSecurityGroupIds(t *testing.T) {
	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	nName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyConfigName(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
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
				Config: testAccAWSDBProxyConfigVpcSecurityGroupIds(rName, nName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "vpc_security_group_ids.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "vpc_security_group_ids.*", "aws_security_group.test2", "id"),
				),
			},
		},
	})
}

func TestAccAWSDBProxy_AuthDescription(t *testing.T) {
	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	description := "foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyConfigName(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "auth.0.description", "test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDBProxyConfigAuthDescription(rName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "auth.0.description", description),
				),
			},
		},
	})
}

func TestAccAWSDBProxy_AuthIamAuth(t *testing.T) {
	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	iamAuth := "REQUIRED"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyConfigName(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "auth.0.iam_auth", "DISABLED"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDBProxyConfigAuthIamAuth(rName, iamAuth),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "auth.0.iam_auth", iamAuth),
				),
			},
		},
	})
}

func TestAccAWSDBProxy_AuthSecretArn(t *testing.T) {
	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	nName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyConfigName(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttrPair(resourceName, "auth.0.secret_arn", "aws_secretsmanager_secret.test", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDBProxyConfigAuthSecretArn(rName, nName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttrPair(resourceName, "auth.0.secret_arn", "aws_secretsmanager_secret.test2", "arn"),
				),
			},
		},
	})
}

func TestAccAWSDBProxy_Tags(t *testing.T) {
	var dbProxy rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	key := "foo"
	value := "bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyConfigName(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDBProxyConfigTags(rName, key, value),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &dbProxy),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", value),
				),
			},
		},
	})
}

func TestAccAWSDBProxy_disappears(t *testing.T) {
	var v rds.DBProxy
	resourceName := "aws_db_proxy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccDBProxyPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBProxyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBProxyExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDbProxy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// testAccDBProxyPreCheck checks if a call to describe db proxies errors out meaning feature not supported
func testAccDBProxyPreCheck(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	input := &rds.DescribeDBProxiesInput{}
	_, err := conn.DescribeDBProxies(input)

	if isAWSErr(err, "InvalidAction", "") {
		t.Skipf("skipping acceptance test, RDS Proxy not supported: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckAWSDBProxyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_proxy" {
			continue
		}

		// Try to find the Group
		resp, err := conn.DescribeDBProxies(
			&rds.DescribeDBProxiesInput{
				DBProxyName: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(resp.DBProxies) != 0 &&
				*resp.DBProxies[0].DBProxyName == rs.Primary.ID {
				return fmt.Errorf("DB Proxy still exists")
			}
		}

		if !isAWSErr(err, rds.ErrCodeDBProxyNotFoundFault, "") {
			return err
		}
	}

	return nil
}

func testAccCheckAWSDBProxyExists(n string, v *rds.DBProxy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DB Proxy ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		opts := rds.DescribeDBProxiesInput{
			DBProxyName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeDBProxies(&opts)

		if err != nil {
			return err
		}

		if len(resp.DBProxies) != 1 ||
			*resp.DBProxies[0].DBProxyName != rs.Primary.ID {
			return fmt.Errorf("DB Proxy not found")
		}

		*v = *resp.DBProxies[0]

		return nil
	}
}

func testAccAWSDBProxyConfigBase(rName string) string {
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
`, rName)
}

func testAccAWSDBProxyConfig(rName string) string {
	return testAccAWSDBProxyConfigBase(rName) + fmt.Sprintf(`
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
`, rName)
}

func testAccAWSDBProxyConfigName(rName, nName string) string {
	return testAccAWSDBProxyConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = "%[2]s"
  engine_family          = "MYSQL"
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
`, rName, nName)
}

func testAccAWSDBProxyConfigDebugLogging(rName string, debugLogging bool) string {
	return testAccAWSDBProxyConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name           = "%[1]s"
  debug_logging  = %[2]t
  engine_family  = "MYSQL"
  role_arn       = aws_iam_role.test.arn
  vpc_subnet_ids = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}
`, rName, debugLogging)
}

func testAccAWSDBProxyConfigIdleClientTimeout(rName string, idleClientTimeout int) string {
	return testAccAWSDBProxyConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                = "%[1]s"
  idle_client_timeout = %[2]d
  engine_family       = "MYSQL"
  role_arn            = aws_iam_role.test.arn
  vpc_subnet_ids      = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}
`, rName, idleClientTimeout)
}

func testAccAWSDBProxyConfigRequireTls(rName string, requireTls bool) string {
	return testAccAWSDBProxyConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name           = "%[1]s"
  require_tls    = %[2]t
  engine_family  = "MYSQL"
  role_arn       = aws_iam_role.test.arn
  vpc_subnet_ids = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}
`, rName, requireTls)
}

func testAccAWSDBProxyConfigRoleArn(rName, nName string) string {
	return testAccAWSDBProxyConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test2
  ]

  name           = "%[1]s"
  engine_family  = "MYSQL"
  role_arn       = aws_iam_role.test2.arn
  vpc_subnet_ids = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}

# IAM setup

resource "aws_iam_role" "test2" {
  name               = "%[2]s"
  assume_role_policy = data.aws_iam_policy_document.assume.json
}

resource "aws_iam_role_policy" "test2" {
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test.json
}
`, rName, nName)
}

func testAccAWSDBProxyConfigVpcSecurityGroupIds(rName, nName string) string {
	return testAccAWSDBProxyConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = "%[1]s"
  engine_family          = "MYSQL"
  role_arn               = aws_iam_role.test.arn
  vpc_security_group_ids = [aws_security_group.test2.id]
  vpc_subnet_ids         = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}

resource "aws_security_group" "test2" {
  name   = "%[2]s"
  vpc_id = aws_vpc.test.id
}
`, rName, nName)
}

func testAccAWSDBProxyConfigAuthDescription(rName, description string) string {
	return testAccAWSDBProxyConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = "%[1]s"
  engine_family          = "MYSQL"
  role_arn               = aws_iam_role.test.arn
  vpc_security_group_ids = [aws_security_group.test.id]
  vpc_subnet_ids         = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "%[2]s"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}
`, rName, description)
}

func testAccAWSDBProxyConfigAuthIamAuth(rName, iamAuth string) string {
	return testAccAWSDBProxyConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = "%[1]s"
  engine_family          = "MYSQL"
  role_arn               = aws_iam_role.test.arn
  require_tls            = true
  vpc_security_group_ids = [aws_security_group.test.id]
  vpc_subnet_ids         = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "%[2]s"
    secret_arn  = aws_secretsmanager_secret.test.arn
  }
}
`, rName, iamAuth)
}

func testAccAWSDBProxyConfigAuthSecretArn(rName, nName string) string {
	return testAccAWSDBProxyConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = "%[1]s"
  engine_family          = "MYSQL"
  role_arn               = aws_iam_role.test.arn
  vpc_security_group_ids = [aws_security_group.test.id]
  vpc_subnet_ids         = aws_subnet.test.*.id

  auth {
    auth_scheme = "SECRETS"
    description = "test"
    iam_auth    = "DISABLED"
    secret_arn  = aws_secretsmanager_secret.test2.arn
  }
}

resource "aws_secretsmanager_secret" "test2" {
  name                    = "%[2]s"
  recovery_window_in_days = 0
}

resource "aws_secretsmanager_secret_version" "test2" {
  secret_id     = aws_secretsmanager_secret.test2.id
  secret_string = "{\"username\":\"db_user\",\"password\":\"db_user_password\"}"
}
`, rName, nName)
}

func testAccAWSDBProxyConfigTags(rName, key, value string) string {
	return testAccAWSDBProxyConfigBase(rName) + fmt.Sprintf(`
resource "aws_db_proxy" "test" {
  depends_on = [
    aws_secretsmanager_secret_version.test,
    aws_iam_role_policy.test
  ]

  name                   = "%[1]s"
  engine_family          = "MYSQL"
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
    %[2]s = "%[3]s"
  }
}
`, rName, key, value)
}
