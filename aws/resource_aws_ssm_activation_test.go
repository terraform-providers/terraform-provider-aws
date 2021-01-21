package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSSMActivation_basic(t *testing.T) {
	var ssmActivation ssm.Activation
	name := acctest.RandomWithPrefix("tf-acc")
	tag := acctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_ssm_activation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMActivationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMActivationBasicConfig(name, tag),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMActivationExists(resourceName, &ssmActivation),
					resource.TestCheckResourceAttrSet(resourceName, "activation_code"),
					testAccCheckResourceAttrRfc3339(resourceName, "expiration_date"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", tag)),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"activation_code",
				},
			},
		},
	})
}

func TestAccAWSSSMActivation_update(t *testing.T) {
	var ssmActivation1, ssmActivation2 ssm.Activation
	name := acctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_ssm_activation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMActivationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMActivationBasicConfig(name, "My Activation"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMActivationExists(resourceName, &ssmActivation1),
					resource.TestCheckResourceAttrSet(resourceName, "activation_code"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "My Activation"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"activation_code",
				},
			},
			{
				Config: testAccAWSSSMActivationBasicConfig(name, "Foo"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMActivationExists(resourceName, &ssmActivation2),
					resource.TestCheckResourceAttrSet(resourceName, "activation_code"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Foo"),
					testAccCheckAWSSSMActivationRecreated(t, &ssmActivation1, &ssmActivation2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"activation_code",
				},
			},
		},
	})
}

func TestAccAWSSSMActivation_expirationDate(t *testing.T) {
	var ssmActivation ssm.Activation
	rName := acctest.RandomWithPrefix("tf-acc")
	expirationTime := time.Now().Add(48 * time.Hour).UTC()
	expirationDateS := expirationTime.Format(time.RFC3339)
	resourceName := "aws_ssm_activation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMActivationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMActivationConfig_expirationDate(rName, expirationDateS),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMActivationExists(resourceName, &ssmActivation),
					resource.TestCheckResourceAttr(resourceName, "expiration_date", expirationDateS),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"activation_code",
				},
			},
		},
	})
}

func TestAccAWSSSMActivation_disappears(t *testing.T) {
	var ssmActivation ssm.Activation
	name := acctest.RandomWithPrefix("tf-acc")
	tag := acctest.RandomWithPrefix("tf-acc")
	resourceName := "aws_ssm_activation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMActivationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMActivationBasicConfig(name, tag),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMActivationExists(resourceName, &ssmActivation),
					testAccCheckAWSSSMActivationDisappears(&ssmActivation),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSSMActivationRecreated(t *testing.T, before, after *ssm.Activation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.ActivationId == *after.ActivationId {
			t.Fatalf("expected SSM activation Ids to be different but got %v == %v", before.ActivationId, after.ActivationId)
		}
		return nil
	}
}

func testAccCheckAWSSSMActivationExists(n string, ssmActivation *ssm.Activation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Activation ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ssmconn

		resp, err := conn.DescribeActivations(&ssm.DescribeActivationsInput{
			Filters: []*ssm.DescribeActivationsFilter{
				{
					FilterKey: aws.String("ActivationIds"),
					FilterValues: []*string{
						aws.String(rs.Primary.ID),
					},
				},
			},
			MaxResults: aws.Int64(1),
		})

		if err != nil {
			return fmt.Errorf("Could not describe the activation - %s", err)
		}

		*ssmActivation = *resp.ActivationList[0]

		return nil
	}
}

func testAccCheckAWSSSMActivationDisappears(a *ssm.Activation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ssmconn

		input := &ssm.DeleteActivationInput{ActivationId: a.ActivationId}
		_, err := conn.DeleteActivation(input)
		if err != nil {
			return fmt.Errorf("Error deleting SSM Activation: %s", err)
		}
		return nil
	}
}

func testAccCheckAWSSSMActivationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ssmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_activation" {
			continue
		}

		out, err := conn.DescribeActivations(&ssm.DescribeActivationsInput{
			Filters: []*ssm.DescribeActivationsFilter{
				{
					FilterKey: aws.String("ActivationIds"),
					FilterValues: []*string{
						aws.String(rs.Primary.ID),
					},
				},
			},
			MaxResults: aws.Int64(1),
		})

		if err == nil {
			if len(out.ActivationList) != 0 &&
				*out.ActivationList[0].ActivationId == rs.Primary.ID {
				return fmt.Errorf("SSM Activation still exists")
			}
		}

		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func testAccAWSSSMActivationBasicConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test_role" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ssm.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

}

resource "aws_iam_role_policy_attachment" "test_attach" {
  role       = aws_iam_role.test_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2RoleforSSM"
}
`, rName)
}

func testAccAWSSSMActivationBasicConfig(rName string, rTag string) string {
	return testAccAWSSSMActivationBasicConfigBase(rName) + fmt.Sprintf(`
resource "aws_ssm_activation" "test" {
  name               = %[1]q
  description        = "Test"
  iam_role           = aws_iam_role.test_role.name
  registration_limit = "5"
  depends_on         = [aws_iam_role_policy_attachment.test_attach]

  tags = {
    Name = %[2]q
  }
}
`, rName, rTag)
}

func testAccAWSSSMActivationConfig_expirationDate(rName, expirationDate string) string {
	return testAccAWSSSMActivationBasicConfigBase(rName) + fmt.Sprintf(`
resource "aws_ssm_activation" "test" {
  name               = "test_ssm_activation-%[1]s"
  description        = "Test"
  expiration_date    = "%[2]s"
  iam_role           = aws_iam_role.test_role.name
  registration_limit = "5"
  depends_on         = [aws_iam_role_policy_attachment.test_attach]
}
`, rName, expirationDate)
}
