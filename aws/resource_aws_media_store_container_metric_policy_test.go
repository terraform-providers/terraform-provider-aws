package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSMediaStoreContainerMetricPolicy_basic(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_media_store_container_metric_policy.test"
	containerResourceName := "aws_media_store_container.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMediaStore(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaStoreContainerMetricPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaStoreContainerMetricPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaStoreContainerMetricPolicyExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "container_name", containerResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "metric_policy.0.container_level_metrics", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "metric_policy.0.metric_policy_rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_policy.0.metric_policy_rule.0.object_group", "baseball/saturday"),
					resource.TestCheckResourceAttr(resourceName, "metric_policy.0.metric_policy_rule.0.object_group_name", "baseballGroup"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMediaStoreContainerMetricPolicyConfigUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaStoreContainerMetricPolicyExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "metric_policy.0.metric_policy_rule.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "metric_policy.0.metric_policy_rule.0.object_group", "baseball/sunday"),
					resource.TestCheckResourceAttr(resourceName, "metric_policy.0.metric_policy_rule.0.object_group_name", "baseballGroup"),
					resource.TestCheckResourceAttr(resourceName, "metric_policy.0.metric_policy_rule.1.object_group", "football/sunday"),
					resource.TestCheckResourceAttr(resourceName, "metric_policy.0.metric_policy_rule.1.object_group_name", "footballGroup"),
				),
			},
		},
	})
}

func testAccCheckAwsMediaStoreContainerMetricPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).mediastoreconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_media_store_container_metric_policy" {
			continue
		}

		input := &mediastore.GetMetricPolicyInput{
			ContainerName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetMetricPolicy(input)
		if err != nil {
			if isAWSErr(err, mediastore.ErrCodeContainerNotFoundException, "") {
				return nil
			}
			if isAWSErr(err, mediastore.ErrCodePolicyNotFoundException, "") {
				return nil
			}
			if isAWSErr(err, mediastore.ErrCodeContainerInUseException, "Container must be ACTIVE in order to perform this operation") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected MediaStore Container Metric Policy to be destroyed, %s found", rs.Primary.ID)
	}
	return nil
}

func testAccCheckAwsMediaStoreContainerMetricPolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).mediastoreconn

		input := &mediastore.GetMetricPolicyInput{
			ContainerName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetMetricPolicy(input)

		return err
	}
}

func testAccMediaStoreContainerMetricPolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_store_container" "test" {
  name = "tf_mediastore_%s"
}

resource "aws_media_store_container_metric_policy" "test" {
  container_name = "${aws_media_store_container.test.name}"

  metric_policy {
	container_level_metrics = "DISABLED"

	metric_policy_rule {
		object_group      = "baseball/saturday"
		object_group_name = "baseballGroup"
	}
  }
}
`, rName)
}

func testAccMediaStoreContainerMetricPolicyConfigUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_store_container" "test" {
  name = "tf_mediastore_%s"
}

resource "aws_media_store_container_metric_policy" "test" {
  container_name = "${aws_media_store_container.test.name}"

  metric_policy {
	container_level_metrics = "DISABLED"

	metric_policy_rule {
		object_group      = "baseball/sunday"
		object_group_name = "baseballGroup"
	}

	metric_policy_rule {
		object_group      = "football/sunday"
		object_group_name = "footballGroup"
	}
  }
}
`, rName)
}
