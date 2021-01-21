package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_transfer_server", &resource.Sweeper{
		Name: "aws_transfer_server",
		F:    testSweepTransferServers,
	})
}

func testSweepTransferServers(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).transferconn
	input := &transfer.ListServersInput{}

	err = conn.ListServersPages(input, func(page *transfer.ListServersOutput, lastPage bool) bool {
		for _, server := range page.Servers {
			id := aws.StringValue(server.ServerId)
			input := &transfer.DeleteServerInput{
				ServerId: server.ServerId,
			}

			log.Printf("[INFO] Deleting Transfer Server: %s", id)
			_, err := conn.DeleteServer(input)

			if err != nil {
				log.Printf("[ERROR] Error deleting Transfer Server (%s): %s", id, err)
				continue
			}

			if err := waitForTransferServerDeletion(conn, id); err != nil {
				log.Printf("[ERROR] Error waiting for Transfer Server (%s) deletion: %s", id, err)
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Transfer Server sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Transfer Servers: %s", err)
	}

	return nil
}

func TestAccAWSTransferServer_basic(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.foo"
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "transfer", regexp.MustCompile(`server/.+`)),
					resource.TestMatchResourceAttr(
						resourceName, "endpoint", regexp.MustCompile(fmt.Sprintf("^s-[a-z0-9]+.server.transfer.%s.amazonaws.com$", testAccGetRegion()))),
					resource.TestCheckResourceAttr(
						resourceName, "identity_provider_type", "SERVICE_MANAGED"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccAWSTransferServerConfig_basicUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.NAME", "tf-acc-test-transfer-server"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.ENV", "test"),
					resource.TestCheckResourceAttrPair(
						resourceName, "logging_role", "aws_iam_role.foo", "arn"),
				),
			},
		},
	})
}

func TestAccAWSTransferServer_Vpc(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_Vpc,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(
						resourceName, "endpoint_type", "VPC"),
					resource.TestCheckResourceAttr(
						resourceName, "endpoint_details.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "endpoint_details.0.address_allocation_ids.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			{
				Config: testAccAWSTransferServerConfig_VpcUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(
						resourceName, "endpoint_type", "VPC"),
					resource.TestCheckResourceAttr(
						resourceName, "endpoint_details.0.address_allocation_ids.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSTransferServer_apigateway(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.foo"
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t); testAccPreCheckAWSTransfer(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_apigateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(
						resourceName, "identity_provider_type", "API_GATEWAY"),
					resource.TestCheckResourceAttrSet(
						resourceName, "invocation_role"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.NAME", "tf-acc-test-transfer-server"),
					resource.TestCheckResourceAttr(
						resourceName, "tags.TYPE", "apigateway"),
				),
			},
		},
	})
}

func TestAccAWSTransferServer_disappears(t *testing.T) {
	var conf transfer.DescribedServer

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists("aws_transfer_server.foo", &conf),
					testAccCheckAWSTransferServerDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSTransferServer_forcedestroy(t *testing.T) {
	var conf transfer.DescribedServer
	var roleConf iam.GetRoleOutput
	resourceName := "aws_transfer_server.foo"
	rName := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_forcedestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					testAccCheckAWSRoleExists("aws_iam_role.foo", &roleConf),
					resource.TestCheckResourceAttr(
						resourceName, "identity_provider_type", "SERVICE_MANAGED"),
					resource.TestCheckResourceAttr(
						resourceName, "force_destroy", "true"),
					testAccCheckAWSTransferCreateUser(&conf, &roleConf, rName),
					testAccCheckAWSTransferCreateSshKey(&conf, rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "host_key"},
			},
		},
	})
}

func TestAccAWSTransferServer_vpcEndpointId(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t); testAccPreCheckAWSTransfer(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_VpcEndPoint,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(
						resourceName, "endpoint_type", "VPC_ENDPOINT"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "host_key"},
			},
		},
	})
}

