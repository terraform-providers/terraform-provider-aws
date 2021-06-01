package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarnotifications"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCodeStarNotificationsNotificationRule_basic(t *testing.T) {
	resourceName := "aws_codestarnotifications_notification_rule.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codestarnotifications.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeStarNotificationsNotificationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeStarNotificationsNotificationRuleConfigBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "codestar-notifications", regexp.MustCompile("notificationrule/.+")),
					resource.TestCheckResourceAttr(resourceName, "detail_type", codestarnotifications.DetailTypeBasic),
					resource.TestCheckResourceAttr(resourceName, "event_type_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "status", codestarnotifications.NotificationRuleStatusEnabled),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "target.#", "1"),
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

func TestAccAWSCodeStarNotificationsNotificationRule_Status(t *testing.T) {
	resourceName := "aws_codestarnotifications_notification_rule.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codestarnotifications.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeStarNotificationsNotificationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeStarNotificationsNotificationRuleConfigStatus(rName, codestarnotifications.NotificationRuleStatusDisabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "status", codestarnotifications.NotificationRuleStatusDisabled),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeStarNotificationsNotificationRuleConfigStatus(rName, codestarnotifications.NotificationRuleStatusEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "status", codestarnotifications.NotificationRuleStatusEnabled),
				),
			},
			{
				Config: testAccAWSCodeStarNotificationsNotificationRuleConfigStatus(rName, codestarnotifications.NotificationRuleStatusDisabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "status", codestarnotifications.NotificationRuleStatusDisabled),
				),
			},
		},
	})
}

func TestAccAWSCodeStarNotificationsNotificationRule_Targets(t *testing.T) {
	resourceName := "aws_codestarnotifications_notification_rule.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codestarnotifications.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeStarNotificationsNotificationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeStarNotificationsNotificationRuleConfigTargets1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "target.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeStarNotificationsNotificationRuleConfigTargets2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "target.#", "2"),
				),
			},
			{
				Config: testAccAWSCodeStarNotificationsNotificationRuleConfigTargets1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "target.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSCodeStarNotificationsNotificationRule_Tags(t *testing.T) {
	resourceName := "aws_codestarnotifications_notification_rule.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codestarnotifications.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeStarNotificationsNotificationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeStarNotificationsNotificationRuleConfigTags1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.TestTag1", "123456"),
					resource.TestCheckResourceAttr(resourceName, "tags.TestTag2", "654321"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeStarNotificationsNotificationRuleConfigTags2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.TestTag2", "654321"),
					resource.TestCheckResourceAttr(resourceName, "tags.TestTag3", "asdfgh"),
				),
			},
			{
				Config: testAccAWSCodeStarNotificationsNotificationRuleConfigTags1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.TestTag1", "123456"),
					resource.TestCheckResourceAttr(resourceName, "tags.TestTag2", "654321"),
				),
			},
		},
	})
}

func TestAccAWSCodeStarNotificationsNotificationRule_EventTypeIds(t *testing.T) {
	resourceName := "aws_codestarnotifications_notification_rule.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codestarnotifications.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeStarNotificationsNotificationRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeStarNotificationsNotificationRuleConfigEventTypeIds1(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_type_ids.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeStarNotificationsNotificationRuleConfigEventTypeIds2(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_type_ids.#", "2"),
				),
			},
			{
				Config: testAccAWSCodeStarNotificationsNotificationRuleConfigEventTypeIds3(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "event_type_ids.#", "1"),
				),
			},
		},
	})
}

