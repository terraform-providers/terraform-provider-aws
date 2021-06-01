package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dlm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSDlmLifecyclePolicy_basic(t *testing.T) {
	resourceName := "aws_dlm_lifecycle_policy.basic"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDlm(t) },
		Providers:    testAccProviders,
		CheckDestroy: dlmLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: dlmLifecyclePolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "dlm", regexp.MustCompile(`policy/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", "tf-acc-basic"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.resource_types.0", "VOLUME"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.name", "tf-acc-basic"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval", "12"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval_unit", "HOURS"),
					resource.TestCheckResourceAttrSet(resourceName, "policy_details.0.schedule.0.create_rule.0.times.0"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.retain_rule.0.count", "10"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.target_tags.tf-acc-test", "basic"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSDlmLifecyclePolicy_Full(t *testing.T) {
	resourceName := "aws_dlm_lifecycle_policy.full"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDlm(t) },
		Providers:    testAccProviders,
		CheckDestroy: dlmLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: dlmLifecyclePolicyFullConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "tf-acc-full"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.resource_types.0", "VOLUME"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.name", "tf-acc-full"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval", "12"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval_unit", "HOURS"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.times.0", "21:42"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.retain_rule.0.count", "10"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.tags_to_add.tf-acc-test-added", "full"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.copy_tags", "false"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.target_tags.tf-acc-test", "full"),
				),
			},
			{
				Config: dlmLifecyclePolicyFullUpdateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "tf-acc-full-updated"),
					resource.TestCheckResourceAttrSet(resourceName, "execution_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "state", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.resource_types.0", "VOLUME"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.name", "tf-acc-full-updated"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval", "24"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.interval_unit", "HOURS"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.create_rule.0.times.0", "09:42"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.retain_rule.0.count", "100"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.tags_to_add.tf-acc-test-added", "full-updated"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.schedule.0.copy_tags", "true"),
					resource.TestCheckResourceAttr(resourceName, "policy_details.0.target_tags.tf-acc-test", "full-updated"),
				),
			},
		},
	})
}

func TestAccAWSDlmLifecyclePolicy_Tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dlm_lifecycle_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDlm(t) },
		Providers:    testAccProviders,
		CheckDestroy: dlmLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: dlmLifecyclePolicyConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
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
				Config: dlmLifecyclePolicyConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: dlmLifecyclePolicyConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					checkDlmLifecyclePolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func dlmLifecyclePolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dlmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dlm_lifecycle_policy" {
			continue
		}

		input := dlm.GetLifecyclePolicyInput{
			PolicyId: aws.String(rs.Primary.ID),
		}

		out, err := conn.GetLifecyclePolicy(&input)

		if isAWSErr(err, dlm.ErrCodeResourceNotFoundException, "") {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error getting DLM Lifecycle Policy (%s): %s", rs.Primary.ID, err)
		}

		if out.Policy != nil {
			return fmt.Errorf("DLM lifecycle policy still exists: %#v", out)
		}
	}

	return nil
}

func checkDlmLifecyclePolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).dlmconn

		input := dlm.GetLifecyclePolicyInput{
			PolicyId: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetLifecyclePolicy(&input)

		if err != nil {
			return fmt.Errorf("error getting DLM Lifecycle Policy (%s): %s", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccPreCheckAWSDlm(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).dlmconn

	input := &dlm.GetLifecyclePoliciesInput{}

	_, err := conn.GetLifecyclePolicies(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func dlmLifecyclePolicyBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "dlm_lifecycle_role" {
  name = %q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "dlm.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_dlm_lifecycle_policy" "basic" {
  description        = "tf-acc-basic"
  execution_role_arn = aws_iam_role.dlm_lifecycle_role.arn

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "tf-acc-basic"

      create_rule {
        interval = 12
      }

      retain_rule {
        count = 10
      }
    }

    target_tags = {
      tf-acc-test = "basic"
    }
  }
}
`, rName)
}

func dlmLifecyclePolicyFullConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "dlm_lifecycle_role" {
  name = %q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "dlm.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_dlm_lifecycle_policy" "full" {
  description        = "tf-acc-full"
  execution_role_arn = aws_iam_role.dlm_lifecycle_role.arn
  state              = "ENABLED"

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "tf-acc-full"

      create_rule {
        interval      = 12
        interval_unit = "HOURS"
        times         = ["21:42"]
      }

      retain_rule {
        count = 10
      }

      tags_to_add = {
        tf-acc-test-added = "full"
      }

      copy_tags = false
    }

    target_tags = {
      tf-acc-test = "full"
    }
  }
}
`, rName)
}

func dlmLifecyclePolicyFullUpdateConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "dlm_lifecycle_role" {
  name = %q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "dlm.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_dlm_lifecycle_policy" "full" {
  description        = "tf-acc-full-updated"
  execution_role_arn = "${aws_iam_role.dlm_lifecycle_role.arn}-doesnt-exist"
  state              = "DISABLED"

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "tf-acc-full-updated"

      create_rule {
        interval      = 24
        interval_unit = "HOURS"
        times         = ["09:42"]
      }

      retain_rule {
        count = 100
      }

      tags_to_add = {
        tf-acc-test-added = "full-updated"
      }

      copy_tags = true
    }

    target_tags = {
      tf-acc-test = "full-updated"
    }
  }
}
`, rName)
}

func dlmLifecyclePolicyConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "dlm.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_dlm_lifecycle_policy" "test" {
  description        = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "test"

      create_rule {
        interval = 12
      }

      retain_rule {
        count = 10
      }
    }

    target_tags = {
      test = "true"
    }
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func dlmLifecyclePolicyConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "dlm.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_dlm_lifecycle_policy" "test" {
  description        = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  policy_details {
    resource_types = ["VOLUME"]

    schedule {
      name = "test"

      create_rule {
        interval = 12
      }

      retain_rule {
        count = 10
      }
    }

    target_tags = {
      test = "true"
    }
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