func TestAccAWSTransferServer_hostKey(t *testing.T) {
	var conf transfer.DescribedServer
	resourceName := "aws_transfer_server.default"
	hostKey := "test-fixtures/transfer-ssh-rsa-key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSTransferServerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSTransferServerConfig_hostKey(hostKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSTransferServerExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "host_key_fingerprint", "SHA256:Z2pW9sPKDD/T34tVfCoolsRcECNTlekgaKvDn9t+9sg="),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy", "host_key"},
			},
		},
	})
}

func testAccCheckAWSTransferServerExists(n string, res *transfer.DescribedServer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Transfer Server ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).transferconn

		describe, err := conn.DescribeServer(&transfer.DescribeServerInput{
			ServerId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		*res = *describe.Server

		return nil
	}
}

func testAccCheckAWSTransferServerDisappears(conf *transfer.DescribedServer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).transferconn

		params := &transfer.DeleteServerInput{
			ServerId: conf.ServerId,
		}

		_, err := conn.DeleteServer(params)
		if err != nil {
			return err
		}

		return waitForTransferServerDeletion(conn, *conf.ServerId)
	}
}

func testAccCheckAWSTransferServerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).transferconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_transfer_server" {
			continue
		}

		_, err := conn.DescribeServer(&transfer.DescribeServerInput{
			ServerId: aws.String(rs.Primary.ID),
		})

		if isAWSErr(err, transfer.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err == nil {
			return fmt.Errorf("Transfer Server (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSTransferCreateUser(describedServer *transfer.DescribedServer, getRoleOutput *iam.GetRoleOutput, userName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).transferconn

		input := &transfer.CreateUserInput{
			ServerId: describedServer.ServerId,
			UserName: aws.String(userName),
			Role:     getRoleOutput.Role.Arn,
		}

		if _, err := conn.CreateUser(input); err != nil {
			return fmt.Errorf("error creating Transfer User (%s) on Server (%s): %s", userName, aws.StringValue(describedServer.ServerId), err)
		}

		return nil
	}
}

func testAccCheckAWSTransferCreateSshKey(describedServer *transfer.DescribedServer, userName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).transferconn

		input := &transfer.ImportSshPublicKeyInput{
			ServerId:         describedServer.ServerId,
			UserName:         aws.String(userName),
			SshPublicKeyBody: aws.String("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"),
		}

		if _, err := conn.ImportSshPublicKey(input); err != nil {
			return fmt.Errorf("error creating Transfer SSH Public Key for  (%s/%s): %s", userName, aws.StringValue(describedServer.ServerId), err)
		}

		return nil
	}
}

