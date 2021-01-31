package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccDataSourceAwsLexBotAlias_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	dataSourceName := "data.aws_lex_bot_alias.test"
	resourceName := "aws_lex_bot_alias.test"

	// If this test runs in parallel with other Lex Bot tests, it loses its description
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lexmodelbuildingservice.EndpointsID, t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAwsLexBotConfig_intent(rName),
					testAccAwsLexBotConfig_createVersion(rName),
					testAccAwsLexBotAliasConfig_basic(rName),
					testAccDataSourceAwsLexBotAliasConfig_basic(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bot_name", resourceName, "bot_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bot_version", resourceName, "bot_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "checksum", resourceName, "checksum"),
					resource.TestCheckResourceAttrPair(dataSourceName, "created_date", resourceName, "created_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_updated_date", resourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsLexBotAliasConfig_basic() string {
	return `
data "aws_lex_bot_alias" "test" {
  name     = aws_lex_bot_alias.test.name
  bot_name = aws_lex_bot.test.name
}
`
}
