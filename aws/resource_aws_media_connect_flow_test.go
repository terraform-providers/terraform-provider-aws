package aws

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediaconnect"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSMediaConnectFlowConfig_Base(t *testing.T) {
	var flow mediaconnect.Flow

	rName := fmt.Sprintf("tf-acctest_%s", acctest.RandString(5))
	resourceName := "aws_media_connect_flow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSMediaConnectFlowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSMediaConnectFlowConfig_Base(rName, "rtp", "3010"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSMediaConnectFlowExists(resourceName, &flow),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "mediaconnect", regexp.MustCompile(`flow:.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "egress_ip"),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					testAccMatchResourceAttrRegionalARN(resourceName, "source.0.arn", "mediaconnect", regexp.MustCompile(`source:.+`)),
					resource.TestCheckResourceAttr(resourceName, "source.0.name", rName),
					resource.TestCheckResourceAttr(resourceName, "source.0.description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "source.0.protocol", "rtp"),
					resource.TestCheckResourceAttr(resourceName, "source.0.ingest_port", "3010"),
					resource.TestCheckResourceAttr(resourceName, "source.0.whitelist_cidr", "10.24.34.0/23"),
					resource.TestCheckResourceAttrSet(resourceName, "source.0.ingest_ip"),
				),
			},
			{
				Config: testAccAWSMediaConnectFlowConfig_Base(rName, "rtp-fec", "3333"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSMediaConnectFlowExists(resourceName, &flow),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.protocol", "rtp-fec"),
					resource.TestCheckResourceAttr(resourceName, "source.0.ingest_port", "3333"),
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

func TestAccAWSMediaConnectFlowConfig_Decryption(t *testing.T) {
	var flow mediaconnect.Flow

	rName := fmt.Sprintf("tfacctest%s", acctest.RandString(5))
	resourceName := "aws_media_connect_flow.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSMediaConnectFlowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSMediaConnectFlowConfig_Decryption(rName, "aes128"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSMediaConnectFlowExists(resourceName, &flow),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "mediaconnect", regexp.MustCompile(`flow:.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "egress_ip"),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					testAccMatchResourceAttrRegionalARN(resourceName, "source.0.arn", "mediaconnect", regexp.MustCompile(`source:.+`)),
					resource.TestCheckResourceAttr(resourceName, "source.0.description", "Managed by Terraform"),
					resource.TestCheckResourceAttr(resourceName, "source.0.decryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.decryption.0.key_type", "static-key"),
					resource.TestCheckResourceAttr(resourceName, "source.0.decryption.0.algorithm", "aes128"),
					resource.TestCheckResourceAttrPair(resourceName, "source.0.decryption.0.role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "source.0.decryption.0.secret_arn", "aws_secretsmanager_secret.test", "arn"),
				),
			},
			{
				Config: testAccAWSMediaConnectFlowConfig_Decryption(rName, "aes256"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSMediaConnectFlowExists(resourceName, &flow),
					resource.TestCheckResourceAttr(resourceName, "source.0.decryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.decryption.0.algorithm", "aes256"),
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

func TestAccAWSMediaConnectFlowConfig_Options(t *testing.T) {
	var flow mediaconnect.Flow

	rName := fmt.Sprintf("tfacctest%s", acctest.RandString(5))
	resourceName := "aws_media_connect_flow.test"
	dataSourceName := "data.aws_availability_zones.available"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSMediaConnectFlowDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSMediaConnectFlowConfig_Options(rName, "Test Description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSMediaConnectFlowExists(resourceName, &flow),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", dataSourceName, "names.0"),
					resource.TestCheckResourceAttr(resourceName, "source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "source.0.description", "Test Description1"),
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

func testAccCheckAWSMediaConnectFlowExists(resourceName string, flow *mediaconnect.Flow) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No  Media Connect Flow ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).mediaconnectconn
		output, err := conn.DescribeFlow(&mediaconnect.DescribeFlowInput{
			FlowArn: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if output == nil || output.Flow == nil {
			return fmt.Errorf("Media Connect Flow (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(output.Flow.FlowArn) != rs.Primary.ID {
			return fmt.Errorf("Media Connect Flow (%s) not found", rs.Primary.ID)
		}

		*flow = *output.Flow

		return nil
	}
}

func testAccCheckAWSMediaConnectFlowDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_media_connect_flow" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).mediaconnectconn

		// Handle eventual consistency
		err := resource.Retry(1*time.Minute, func() *resource.RetryError {
			output, err := conn.DescribeFlow(&mediaconnect.DescribeFlowInput{
				FlowArn: aws.String(rs.Primary.ID),
			})

			if err != nil {
				if isAWSErr(err, mediaconnect.ErrCodeNotFoundException, "") {
					return nil
				}
				return resource.NonRetryableError(err)
			}

			if output != nil && output.Flow != nil && aws.StringValue(output.Flow.FlowArn) == rs.Primary.ID {
				return resource.RetryableError(fmt.Errorf("Media Connect Flow %s still exists", rs.Primary.ID))
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccAWSMediaConnectFlowConfig_IamRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %q
  
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "mediaconnect.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}`, rName)
}

func testAccAWSMediaConnectFlowConfig_SecretsManagerSecret(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = "%s"
}
`, rName)
}

func testAccAWSMediaConnectFlowConfig_Base(rName, protocol, ingressPort string) string {
	return fmt.Sprintf(`
resource "aws_media_connect_flow" "test" {
  name = "%s"

  source {
    name           = "%s"
    protocol       = "%s"
    ingest_port    = %s
    whitelist_cidr = "10.24.34.0/23"
  }
}
`, rName, rName, protocol, ingressPort)
}

func testAccAWSMediaConnectFlowConfig_Decryption(rName string, algorithm string) string {
	return fmt.Sprintf(`
%s

%s

resource "aws_media_connect_flow" "test" {
  name = "%s"

  source {
    name           = "%s"
    protocol       = "zixi-push"
    ingest_port    = 2088
    whitelist_cidr = "10.24.34.0/23"

    decryption {
      key_type   = "static-key"
      algorithm  = "%s"
      role_arn   = "${aws_iam_role.test.arn}"
      secret_arn = "${aws_secretsmanager_secret.test.arn}"
    }
  }
}
`, testAccAWSMediaConnectFlowConfig_IamRole(rName), testAccAWSMediaConnectFlowConfig_SecretsManagerSecret(rName), rName, rName, algorithm)
}

func testAccAWSMediaConnectFlowConfig_Options(rName, description string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_media_connect_flow" "test" {
  name              = "%s"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"

  source {
    name           = "%s"
    description    = "%s"
    protocol       = "rtp"
    ingest_port    = 3010
    whitelist_cidr = "10.24.34.0/23"
  }
}
`, rName, rName, description)
}