func testAccPreCheckAWSTransfer(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).transferconn

	input := &transfer.ListServersInput{}

	_, err := conn.ListServers(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

const testAccAWSTransferServerConfig_basic = `
resource "aws_transfer_server" "foo" {}
`

func testAccAWSTransferServerConfig_basicUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "foo" {
  name = "tf-test-transfer-server-iam-role-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "transfer.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "foo" {
  name = "tf-test-transfer-server-iam-policy-%[1]s"
  role = aws_iam_role.foo.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowFullAccesstoCloudWatchLogs",
      "Effect": "Allow",
      "Action": [
        "logs:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_transfer_server" "foo" {
  identity_provider_type = "SERVICE_MANAGED"
  logging_role           = aws_iam_role.foo.arn

  tags = {
    NAME = "tf-acc-test-transfer-server"
    ENV  = "test"
  }
}
`, rName)
}

func testAccAWSTransferServerConfig_apigateway(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = "test"
}

resource "aws_api_gateway_resource" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "test"
}

resource "aws_api_gateway_method" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.test.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "error" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method
  status_code = "400"
}

resource "aws_api_gateway_integration" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_method.test.http_method

  type                    = "HTTP"
  uri                     = "https://www.google.de"
  integration_http_method = "GET"
}

resource "aws_api_gateway_integration_response" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.test.id
  http_method = aws_api_gateway_integration.test.http_method
  status_code = aws_api_gateway_method_response.error.status_code
}

resource "aws_api_gateway_deployment" "test" {
  depends_on = [aws_api_gateway_integration.test]

  rest_api_id       = aws_api_gateway_rest_api.test.id
  stage_name        = "test"
  description       = "%[1]s"
  stage_description = "%[1]s"

  variables = {
    "a" = "2"
  }
}

resource "aws_iam_role" "foo" {
  name = "tf-test-transfer-server-iam-role-for-apigateway-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "transfer.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "foo" {
  name = "tf-test-transfer-server-iam-policy-%[1]s"
  role = aws_iam_role.foo.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowFullAccesstoCloudWatchLogs",
      "Effect": "Allow",
      "Action": [
        "logs:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}

data "aws_region" "current" {}

resource "aws_transfer_server" "foo" {
  identity_provider_type = "API_GATEWAY"
  url                    = "https://${aws_api_gateway_rest_api.test.id}.execute-api.${data.aws_region.current.name}.amazonaws.com${aws_api_gateway_resource.test.path}"
  invocation_role        = aws_iam_role.foo.arn
  logging_role           = aws_iam_role.foo.arn

  tags = {
    NAME = "tf-acc-test-transfer-server"
    TYPE = "apigateway"
  }
}
`, rName)

}

func testAccAWSTransferServerConfig_forcedestroy(rName string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "foo" {
  force_destroy = true
}

resource "aws_iam_role" "foo" {
  name = "tf-test-transfer-user-iam-role-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "transfer.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "foo" {
  name = "tf-test-transfer-user-iam-policy-%[1]s"
  role = aws_iam_role.foo.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AllowFullAccesstoS3",
      "Effect": "Allow",
      "Action": [
        "s3:*"
      ],
      "Resource": "*"
    }
  ]
}
POLICY
}
`, rName)
}

const testAccAWSTransferServerConfig_VpcEndPoint = `

data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-default-route-table-vpc-endpoint"
  }
}

resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "test"
  }
}

resource "aws_security_group" "sg" {
  name        = "allow-transfer-server"
  description = "Allow TLS inbound traffic"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }
}

resource "aws_vpc_endpoint" "transfer" {
  vpc_id            = aws_vpc.test.id
  vpc_endpoint_type = "Interface"
  service_name      = "com.amazonaws.${data.aws_region.current.name}.transfer.server"

  security_group_ids = [
    aws_security_group.sg.id,
  ]
}

resource "aws_default_route_table" "foo" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  tags = {
    Name = "test"
  }

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }
}

resource "aws_transfer_server" "default" {
  endpoint_type = "VPC_ENDPOINT"

  endpoint_details {
    vpc_endpoint_id = aws_vpc_endpoint.transfer.id
  }
}
`

const testAccAWSTransferServerConfig_VpcDefault = `
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "terraform-testacc-igw"
  }
}

resource "aws_subnet" "test" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.0.0/24"
  map_public_ip_on_launch = true

  depends_on = [aws_internet_gateway.test]
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = "terraform-testacc-subnet"
  }
}

resource "aws_security_group" "test" {
  name   = "terraform-testacc-security-group"
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "terraform-testacc-security-group"
  }
}

resource "aws_eip" "testa" {
  vpc = true
}

resource "aws_eip" "testb" {
  vpc = true
}
`

const testAccAWSTransferServerConfig_Vpc = testAccAWSTransferServerConfig_VpcDefault + `
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    address_allocation_ids = [aws_eip.testa.id]
    subnet_ids             = [aws_subnet.test.id]
    vpc_id                 = aws_vpc.test.id
  }
}
`

const testAccAWSTransferServerConfig_VpcUpdate = testAccAWSTransferServerConfig_VpcDefault + `
resource "aws_transfer_server" "test" {
  endpoint_type = "VPC"

  endpoint_details {
    address_allocation_ids = [aws_eip.testb.id]
    subnet_ids             = [aws_subnet.test.id]
    vpc_id                 = aws_vpc.test.id
  }
}
`

func testAccAWSTransferServerConfig_hostKey(hostKey string) string {
	return fmt.Sprintf(`
resource "aws_transfer_server" "default" {
  host_key = file("%s")
}
`, hostKey)
}
