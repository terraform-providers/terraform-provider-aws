package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSGreengrassSubscriptionDefinition_basic(t *testing.T) {
	rString := acctest.RandString(8)
	resourceName := "aws_greengrass_subscription_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGreengrassSubscriptionDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGreengrassSubscriptionDefinitionConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("subscription_definition_%s", rString)),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr("aws_greengrass_subscription_definition.test", "tags.tagKey", "tagValue"),
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

func TestAccAWSGreengrassSubscriptionDefinition_DefinitionVersion(t *testing.T) {
	rString := acctest.RandString(8)
	resourceName := "aws_greengrass_subscription_definition.test"

	subscription := map[string]interface{}{
		"id":      "test_id",
		"subject": "test_subject",
		"source":  "arn:aws:iot:eu-west-1:111111111111:thing/Source",
		"target":  "arn:aws:iot:eu-west-1:222222222222:thing/Target",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGreengrassSubscriptionDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGreengrassSubscriptionDefinitionConfig_definitionVersion(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("subscription_definition_%s", rString)),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					testAccCheckGreengrassSubscription_checkSubscription(resourceName, subscription),
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

func testAccCheckGreengrassSubscription_checkSubscription(n string, expectedSubscription map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Greengrass Subscription Definition ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).greengrassconn

		getSubscriptionInput := &greengrass.GetSubscriptionDefinitionInput{
			SubscriptionDefinitionId: aws.String(rs.Primary.ID),
		}
		definitionOut, err := conn.GetSubscriptionDefinition(getSubscriptionInput)

		if err != nil {
			return err
		}

		getVersionInput := &greengrass.GetSubscriptionDefinitionVersionInput{
			SubscriptionDefinitionId:        aws.String(rs.Primary.ID),
			SubscriptionDefinitionVersionId: definitionOut.LatestVersion,
		}
		versionOut, err := conn.GetSubscriptionDefinitionVersion(getVersionInput)
		if err != nil {
			return err
		}
		subscription := versionOut.Definition.Subscriptions[0]

		expectedSubscriptionId := expectedSubscription["id"].(string)
		if *subscription.Id != expectedSubscriptionId {
			return fmt.Errorf("Subscription Id %s is not equal expected %s", *subscription.Id, expectedSubscriptionId)
		}

		expectedSubscriptionSubject := expectedSubscription["subject"].(string)
		if *subscription.Subject != expectedSubscriptionSubject {
			return fmt.Errorf("Subscription Subject %s is not equal expected %s", *subscription.Subject, expectedSubscriptionSubject)
		}

		expectedSubscriptionSource := expectedSubscription["source"].(string)
		if *subscription.Source != expectedSubscriptionSource {
			return fmt.Errorf("Subscription Source %s is not equal expected %s", *subscription.Source, expectedSubscriptionSource)
		}

		expectedSubscriptionTarget := expectedSubscription["target"].(string)
		if *subscription.Target != expectedSubscriptionTarget {
			return fmt.Errorf("Subscription Target %s is not equal expected %s", *subscription.Target, expectedSubscriptionTarget)
		}
		return nil
	}
}

func testAccCheckAWSGreengrassSubscriptionDefinitionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).greengrassconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_greengrass_subscription_definition" {
			continue
		}

		params := &greengrass.ListSubscriptionDefinitionsInput{
			MaxResults: aws.String("20"),
		}

		out, err := conn.ListSubscriptionDefinitions(params)
		if err != nil {
			return err
		}
		for _, definition := range out.Definitions {
			if *definition.Id == rs.Primary.ID {
				return fmt.Errorf("Expected Greengrass Subscription Definition to be destroyed, %s found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccAWSGreengrassSubscriptionDefinitionConfig_basic(rString string) string {
	return fmt.Sprintf(`
resource "aws_greengrass_subscription_definition" "test" {
  name = "subscription_definition_%s"

  tags = {
	"tagKey" = "tagValue"
  } 
}
`, rString)
}

func testAccAWSGreengrassSubscriptionDefinitionConfig_definitionVersion(rString string) string {
	return fmt.Sprintf(`
resource "aws_greengrass_subscription_definition" "test" {
	name = "subscription_definition_%[1]s"
	subscription_definition_version {
		subscription {
			id = "test_id"
			subject = "test_subject"
			source = "arn:aws:iot:eu-west-1:111111111111:thing/Source"
			target = "arn:aws:iot:eu-west-1:222222222222:thing/Target"	
		}
	}
}
`, rString)
}
