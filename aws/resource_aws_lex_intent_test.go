package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsLexIntent_basic(t *testing.T) {
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexIntentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexIntentConfig_basic(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexIntentExists(rName, &v),
					testAccCheckAwsLexIntentNotExists(testIntentID, "1"),

					resource.TestCheckResourceAttrSet(rName, "arn"),
					resource.TestCheckResourceAttrSet(rName, "checksum"),
					resource.TestCheckNoResourceAttr(rName, "conclusion_statement"),
					resource.TestCheckNoResourceAttr(rName, "confirmation_prompt"),
					resource.TestCheckResourceAttr(rName, "create_version", "false"),
					testAccCheckResourceAttrRfc3339(rName, "created_date"),
					resource.TestCheckResourceAttr(rName, "description", ""),
					resource.TestCheckNoResourceAttr(rName, "dialog_code_hook"),
					resource.TestCheckNoResourceAttr(rName, "follow_up_prompt"),
					resource.TestCheckNoResourceAttr(rName, "fulfillment_activity"),
					testAccCheckResourceAttrRfc3339(rName, "last_updated_date"),
					resource.TestCheckResourceAttr(rName, "name", testIntentID),
					resource.TestCheckNoResourceAttr(rName, "parent_intent_signature"),
					resource.TestCheckNoResourceAttr(rName, "rejection_statement"),
					resource.TestCheckNoResourceAttr(rName, "sample_utterances"),
					resource.TestCheckNoResourceAttr(rName, "slot"),
					resource.TestCheckResourceAttr(rName, "version", LexIntentVersionLatest),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexIntent_createVersion(t *testing.T) {
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexIntentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexIntentConfig_basic(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexIntentExists(rName, &v),
					testAccCheckAwsLexIntentNotExists(testIntentID, "1"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
			{
				Config: testAccAwsLexIntentConfig_createVersion(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexIntentExists(rName, &v),
					testAccCheckAwsLexIntentExistsWithVersion(rName, "1", &v),
					resource.TestCheckResourceAttr(rName, "version", "1"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexIntent_conclusionStatement(t *testing.T) {
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexIntentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexIntentConfig_conclusionStatement(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexIntentExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.#", "1"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.#", "1"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.0.content", "Your order for {FlowerType} has been placed and will be ready by {PickupTime} on {PickupDate}"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.0.content_type", "PlainText"),
					resource.TestCheckNoResourceAttr(rName, "conclusion_statement.0.message.0.group_number"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.response_card", "Your order for {FlowerType} has been placed and will be ready by {PickupTime} on {PickupDate}"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
			{
				Config: testAccAwsLexIntentConfig_conclusionStatementUpdate(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexIntentExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.#", "2"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.0.content", "Your order for {FlowerType} has been placed and will be ready by {PickupTime} on {PickupDate}"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.0.group_number", "1"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.1.content", "Your order for {FlowerType} has been placed"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.1.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.1.group_number", "1"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.response_card", "Your order for {FlowerType} has been placed"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexIntent_confirmationPromptAndRejectionStatement(t *testing.T) {
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexIntentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexIntentConfig_confirmationPromptAndRejectionStatement(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexIntentExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.#", "1"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.max_attempts", "1"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.message.#", "1"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.message.0.content", "Okay, your {FlowerType} will be ready for pickup by {PickupTime} on {PickupDate}. Does this sound okay?"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.response_card", "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"Okay, your {FlowerType} will be ready for pickup by {PickupTime} on {PickupDate}. Does this sound okay?\",\"buttons\":[{\"text\":\"Yes\",\"value\":\"yes\"},{\"text\":\"No\",\"value\":\"no\"}]}]}"),
					resource.TestCheckResourceAttr(rName, "rejection_statement.#", "1"),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.message.#", "1"),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.message.0.content", "Okay, I will not place your order."),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.response_card", "Okay, I will not place your order."),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
			{
				Config: testAccAwsLexIntentConfig_confirmationPromptAndRejectionStatementUpdate(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexIntentExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.max_attempts", "2"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.message.#", "2"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.message.0.content", "Okay, your {FlowerType} will be ready for pickup by {PickupTime} on {PickupDate}. Does this sound okay?"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.message.1.content", "Okay, your {FlowerType} will be ready for pickup on {PickupDate}. Does this sound okay?"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.message.1.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.response_card", "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"Okay, your {FlowerType} will be ready for pickup on {PickupDate}. Does this sound okay?\",\"buttons\":[{\"text\":\"Yes\",\"value\":\"yes\"},{\"text\":\"No\",\"value\":\"no\"}]}]}"),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.message.#", "2"),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.message.0.content", "Okay, I will not place your order."),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.message.1.content", "Okay, your order has been cancelled."),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.message.1.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.response_card", "Okay, your order has been cancelled."),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexIntent_dialogCodeHook(t *testing.T) {
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexIntentDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAwsLexIntentConfig_lambda(testIntentID),
					testAccAwsLexIntentConfig_dialogCodeHook(testIntentID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexIntentExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "dialog_code_hook.#", "1"),
					resource.TestCheckResourceAttr(rName, "dialog_code_hook.0.message_version", "1"),
					resource.TestCheckResourceAttrSet(rName, "dialog_code_hook.0.uri"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexIntent_followUpPrompt(t *testing.T) {
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexIntentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexIntentConfig_followUpPrompt(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexIntentExists(rName, &v),

					resource.TestCheckResourceAttr(rName, "follow_up_prompt.#", "1"),

					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.#", "1"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.max_attempts", "1"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.message.#", "1"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.message.0.content", "Would you like to order more flowers?"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.response_card", "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"Would you like to order more flowers?\",\"buttons\":[{\"text\":\"Yes\",\"value\":\"yes\"},{\"text\":\"No\",\"value\":\"no\"}]}]}"),

					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.#", "1"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.message.#", "1"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.message.0.content", "Okay, no additional flowers will be ordered."),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.response_card", "Okay, no additional flowers will be ordered."),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
			{
				Config: testAccAwsLexIntentConfig_followUpPromptUpdate(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexIntentExists(rName, &v),

					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.max_attempts", "2"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.message.#", "2"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.message.0.content", "Would you like to order more flowers?"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.message.1.content", "Would you like to start another order?"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.message.1.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.response_card", "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"Would you like to start another order?\",\"buttons\":[{\"text\":\"Yes\",\"value\":\"yes\"},{\"text\":\"No\",\"value\":\"no\"}]}]}"),

					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.message.#", "2"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.message.0.content", "Okay, additional flowers will be ordered."),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.message.1.content", "Okay, no additional flowers will be ordered."),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.message.1.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.response_card", "Okay, additional flowers will be ordered."),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexIntent_fulfillmentActivity(t *testing.T) {
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexIntentDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAwsLexIntentConfig_lambda(testIntentID),
					testAccAwsLexIntentConfig_fulfillmentActivity(testIntentID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexIntentExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "fulfillment_activity.#", "1"),
					resource.TestCheckResourceAttr(rName, "fulfillment_activity.0.code_hook.#", "1"),
					resource.TestCheckResourceAttr(rName, "fulfillment_activity.0.code_hook.0.message_version", "1"),
					resource.TestCheckResourceAttrSet(rName, "fulfillment_activity.0.code_hook.0.uri"),
					resource.TestCheckResourceAttr(rName, "fulfillment_activity.0.type", "CodeHook"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexIntent_sampleUtterances(t *testing.T) {
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexIntentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexIntentConfig_sampleUtterances(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexIntentExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "sample_utterances.#", "1"),
					resource.TestCheckResourceAttr(rName, "sample_utterances.0", "I would like to pick up flowers"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
			{
				Config: testAccAwsLexIntentConfig_sampleUtterancesUpdate(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexIntentExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "sample_utterances.#", "2"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexIntent_slots(t *testing.T) {
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexIntentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexIntentConfig_slots(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexIntentExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "slot.#", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.description", "The date to pick up the flowers"),
					resource.TestCheckResourceAttr(rName, "slot.0.name", "PickupDate"),
					resource.TestCheckResourceAttr(rName, "slot.0.priority", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.sample_utterances.#", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.sample_utterances.0", "I would like to order {FlowerType}"),
					resource.TestCheckResourceAttr(rName, "slot.0.slot_constraint", "Required"),
					resource.TestCheckResourceAttr(rName, "slot.0.slot_type", "AMAZON.DATE"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.#", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.0.max_attempts", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.0.message.#", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.0.message.0.content", "What day do you want the {FlowerType} to be picked up?"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.0.message.0.content_type", "PlainText"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
			{
				Config: testAccAwsLexIntentConfig_slotsUpdate(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexIntentExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "slot.#", "2"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexIntent_slotsCustom(t *testing.T) {
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexIntentDestroy,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAwsLexSlotTypeConfig_basic(testIntentID),
					testAccAwsLexIntentConfig_slotsCustom(testIntentID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexIntentExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "slot.#", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.description", "Types of flowers to pick up"),
					resource.TestCheckResourceAttr(rName, "slot.0.name", "FlowerType"),
					resource.TestCheckResourceAttr(rName, "slot.0.priority", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.sample_utterances.#", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.sample_utterances.0", "I would like to order {FlowerType}"),
					resource.TestCheckResourceAttr(rName, "slot.0.slot_constraint", "Required"),
					resource.TestCheckResourceAttr(rName, "slot.0.slot_type", testIntentID),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.#", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.0.max_attempts", "2"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.0.message.#", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.0.message.0.content", "What type of flowers would you like to order?"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.0.response_card", "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"What type of flowers?\",\"buttons\":[{\"text\":\"Tulips\",\"value\":\"tulips\"},{\"text\":\"Lilies\",\"value\":\"lilies\"},{\"text\":\"Roses\",\"value\":\"roses\"}]}]}"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexIntent_disappears(t *testing.T) {
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexIntentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexIntentConfig_basic(testIntentID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLexIntentExists(rName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsLexIntent(), rName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsLexIntent_updateWithExternalChange(t *testing.T) {
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	testAccCheckAwsLexIntentUpdateDescription := func(provider *schema.Provider, _ *schema.Resource, resourceName string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			conn := provider.Meta().(*AWSClient).lexmodelconn

			resourceState, ok := s.RootModule().Resources[resourceName]
			if !ok {
				return fmt.Errorf("intent not found: %s", resourceName)
			}

			input := &lexmodelbuildingservice.PutIntentInput{
				Checksum:    aws.String(resourceState.Primary.Attributes["checksum"]),
				Description: aws.String("Updated externally without Terraform"),
				Name:        aws.String(resourceState.Primary.ID),
				FulfillmentActivity: &lexmodelbuildingservice.FulfillmentActivity{
					Type: aws.String("ReturnIntent"),
				},
			}
			err := resource.Retry(1*time.Minute, func() *resource.RetryError {
				_, err := conn.PutIntent(input)

				if isAWSErr(err, lexmodelbuildingservice.ErrCodeConflictException, "") {
					return resource.RetryableError(fmt.Errorf("%q: intent still updating", resourceName))
				}
				if err != nil {
					return resource.NonRetryableError(err)
				}

				return nil
			})
			if err != nil {
				return fmt.Errorf("error updating intent %s: %w", resourceName, err)
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexIntentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexIntentConfig_basic(testIntentID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLexIntentExists(rName, &v),
					testAccCheckAwsLexIntentUpdateDescription(testAccProvider, resourceAwsLexIntent(), rName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAwsLexIntentConfig_basic(testIntentID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLexIntentExists(rName, &v),
				),
			},
		},
	})
}

func testAccCheckAwsLexIntentExistsWithVersion(rName, intentVersion string, output *lexmodelbuildingservice.GetIntentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Lex intent ID is set")
		}

		var err error
		conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

		output, err = conn.GetIntent(&lexmodelbuildingservice.GetIntentInput{
			Name:    aws.String(rs.Primary.ID),
			Version: aws.String(intentVersion),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return fmt.Errorf("error intent %q version %s not found", rs.Primary.ID, intentVersion)
		}
		if err != nil {
			return fmt.Errorf("error getting intent %q version %s: %w", rs.Primary.ID, intentVersion, err)
		}

		return nil
	}
}

func testAccCheckAwsLexIntentExists(rName string, output *lexmodelbuildingservice.GetIntentOutput) resource.TestCheckFunc {
	return testAccCheckAwsLexIntentExistsWithVersion(rName, LexIntentVersionLatest, output)
}

func testAccCheckAwsLexIntentNotExists(intentName, intentVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

		_, err := conn.GetIntent(&lexmodelbuildingservice.GetIntentInput{
			Name:    aws.String(intentName),
			Version: aws.String(intentVersion),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error getting intent %s version %s: %s", intentName, intentVersion, err)
		}

		return fmt.Errorf("error intent %s version %s exists", intentName, intentVersion)
	}
}

func testAccCheckAwsLexIntentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lex_intent" {
			continue
		}

		output, err := conn.GetIntentVersions(&lexmodelbuildingservice.GetIntentVersionsInput{
			Name: aws.String(rs.Primary.ID),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}

		if output == nil || len(output.Intents) == 0 {
			return nil
		}

		return fmt.Errorf("Lex intent %q still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAwsLexIntentConfig_lambda(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = "%[1]s"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

resource "aws_lambda_permission" "lex" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = "lex.amazonaws.com"
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s"
  handler       = "lambdatest.handler"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs12.x"
}
`, rName)
}

func testAccAwsLexIntentConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
}
`, rName)
}

func testAccAwsLexIntentConfig_createVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name           = "%s"
  create_version = true
  fulfillment_activity {
    type = "ReturnIntent"
  }
}
`, rName)
}

func testAccAwsLexIntentConfig_conclusionStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  conclusion_statement {
    message {
      content      = "Your order for {FlowerType} has been placed and will be ready by {PickupTime} on {PickupDate}"
      content_type = "PlainText"
    }
    response_card = "Your order for {FlowerType} has been placed and will be ready by {PickupTime} on {PickupDate}"
  }
}
`, rName)
}

func testAccAwsLexIntentConfig_conclusionStatementUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  conclusion_statement {
    message {
      content      = "Your order for {FlowerType} has been placed and will be ready by {PickupTime} on {PickupDate}"
      content_type = "PlainText"
      group_number = 1
    }
    message {
      content      = "Your order for {FlowerType} has been placed"
      content_type = "PlainText"
      group_number = 1
    }
    response_card = "Your order for {FlowerType} has been placed"
  }
}
`, rName)
}

func testAccAwsLexIntentConfig_confirmationPromptAndRejectionStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  confirmation_prompt {
    max_attempts = 1
    message {
      content      = "Okay, your {FlowerType} will be ready for pickup by {PickupTime} on {PickupDate}. Does this sound okay?"
      content_type = "PlainText"
    }
    response_card = "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"Okay, your {FlowerType} will be ready for pickup by {PickupTime} on {PickupDate}. Does this sound okay?\",\"buttons\":[{\"text\":\"Yes\",\"value\":\"yes\"},{\"text\":\"No\",\"value\":\"no\"}]}]}"
  }
  rejection_statement {
    message {
      content      = "Okay, I will not place your order."
      content_type = "PlainText"
    }
    response_card = "Okay, I will not place your order."
  }
}
`, rName)
}

func testAccAwsLexIntentConfig_confirmationPromptAndRejectionStatementUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  confirmation_prompt {
    max_attempts = 2
    message {
      content      = "Okay, your {FlowerType} will be ready for pickup by {PickupTime} on {PickupDate}. Does this sound okay?"
      content_type = "PlainText"
    }
    message {
      content      = "Okay, your {FlowerType} will be ready for pickup on {PickupDate}. Does this sound okay?"
      content_type = "PlainText"
    }
    response_card = "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"Okay, your {FlowerType} will be ready for pickup on {PickupDate}. Does this sound okay?\",\"buttons\":[{\"text\":\"Yes\",\"value\":\"yes\"},{\"text\":\"No\",\"value\":\"no\"}]}]}"
  }
  rejection_statement {
    message {
      content      = "Okay, I will not place your order."
      content_type = "PlainText"
    }
    message {
      content      = "Okay, your order has been cancelled."
      content_type = "PlainText"
    }
    response_card = "Okay, your order has been cancelled."
  }
}
`, rName)
}

func testAccAwsLexIntentConfig_dialogCodeHook(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  dialog_code_hook {
    message_version = "1"
    uri             = aws_lambda_function.test.arn
  }
}
`, rName)
}

func testAccAwsLexIntentConfig_followUpPrompt(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  follow_up_prompt {
    prompt {
      max_attempts = 1
      message {
        content      = "Would you like to order more flowers?"
        content_type = "PlainText"
      }
      response_card = "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"Would you like to order more flowers?\",\"buttons\":[{\"text\":\"Yes\",\"value\":\"yes\"},{\"text\":\"No\",\"value\":\"no\"}]}]}"
    }
    rejection_statement {
      message {
        content      = "Okay, no additional flowers will be ordered."
        content_type = "PlainText"
      }
      response_card = "Okay, no additional flowers will be ordered."
    }
  }
}
`, rName)
}

func testAccAwsLexIntentConfig_followUpPromptUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  follow_up_prompt {
    prompt {
      max_attempts = 2
      message {
        content      = "Would you like to order more flowers?"
        content_type = "PlainText"
      }
      message {
        content      = "Would you like to start another order?"
        content_type = "PlainText"
      }
      response_card = "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"Would you like to start another order?\",\"buttons\":[{\"text\":\"Yes\",\"value\":\"yes\"},{\"text\":\"No\",\"value\":\"no\"}]}]}"
    }
    rejection_statement {
      message {
        content      = "Okay, no additional flowers will be ordered."
        content_type = "PlainText"
      }
      message {
        content      = "Okay, additional flowers will be ordered."
        content_type = "PlainText"
      }
      response_card = "Okay, additional flowers will be ordered."
    }
  }
}
`, rName)
}

func testAccAwsLexIntentConfig_fulfillmentActivity(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    code_hook {
      message_version = "1"
      uri             = aws_lambda_function.test.arn
    }
    type = "CodeHook"
  }
}
`, rName)
}

func testAccAwsLexIntentConfig_sampleUtterances(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  sample_utterances = [
    "I would like to pick up flowers",
  ]
}
`, rName)
}

func testAccAwsLexIntentConfig_sampleUtterancesUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  sample_utterances = [
    "I would like to pick up flowers",
    "I would like to order some flowers",
  ]
}
`, rName)
}

func testAccAwsLexIntentConfig_slots(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  slot {
    description = "The date to pick up the flowers"
    name        = "PickupDate"
    priority    = 1
    sample_utterances = [
      "I would like to order {FlowerType}",
    ]
    slot_constraint = "Required"
    slot_type       = "AMAZON.DATE"
    value_elicitation_prompt {
      max_attempts = 1
      message {
        content      = "What day do you want the {FlowerType} to be picked up?"
        content_type = "PlainText"
      }
    }
  }
}
`, rName)
}

func testAccAwsLexIntentConfig_slotsUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  slot {
    description = "The date to pick up the flowers"
    name        = "PickupDate"
    priority    = 2
    sample_utterances = [
      "I would like to order {FlowerType}",
    ]
    slot_constraint = "Required"
    slot_type       = "AMAZON.DATE"
    value_elicitation_prompt {
      max_attempts = 2
      message {
        content      = "What day do you want the {FlowerType} to be picked up?"
        content_type = "PlainText"
      }
    }
  }
  slot {
    description = "The time to pick up the flowers"
    name        = "PickupTime"
    priority    = 1
    sample_utterances = [
      "I would like to order {FlowerType}",
    ]
    slot_constraint = "Required"
    slot_type       = "AMAZON.TIME"
    value_elicitation_prompt {
      max_attempts = 2
      message {
        content      = "Pick up the {FlowerType} at what time on {PickupDate}?"
        content_type = "PlainText"
      }
    }
  }
}
`, rName)
}

func testAccAwsLexIntentConfig_slotsCustom(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%[1]s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  slot {
    description = "Types of flowers to pick up"
    name        = "FlowerType"
    priority    = 1
    sample_utterances = [
      "I would like to order {FlowerType}",
    ]
    slot_constraint   = "Required"
    slot_type         = aws_lex_slot_type.test.name
    slot_type_version = "$LATEST"
    value_elicitation_prompt {
      max_attempts = 2
      message {
        content      = "What type of flowers would you like to order?"
        content_type = "PlainText"
      }
      response_card = "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"What type of flowers?\",\"buttons\":[{\"text\":\"Tulips\",\"value\":\"tulips\"},{\"text\":\"Lilies\",\"value\":\"lilies\"},{\"text\":\"Roses\",\"value\":\"roses\"}]}]}"
    }
  }
}
`, rName)
}
