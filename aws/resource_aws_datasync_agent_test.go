package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/datasync/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_datasync_agent", &resource.Sweeper{
		Name: "aws_datasync_agent",
		F:    testSweepDataSyncAgents,
	})
}

func testSweepDataSyncAgents(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).datasyncconn

	input := &datasync.ListAgentsInput{}
	for {
		output, err := conn.ListAgents(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Agent sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving DataSync Agents: %s", err)
		}

		if len(output.Agents) == 0 {
			log.Print("[DEBUG] No DataSync Agents to sweep")
			return nil
		}

		for _, agent := range output.Agents {
			name := aws.StringValue(agent.Name)

			log.Printf("[INFO] Deleting DataSync Agent: %s", name)
			input := &datasync.DeleteAgentInput{
				AgentArn: agent.AgentArn,
			}

			_, err := conn.DeleteAgent(input)

			if isAWSErr(err, "InvalidRequestException", "does not exist") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Failed to delete DataSync Agent (%s): %s", name, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSDataSyncAgent_basic(t *testing.T) {
	var agent1 datasync.DescribeAgentOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_agent.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncAgentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncAgentConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncAgentExists(resourceName, &agent1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`agent/agent-.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", ""),
					resource.TestCheckResourceAttr(resourceName, "private_link_endpoint", ""),
					resource.TestCheckResourceAttr(resourceName, "security_group_arns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "ip_address"},
			},
		},
	})
}

func TestAccAWSDataSyncAgent_disappears(t *testing.T) {
	var agent1 datasync.DescribeAgentOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_agent.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncAgentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncAgentConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncAgentExists(resourceName, &agent1),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDataSyncAgent(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDataSyncAgent_AgentName(t *testing.T) {
	var agent1, agent2 datasync.DescribeAgentOutput
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_agent.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncAgentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncAgentConfigName(rName1, rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncAgentExists(resourceName, &agent1),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				Config: testAccAWSDataSyncAgentConfigName(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncAgentExists(resourceName, &agent2),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "ip_address"},
			},
		},
	})
}

func TestAccAWSDataSyncAgent_Tags(t *testing.T) {
	var agent1, agent2, agent3 datasync.DescribeAgentOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_agent.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncAgentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncAgentConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncAgentExists(resourceName, &agent1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "ip_address"},
			},
			{
				Config: testAccAWSDataSyncAgentConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncAgentExists(resourceName, &agent2),
					testAccCheckAWSDataSyncAgentNotRecreated(&agent1, &agent2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSDataSyncAgentConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncAgentExists(resourceName, &agent3),
					testAccCheckAWSDataSyncAgentNotRecreated(&agent2, &agent3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func TestAccAWSDataSyncAgent_VpcEndpointId(t *testing.T) {
	var agent datasync.DescribeAgentOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_datasync_agent.test"
	securityGroupResourceName := "aws_security_group.test"
	subnetResourceName := "aws_subnet.test"
	vpcEndpointResourceName := "aws_vpc_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		ErrorCheck:   testAccErrorCheck(t, datasync.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncAgentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncAgentConfigVpcEndpointId(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncAgentExists(resourceName, &agent),
					resource.TestCheckResourceAttr(resourceName, "security_group_arns.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "security_group_arns.*", securityGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "subnet_arns.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "subnet_arns.*", subnetResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_endpoint_id", vpcEndpointResourceName, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "ip_address", "private_link_ip"},
			},
		},
	})
}

func testAccCheckAWSDataSyncAgentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).datasyncconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datasync_agent" {
			continue
		}

		_, err := finder.AgentByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("DataSync Agent %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSDataSyncAgentExists(resourceName string, agent *datasync.DescribeAgentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).datasyncconn

		output, err := finder.AgentByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*agent = *output

		return nil
	}
}

func testAccCheckAWSDataSyncAgentNotRecreated(i, j *datasync.DescribeAgentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreationTime).Equal(aws.TimeValue(j.CreationTime)) {
			return errors.New("DataSync Agent was recreated")
		}

		return nil
	}
}

func testAccAWSDataSyncAgentConfigAgentBase(rName string) string {
	return fmt.Sprintf(`
# Reference: https://docs.aws.amazon.com/datasync/latest/userguide/deploy-agents.html
data "aws_ssm_parameter" "aws_service_datasync_ami" {
  name = "/aws/service/datasync/ami"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = aws_subnet.test.id
  route_table_id = aws_route_table.test.id
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  depends_on = [aws_internet_gateway.test]

  ami                         = data.aws_ssm_parameter.aws_service_datasync_ami.value
  associate_public_ip_address = true

  # Default instance type from sync.sh
  instance_type          = "c5.2xlarge"
  vpc_security_group_ids = [aws_security_group.test.id]
  subnet_id              = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSDataSyncAgentConfig(rName string) string {
	return composeConfig(testAccAWSDataSyncAgentConfigAgentBase(rName), `
resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip
}
`)
}

func testAccAWSDataSyncAgentConfigName(rName, agentName string) string {
	return composeConfig(testAccAWSDataSyncAgentConfigAgentBase(rName), fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip
  name       = %[1]q
}
`, agentName))
}

func testAccAWSDataSyncAgentConfigTags1(rName, key1, value1 string) string {
	return composeConfig(testAccAWSDataSyncAgentConfigAgentBase(rName), fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip

  tags = {
    %[1]q = %[2]q
  }
}
`, key1, value1))
}

func testAccAWSDataSyncAgentConfigTags2(rName, key1, value1, key2, value2 string) string {
	return composeConfig(testAccAWSDataSyncAgentConfigAgentBase(rName), fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, key1, value1, key2, value2))
}

func testAccAWSDataSyncAgentConfigVpcEndpointId(rName string) string {
	return composeConfig(testAccAWSDataSyncAgentConfigAgentBase(rName), fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  name                  = %[1]q
  security_group_arns   = [aws_security_group.test.arn]
  subnet_arns           = [aws_subnet.test.arn]
  vpc_endpoint_id       = aws_vpc_endpoint.test.id
  ip_address            = aws_instance.test.public_ip
  private_link_endpoint = data.aws_network_interface.test.private_ip
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  service_name       = "com.amazonaws.${data.aws_region.current.name}.datasync"
  vpc_id             = aws_vpc.test.id
  security_group_ids = [aws_security_group.test.id]
  subnet_ids         = [aws_subnet.test.id]
  vpc_endpoint_type  = "Interface"

  tags = {
    Name = %[1]q
  }
}

data "aws_network_interface" "test" {
  id = tolist(aws_vpc_endpoint.test.network_interface_ids)[0]
}
`, rName))
}
