package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSDmsEventSubscription_basic(t *testing.T) {
	var eventSubscription dms.EventSubscription
	resourceName := "aws_dms_event_subscription.test"
	snsTopicResourceName := "aws_sns_topic.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDmsEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDmsEventSubscriptionConfigEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDmsEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "replication-instance"),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "creation"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "failure"),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "sns_topic_arn", snsTopicResourceName, "arn"),
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

func TestAccAWSDmsEventSubscription_disappears(t *testing.T) {
	var eventSubscription dms.EventSubscription
	resourceName := "aws_dms_event_subscription.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDmsEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDmsEventSubscriptionConfigEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDmsEventSubscriptionExists(resourceName, &eventSubscription),
					testAccCheckDmsEventSubscriptionDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDmsEventSubscription_Enabled(t *testing.T) {
	var eventSubscription dms.EventSubscription
	resourceName := "aws_dms_event_subscription.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDmsEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDmsEventSubscriptionConfigEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDmsEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDmsEventSubscriptionConfigEnabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDmsEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			{
				Config: testAccDmsEventSubscriptionConfigEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDmsEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
		},
	})
}

func TestAccAWSDmsEventSubscription_EventCategories(t *testing.T) {
	var eventSubscription dms.EventSubscription
	resourceName := "aws_dms_event_subscription.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDmsEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDmsEventSubscriptionConfigEventCategories2(rName, "creation", "failure"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDmsEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "creation"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "failure"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDmsEventSubscriptionConfigEventCategories2(rName, "configuration change", "deletion"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDmsEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "configuration change"),
					resource.TestCheckTypeSetElemAttr(resourceName, "event_categories.*", "deletion"),
				),
			},
		},
	})
}

func TestAccAWSDmsEventSubscription_Tags(t *testing.T) {
	var eventSubscription dms.EventSubscription
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dms_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDmsEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDmsEventSubscriptionConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDmsEventSubscriptionExists(resourceName, &eventSubscription),
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
				Config: testAccDmsEventSubscriptionConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDmsEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDmsEventSubscriptionConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDmsEventSubscriptionExists(resourceName, &eventSubscription),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckDmsEventSubscriptionDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dms_event_subscription" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).dmsconn

		resp, err := conn.DescribeEventSubscriptions(&dms.DescribeEventSubscriptionsInput{
			SubscriptionName: aws.String(rs.Primary.ID),
		})

		if isAWSErr(err, dms.ErrCodeResourceNotFoundFault, "") {
			continue
		}

		if err != nil {
			return err
		}

		if resp != nil && len(resp.EventSubscriptionsList) > 0 {
			return fmt.Errorf("DMS event subscription still exists: %s", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckDmsEventSubscriptionDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resource := resourceAwsDmsEventSubscription()

		return resource.Delete(resource.Data(rs.Primary), testAccProvider.Meta())
	}
}

func testAccCheckDmsEventSubscriptionExists(n string, eventSubscription *dms.EventSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).dmsconn
		resp, err := conn.DescribeEventSubscriptions(&dms.DescribeEventSubscriptionsInput{
			SubscriptionName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return fmt.Errorf("DMS event subscription error: %v", err)
		}

		if resp == nil || len(resp.EventSubscriptionsList) == 0 || resp.EventSubscriptionsList[0] == nil {
			return fmt.Errorf("DMS event subscription not found")
		}

		*eventSubscription = *resp.EventSubscriptionsList[0]

		return nil
	}
}

func testAccDmsEventSubscriptionConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_partition" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.1.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = aws_vpc.test.tags["Name"]
  }
}

resource "aws_dms_replication_subnet_group" "test" {
  replication_subnet_group_description = %[1]q
  replication_subnet_group_id          = %[1]q
  subnet_ids                           = aws_subnet.test[*].id
}

resource "aws_dms_replication_instance" "test" {
  apply_immediately           = true
  replication_instance_class  = data.aws_partition.current.partition == "aws" ? "dms.t2.micro" : "dms.c4.large"
  replication_instance_id     = %[1]q
  replication_subnet_group_id = aws_dms_replication_subnet_group.test.replication_subnet_group_id
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}
`, rName)
}

func testAccDmsEventSubscriptionConfigEnabled(rName string, enabled bool) string {
	return composeConfig(
		testAccDmsEventSubscriptionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_dms_event_subscription" "test" {
  name             = %[1]q
  enabled          = %[2]t
  event_categories = ["creation", "failure"]
  source_type      = "replication-instance"
  source_ids       = [aws_dms_replication_instance.test.replication_instance_id]
  sns_topic_arn    = aws_sns_topic.test.arn
}
`, rName, enabled))
}

func testAccDmsEventSubscriptionConfigEventCategories2(rName string, eventCategory1 string, eventCategory2 string) string {
	return composeConfig(
		testAccDmsEventSubscriptionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_dms_event_subscription" "test" {
  name             = %[1]q
  enabled          = false
  event_categories = [%[2]q, %[3]q]
  source_type      = "replication-instance"
  source_ids       = [aws_dms_replication_instance.test.replication_instance_id]
  sns_topic_arn    = aws_sns_topic.test.arn
}
`, rName, eventCategory1, eventCategory2))
}

func testAccDmsEventSubscriptionConfigTags1(rName, tagKey1, tagValue1 string) string {
	return composeConfig(
		testAccDmsEventSubscriptionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_dms_event_subscription" "test" {
  name             = %[1]q
  enabled          = true
  event_categories = ["creation", "failure"]
  source_type      = "replication-instance"
  source_ids       = [aws_dms_replication_instance.test.replication_instance_id]
  sns_topic_arn    = aws_sns_topic.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccDmsEventSubscriptionConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(
		testAccDmsEventSubscriptionConfigBase(rName),
		fmt.Sprintf(`
resource "aws_dms_event_subscription" "test" {
  name             = %[1]q
  enabled          = true
  event_categories = ["creation", "failure"]
  source_type      = "replication-instance"
  source_ids       = [aws_dms_replication_instance.test.replication_instance_id]
  sns_topic_arn    = aws_sns_topic.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