func testAccCheckAWSCodeStarNotificationsNotificationRuleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codestarnotificationsconn

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "aws_codestarnotifications_notification_rule":
			_, err := conn.DescribeNotificationRule(&codestarnotifications.DescribeNotificationRuleInput{
				Arn: aws.String(rs.Primary.ID),
			})

			if err != nil && !isAWSErr(err, codestarnotifications.ErrCodeResourceNotFoundException, "") {
				return err
			}
		case "aws_sns_topic":
			res, err := conn.ListTargets(&codestarnotifications.ListTargetsInput{
				Filters: []*codestarnotifications.ListTargetsFilter{
					{
						Name:  aws.String("TARGET_ADDRESS"),
						Value: aws.String(rs.Primary.ID),
					},
					{
						Name:  aws.String("TARGET_TYPE"),
						Value: aws.String("SNS"),
					},
				},
				MaxResults: aws.Int64(1),
			})
			if err != nil {
				return err
			}
			if len(res.Targets) > 0 {
				return fmt.Errorf("codestar notification target (%s) is not removed", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccAWSCodeStarNotificationsNotificationRuleConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_codecommit_repository" "test" {
  repository_name = %[1]q
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}
`, rName)
}

func testAccAWSCodeStarNotificationsNotificationRuleConfigBasic(rName string) string {
	return testAccAWSCodeStarNotificationsNotificationRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = %[1]q
  resource       = aws_codecommit_repository.test.arn
  status         = "ENABLED"

  tags = {
    TestTag = "123456"
  }

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName)
}

func testAccAWSCodeStarNotificationsNotificationRuleConfigStatus(rName, status string) string {
	return testAccAWSCodeStarNotificationsNotificationRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = %[1]q
  resource       = aws_codecommit_repository.test.arn
  status         = %[2]q

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName, status)
}

func testAccAWSCodeStarNotificationsNotificationRuleConfigTargets1(rName string) string {
	return testAccAWSCodeStarNotificationsNotificationRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = %[1]q
  resource       = aws_codecommit_repository.test.arn

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName)
}

func testAccAWSCodeStarNotificationsNotificationRuleConfigTargets2(rName string) string {
	return testAccAWSCodeStarNotificationsNotificationRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_sns_topic" "test2" {
  name = "%[1]s2"
}

resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = %[1]q
  resource       = aws_codecommit_repository.test.arn

  target {
    address = aws_sns_topic.test.arn
  }

  target {
    address = aws_sns_topic.test2.arn
  }
}
`, rName)
}

func testAccAWSCodeStarNotificationsNotificationRuleConfigTags1(rName string) string {
	return testAccAWSCodeStarNotificationsNotificationRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = %[1]q
  resource       = aws_codecommit_repository.test.arn
  status         = "ENABLED"

  tags = {
    TestTag1 = "123456"
    TestTag2 = "654321"
  }

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName)
}

func testAccAWSCodeStarNotificationsNotificationRuleConfigTags2(rName string) string {
	return testAccAWSCodeStarNotificationsNotificationRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type    = "BASIC"
  event_type_ids = ["codecommit-repository-comments-on-commits"]
  name           = %[1]q
  resource       = aws_codecommit_repository.test.arn
  status         = "ENABLED"

  tags = {
    TestTag2 = "654321"
    TestTag3 = "asdfgh"
  }

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName)
}

func testAccAWSCodeStarNotificationsNotificationRuleConfigEventTypeIds1(rName string) string {
	return testAccAWSCodeStarNotificationsNotificationRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type = "BASIC"
  event_type_ids = [
    "codecommit-repository-comments-on-commits",
  ]
  name     = %[1]q
  resource = aws_codecommit_repository.test.arn
  status   = "ENABLED"

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName)
}

func testAccAWSCodeStarNotificationsNotificationRuleConfigEventTypeIds2(rName string) string {
	return testAccAWSCodeStarNotificationsNotificationRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type = "BASIC"
  event_type_ids = [
    "codecommit-repository-comments-on-commits",
    "codecommit-repository-pull-request-created",
  ]
  name     = %[1]q
  resource = aws_codecommit_repository.test.arn
  status   = "ENABLED"

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName)
}

func testAccAWSCodeStarNotificationsNotificationRuleConfigEventTypeIds3(rName string) string {
	return testAccAWSCodeStarNotificationsNotificationRuleConfigBase(rName) + fmt.Sprintf(`
resource "aws_codestarnotifications_notification_rule" "test" {
  detail_type = "BASIC"
  event_type_ids = [
    "codecommit-repository-pull-request-created",
  ]
  name     = %[1]q
  resource = aws_codecommit_repository.test.arn
  status   = "ENABLED"

  target {
    address = aws_sns_topic.test.arn
  }
}
`, rName)
}
