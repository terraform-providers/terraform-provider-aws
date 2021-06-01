package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/msk/finder"
)

func TestAccAwsMskScramSecretAssociation_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_scram_secret_association.test"
	clusterResourceName := "aws_msk_cluster.test"
	secretResourceName := "aws_secretsmanager_secret.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskScramSecretAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskScramSecretAssociation_basic(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskScramSecretAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_arn", clusterResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "secret_arn_list.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secret_arn_list.*", secretResourceName, "arn"),
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

func TestAccAwsMskScramSecretAssociation_update(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_scram_secret_association.test"
	secretResourceName := "aws_secretsmanager_secret.test.0"
	secretResourceName2 := "aws_secretsmanager_secret.test.1"
	secretResourceName3 := "aws_secretsmanager_secret.test.2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskScramSecretAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskScramSecretAssociation_basic(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskScramSecretAssociationExists(resourceName),
				),
			},
			{
				Config: testAccMskScramSecretAssociation_basic(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskScramSecretAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "secret_arn_list.#", "3"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secret_arn_list.*", secretResourceName, "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secret_arn_list.*", secretResourceName2, "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secret_arn_list.*", secretResourceName3, "arn"),
				),
			},
			{
				Config: testAccMskScramSecretAssociation_basic(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskScramSecretAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "secret_arn_list.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secret_arn_list.*", secretResourceName, "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "secret_arn_list.*", secretResourceName2, "arn"),
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

func TestAccAwsMskScramSecretAssociation_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_scram_secret_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskScramSecretAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskScramSecretAssociation_basic(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskScramSecretAssociationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsMskScramSecretAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsMskScramSecretAssociation_disappears_Cluster(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_msk_scram_secret_association.test"
	clusterResourceName := "aws_msk_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMskScramSecretAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMskScramSecretAssociation_basic(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMskScramSecretAssociationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsMskCluster(), clusterResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckMskScramSecretAssociationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_msk_scram_secret_association" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).kafkaconn
		input := &kafka.ListScramSecretsInput{
			ClusterArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.ListScramSecrets(input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, kafka.ErrCodeNotFoundException) {
				continue
			}
			return err
		}
	}
	return nil
}

func testAccCheckMskScramSecretAssociationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set for %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).kafkaconn
		_, err := finder.ScramSecrets(conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccMskScramSecretAssociationBaseConfig(rName string, count int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.5.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = [aws_subnet.example_subnet_az1.id, aws_subnet.example_subnet_az2.id, aws_subnet.example_subnet_az3.id]
    ebs_volume_size = 10
    instance_type   = "kafka.t3.small"
    security_groups = [aws_security_group.example_sg.id]
  }

  client_authentication {
    sasl {
      scram = true
    }
  }
}

resource "aws_kms_key" "test" {
  count       = %[2]d
  description = "%[1]s-${count.index + 1}"
}

resource "aws_secretsmanager_secret" "test" {
  count      = %[2]d
  name       = "AmazonMSK_%[1]s-${count.index + 1}"
  kms_key_id = aws_kms_key.test[count.index].id
}

resource "aws_secretsmanager_secret_version" "test" {
  count         = %[2]d
  secret_id     = aws_secretsmanager_secret.test[count.index].id
  secret_string = jsonencode({ username = "user", password = "pass" })
}

resource "aws_secretsmanager_secret_policy" "test" {
  count      = %[2]d
  secret_arn = aws_secretsmanager_secret.test[count.index].arn
  policy     = <<POLICY
{
  "Version" : "2012-10-17",
  "Statement" : [ {
    "Sid": "AWSKafkaResourcePolicy",
    "Effect" : "Allow",
    "Principal" : {
      "Service" : "kafka.${data.aws_partition.current.dns_suffix}"
    },
    "Action" : "secretsmanager:getSecretValue",
    "Resource" : "${aws_secretsmanager_secret.test[count.index].arn}"
  } ]
}
POLICY
}
`, rName, count)
}

func testAccMskScramSecretAssociation_basic(rName string, count int) string {
	return composeConfig(
		testAccMskClusterBaseConfig(),
		testAccMskScramSecretAssociationBaseConfig(rName, count), `
resource "aws_msk_scram_secret_association" "test" {
  cluster_arn     = aws_msk_cluster.test.arn
  secret_arn_list = aws_secretsmanager_secret.test[*].arn

  depends_on = [aws_secretsmanager_secret_version.test]
}
`)
}
