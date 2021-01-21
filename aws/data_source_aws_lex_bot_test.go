package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsLexBot_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	dataSourceName := "data.aws_lex_bot.test"
	resourceName := "aws_lex_bot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAwsLexBotConfig_intent(rName),
					testAccAwsLexBotConfig_basic(rName),
					testAccDataSourceAwsLexBotConfig_basic(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "checksum", resourceName, "checksum"),
					resource.TestCheckResourceAttrPair(dataSourceName, "child_directed", resourceName, "child_directed"),
					resource.TestCheckResourceAttrPair(dataSourceName, "created_date", resourceName, "created_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "detect_sentiment", resourceName, "detect_sentiment"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_model_improvements", resourceName, "enable_model_improvements"),
					resource.TestCheckResourceAttrPair(dataSourceName, "failure_reason", resourceName, "failure_reason"),
					resource.TestCheckResourceAttrPair(dataSourceName, "idle_session_ttl_in_seconds", resourceName, "idle_session_ttl_in_seconds"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_updated_date", resourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "locale", resourceName, "locale"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "nlu_intent_confidence_threshold", resourceName, "nlu_intent_confidence_threshold"),
					resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", resourceName, "version"),
				),
			},
		},
	})
}

func testAccDataSourceAwsLexBot_withVersion(t *testing.T) {
	rName := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	dataSourceName := "data.aws_lex_bot.test"
	resourceName := "aws_lex_bot.test"

	// If this test runs in parallel with other Lex Bot tests, it loses its description
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAwsLexBotConfig_intent(rName),
					testAccAwsLexBotConfig_createVersion(rName),
					testAccDataSourceAwsLexBotConfig_withVersion(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "checksum", resourceName, "checksum"),
					resource.TestCheckResourceAttrPair(dataSourceName, "child_directed", resourceName, "child_directed"),
					resource.TestCheckResourceAttrPair(dataSourceName, "created_date", resourceName, "created_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "detect_sentiment", resourceName, "detect_sentiment"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_model_improvements", resourceName, "enable_model_improvements"),
					resource.TestCheckResourceAttrPair(dataSourceName, "failure_reason", resourceName, "failure_reason"),
					resource.TestCheckResourceAttrPair(dataSourceName, "idle_session_ttl_in_seconds", resourceName, "idle_session_ttl_in_seconds"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_updated_date", resourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "locale", resourceName, "locale"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "nlu_intent_confidence_threshold", resourceName, "nlu_intent_confidence_threshold"),
					resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", resourceName, "version"),
				),
			},
		},
	})
}

func testAccDataSourceAwsLexBotConfig_basic() string {
	return `
data "aws_lex_bot" "test" {
  name = aws_lex_bot.test.name
}
`
}

func testAccDataSourceAwsLexBotConfig_withVersion() string {
	return `
data "aws_lex_bot" "test" {
  name    = aws_lex_bot.test.name
  version = "1"
}
`
}
