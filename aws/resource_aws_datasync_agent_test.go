package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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
			if !strings.HasPrefix(name, "tf-acc-test-") {
				log.Printf("[INFO] Skipping DataSync Agent: %s", name)
				continue
			}
			log.Printf("[INFO] Deleting DataSync Agent: %s", name)
			input := &datasync.DeleteAgentInput{
				AgentArn: agent.AgentArn,
			}

			_, err := conn.DeleteAgent(input)

			if isAWSErr(err, "InvalidRequestException", "not found") {
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
	resourceName := "aws_datasync_agent.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncAgentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncAgentConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncAgentExists(resourceName, &agent1),
					resource.TestCheckResourceAttr(resourceName, "name", ""),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`agent/agent-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	resourceName := "aws_datasync_agent.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncAgentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncAgentConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncAgentExists(resourceName, &agent1),
					testAccCheckAWSDataSyncAgentDisappears(&agent1),
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
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncAgentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncAgentConfigName(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncAgentExists(resourceName, &agent1),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
				),
			},
			{
				Config: testAccAWSDataSyncAgentConfigName(rName2),
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
	resourceName := "aws_datasync_agent.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDataSync(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDataSyncAgentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSyncAgentConfigTags1("key1", "value1"),
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
				Config: testAccAWSDataSyncAgentConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDataSyncAgentExists(resourceName, &agent2),
					testAccCheckAWSDataSyncAgentNotRecreated(&agent1, &agent2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSDataSyncAgentConfigTags1("key1", "value1"),
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

func testAccCheckAWSDataSyncAgentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).datasyncconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datasync_agent" {
			continue
		}

		input := &datasync.DescribeAgentInput{
			AgentArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeAgent(input)

		if isAWSErr(err, "InvalidRequestException", "not found") {
			return nil
		}

		if err != nil {
			return err
		}
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
		input := &datasync.DescribeAgentInput{
			AgentArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeAgent(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Agent %q does not exist", rs.Primary.ID)
		}

		*agent = *output

		return nil
	}
}

func testAccCheckAWSDataSyncAgentDisappears(agent *datasync.DescribeAgentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).datasyncconn

		input := &datasync.DeleteAgentInput{
			AgentArn: agent.AgentArn,
		}

		_, err := conn.DeleteAgent(input)

		return err
	}
}

func testAccCheckAWSDataSyncAgentNotRecreated(i, j *datasync.DescribeAgentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.TimeValue(i.CreationTime) != aws.TimeValue(j.CreationTime) {
			return errors.New("DataSync Agent was recreated")
		}

		return nil
	}
}

// testAccAWSDataSyncAgentConfigAgentBase uses the "thinstaller" AMI
func testAccAWSDataSyncAgentConfigAgentBase() string {
	return fmt.Sprintf(`
data "aws_ami" "aws-thinstaller" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["aws-thinstaller-*"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-datasync-agent"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-datasync-agent"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-datasync-agent"
  }
}

resource "aws_route_table" "test" {
  vpc_id = "${aws_vpc.test.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.test.id}"
  }

  tags = {
    Name = "tf-acc-test-datasync-agent"
  }
}

resource "aws_route_table_association" "test" {
  subnet_id      = "${aws_subnet.test.id}"
  route_table_id = "${aws_route_table.test.id}"
}

resource "aws_security_group" "test" {
  vpc_id = "${aws_vpc.test.id}"

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
    Name = "tf-acc-test-datasync-agent"
  }
}

resource "aws_instance" "test" {
  depends_on = ["aws_internet_gateway.test"]

  ami                         = "${data.aws_ami.aws-thinstaller.id}"
  associate_public_ip_address = true
  # Default instance type from sync.sh
  instance_type               = "c5.2xlarge"
  vpc_security_group_ids      = ["${aws_security_group.test.id}"]
  subnet_id                   = "${aws_subnet.test.id}"

  tags = {
    Name = "tf-acc-test-datasync-agent"
  }
}
`)
}

func testAccAWSDataSyncAgentConfig() string {
	return testAccAWSDataSyncAgentConfigAgentBase() + fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  ip_address = "${aws_instance.test.public_ip}"
}
`)
}

func testAccAWSDataSyncAgentConfigName(rName string) string {
	return testAccAWSDataSyncAgentConfigAgentBase() + fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  ip_address = "${aws_instance.test.public_ip}"
  name       = %q
}
`, rName)
}

func testAccAWSDataSyncAgentConfigTags1(key1, value1 string) string {
	return testAccAWSDataSyncAgentConfigAgentBase() + fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  ip_address = "${aws_instance.test.public_ip}"

  tags = {
    %q = %q
  }
}
`, key1, value1)
}

func testAccAWSDataSyncAgentConfigTags2(key1, value1, key2, value2 string) string {
	return testAccAWSDataSyncAgentConfigAgentBase() + fmt.Sprintf(`
resource "aws_datasync_agent" "test" {
  ip_address = "${aws_instance.test.public_ip}"

  tags = {
    %q = %q
    %q = %q
  }
}
`, key1, value1, key2, value2)
}
