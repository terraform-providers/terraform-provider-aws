package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/personalize"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"
)

func TestAccPersonalizeDatasetGroup_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_personalize_dataset_group.group"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPersonalizeDatasetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPersonalizeDatasetGroupBasicConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPersonalizeDatasetGroupExists(resourceName),
					testAccCheckResourceAttrRegionalARN(
						resourceName, "arn", "personalize",
						testAccPersonalizeDatasetGroupName("dataset-group/dsg-", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "name", testAccPersonalizeDatasetGroupName("dsg-", rInt)),
					resource.TestCheckNoResourceAttr(resourceName, "kms"),
					resource.TestCheckNoResourceAttr(resourceName, "kms.0.role_arn"),
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

func TestAccPersonalizeDatasetGroup_kms(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_personalize_dataset_group.group"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPersonalizeDatasetGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPersonalizeDatasetGroupKMSConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPersonalizeDatasetGroupExists(resourceName),
					testAccCheckResourceAttrRegionalARN(
						resourceName, "arn", "personalize",
						testAccPersonalizeDatasetGroupName("dataset-group/dsg-", rInt)),
					resource.TestCheckResourceAttr(
						resourceName, "name", testAccPersonalizeDatasetGroupName("dsg-", rInt)),
					resource.TestCheckResourceAttrSet(resourceName, "kms.0.key_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "kms.0.role_arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"kms"},
			},
		},
	})
}

func testAccPersonalizeDatasetGroupBasicConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_personalize_dataset_group" "group" {
  name = "dsg-%d"
}
`, randInt)
}

func testAccPersonalizeDatasetGroupKMSConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_personalize_dataset_group" "group" {
  name = "dsg-%[1]d"

  kms {
    key_arn  = aws_kms_key.key.arn
    role_arn = aws_iam_role.role.arn
  }
}

resource "aws_kms_key" "key" {
}

resource "aws_iam_role" "role" {
  name               = "dsg-assume-%[1]d"
  assume_role_policy = data.aws_iam_policy_document.assume_policy.json
}

data "aws_iam_policy_document" "assume_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["personalize.amazonaws.com"]
    }
  }
}

resource "aws_iam_policy" "policy" {
  name   = "dsg-policy-%[1]d"
  policy = data.aws_iam_policy_document.policy.json
}

resource "aws_iam_role_policy_attachment" "attach" {
  role       = aws_iam_role.role.name
  policy_arn = aws_iam_policy.policy.arn
}

data "aws_iam_policy_document" "policy" {
  statement {

    actions = [
      "kms:*",
    ]

    resources = [
      "${aws_kms_key.key.arn}"
    ]
  }
}
`, randInt)
}

func testAccPersonalizeDatasetGroupName(prefix string, randInt int) string {
	return fmt.Sprintf("%s%d", prefix, randInt)
}

func testAccCheckPersonalizeDatasetGroupExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s\n", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).personalizeconn
		r, err := conn.DescribeDatasetGroup(&personalize.DescribeDatasetGroupInput{
			DatasetGroupArn: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if r == nil {
			return fmt.Errorf("Personalize dataset group not found")
		}

		return nil
	}
}

func testAccCheckPersonalizeDatasetGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).personalizeconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_personalize_dataset_group" {
			continue
		}

		_, err := conn.DescribeDatasetGroup(&personalize.DescribeDatasetGroupInput{
			DatasetGroupArn: aws.String(rs.Primary.ID),
		})

		if isAWSErr(err, personalize.ErrCodeResourceNotFoundException, "") {
			continue
		}

		return fmt.Errorf("Personalize Dataset Group (%s) still exists", rs.Primary.ID)
	}

	return nil
}
