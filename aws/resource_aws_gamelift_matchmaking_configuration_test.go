package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const testAccGameliftMatchmakingConfigurationPrefix = "tfAccMMConfiguration-"
const testAccGameliftMatchmakingQueuePrefix = "tfAccMatchMakingQueue-"

func init() {
	resource.AddTestSweepers("aws_gamelift_matchmaking_configuration", &resource.Sweeper{
		Name: "aws_gamelift_matchmaking_configuration",
		F:    testSweepGameliftMatchmakingConfiguration,
	})
}

func testSweepGameliftMatchmakingConfiguration(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).gameliftconn

	out, err := conn.DescribeMatchmakingConfigurations(&gamelift.DescribeMatchmakingConfigurationsInput{})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Gamelift Matchmaking Configuration sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Gamelift Configuration: %s", err)
	}

	if len(out.Configurations) == 0 {
		log.Print("[DEBUG] No Gamelift Matchmaking Configuration to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d Gamelift Matchmaking Configurations", len(out.Configurations))

	for _, config := range out.Configurations {
		log.Printf("[INFO] Deleting Gamelift Matchmaking Configuration %q", *config.Name)
		_, err := conn.DeleteMatchmakingConfiguration(&gamelift.DeleteMatchmakingConfigurationInput{
			Name: aws.String(*config.Name),
		})
		if err != nil {
			return fmt.Errorf("error deleting Gamelift Matchmaking Configuration (%s): %s",
				*config.Name, err)
		}
	}

	return nil
}

func TestAccAWSGameliftMatchmakingConfiguration_basic(t *testing.T) {
	var conf gamelift.MatchmakingConfiguration

	resourceName := "aws_gamelift_matchmaking_configuration.test"
	configurationName := testAccGameliftMatchmakingConfigurationPrefix + acctest.RandString(8)
	uConfigurationName := configurationName + "-updated"

	queueName := testAccGameliftMatchmakingQueuePrefix + acctest.RandString(8)
	ruleSetName := testAccGameliftMatchmakingRuleSetPrefix + acctest.RandString(8)

	backfillMode := gamelift.BackfillModeManual

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGamelift(t) },
		Providers:    testAccProviders,
		ErrorCheck:   testAccErrorCheck(t, gamelift.EndpointsID),
		CheckDestroy: testAccCheckAWSGameliftMatchmakingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftMatchmakingConfigurationBasicConfig(configurationName, queueName, ruleSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftMatchmakingConfigurationExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`matchmakingconfiguration/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", configurationName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "custom_event_data", "pvp"),
					resource.TestCheckResourceAttr(resourceName, "game_session_data", "game_session_data"),
					resource.TestCheckResourceAttr(resourceName, "acceptance_required", "false"),
					resource.TestCheckResourceAttr(resourceName, "request_timeout_seconds", "10"),
					resource.TestCheckResourceAttr(resourceName, "backfill_mode", backfillMode),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSGameliftMatchmakingConfigurationRequestTimeout(uConfigurationName, queueName, ruleSetName, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftMatchmakingConfigurationExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", uConfigurationName),
					resource.TestCheckResourceAttr(resourceName, "request_timeout_seconds", "20"),
				),
			},
		},
	})
}

func TestAccAWSGameliftMatchmakingConfiguration_tags(t *testing.T) {
	var conf gamelift.MatchmakingConfiguration

	resourceName := "aws_gamelift_matchmaking_configuration.test"
	configurationName := testAccGameliftMatchmakingConfigurationPrefix + acctest.RandString(8)

	queueName := testAccGameliftMatchmakingQueuePrefix + acctest.RandString(8)
	ruleSetName := testAccGameliftMatchmakingRuleSetPrefix + acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGamelift(t) },
		Providers:    testAccProviders,
		ErrorCheck:   testAccErrorCheck(t, gamelift.EndpointsID),
		CheckDestroy: testAccCheckAWSGameliftMatchmakingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftMatchmakingConfigurationTags1(configurationName, queueName, ruleSetName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftMatchmakingConfigurationExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`matchmakingconfiguration/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", configurationName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAWSGameliftMatchmakingConfigurationTags2(configurationName, queueName, ruleSetName, "key1", "value1", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftMatchmakingConfigurationExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`matchmakingconfiguration/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", configurationName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

func TestAccAWSGameliftMatchmakingConfiguration_disappears(t *testing.T) {
	var conf gamelift.MatchmakingConfiguration

	resourceName := "aws_gamelift_matchmaking_configuration.test"
	configurationName := testAccGameliftMatchmakingConfigurationPrefix + acctest.RandString(8)
	
	queueName := testAccGameliftMatchmakingQueuePrefix + acctest.RandString(8)
	ruleSetName := testAccGameliftMatchmakingRuleSetPrefix + acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGamelift(t) },
		Providers:    testAccProviders,
		ErrorCheck:   testAccErrorCheck(t, gamelift.EndpointsID),
		CheckDestroy: testAccCheckAWSGameliftMatchmakingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftMatchmakingConfigurationBasicConfig(configurationName, queueName, ruleSetName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftMatchmakingConfigurationExists(resourceName, &conf),
					testAccCheckAWSGameliftMatchmakingConfigurationDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSGameliftMatchmakingConfigurationExists(n string, res *gamelift.MatchmakingConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Gamelift Matchmaking Configuration Name is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).gameliftconn

		name := rs.Primary.Attributes["name"]
		out, err := conn.DescribeMatchmakingConfigurations(&gamelift.DescribeMatchmakingConfigurationsInput{
			Names: aws.StringSlice([]string{name}),
		})
		if err != nil {
			return err
		}
		configurations := out.Configurations
		if len(configurations) == 0 {
			return fmt.Errorf("GameLift Matchmaking Configuration %q not found", name)
		}

		*res = *configurations[0]

		return nil
	}
}

func testAccCheckAWSGameliftMatchmakingConfigurationDisappears(res *gamelift.MatchmakingConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).gameliftconn

		input := &gamelift.DeleteMatchmakingConfigurationInput{Name: res.Name}
		err := resource.Retry(60*time.Minute, func() *resource.RetryError {
			_, err := conn.DeleteMatchmakingConfiguration(input)
			if err != nil {
				if isAWSErr(err, gamelift.ErrCodeInvalidRequestException, "Configuration not found") {
					return nil
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if isResourceTimeoutError(err) {
			_, err = conn.DeleteMatchmakingConfiguration(input)
		}
		if err != nil {
			return fmt.Errorf("Error deleting match making configuration: %s", err)
		}

		return nil
	}
}

func testAccCheckAWSGameliftMatchmakingConfigurationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).gameliftconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_gamelift_matchmaking_configuration" {
			continue
		}

		name := rs.Primary.Attributes["name"]

		input := &gamelift.DescribeMatchmakingConfigurationsInput{
			Names: aws.StringSlice([]string{name}),
		}

		// Deletions can take a few seconds
		err := resource.Retry(30*time.Second, func() *resource.RetryError {
			out, err := conn.DescribeMatchmakingConfigurations(input)

			if isAWSErr(err, gamelift.ErrCodeInvalidRequestException, "Configuration not found") {
				return nil
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			configurations := out.Configurations

			if len(configurations) > 0 {
				return resource.RetryableError(fmt.Errorf("GameLift Matchmaking Configuration still exists"))
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccAWSGameliftMatchMakingConfigurationRuleSetBody() string {
	maxPlayers := int64(1)
	return fmt.Sprintf(`{
		"name": "test",
		"ruleLanguageVersion": "1.0",
		"teams": [{
			"name": "alpha",
			"minPlayers": 1,
			"maxPlayers": %[1]d
		}]
	}`, maxPlayers)
}

func testAccAWSGameliftMatchmakingConfigurationBasicConfig(rName string, queueName string, ruleSetName string) string {
	return testAccAWSGameliftMatchmakingConfigurationSharedConfig(rName, queueName, ruleSetName, "", 10)
}

func testAccAWSGameliftMatchmakingConfigurationRequestTimeout(rName string, queueName string, ruleSetName string, requestTimeoutSeconds int) string {
	return testAccAWSGameliftMatchmakingConfigurationSharedConfig(rName, queueName, ruleSetName, "", requestTimeoutSeconds)
}

func testAccAWSGameliftMatchmakingConfigurationTags1(rName string, queueName string, ruleSetName string, tagName string, tagValue string) string {
	parameters := fmt.Sprintf("tags = {\n%q = %q\n}", tagName, tagValue)
	
	return testAccAWSGameliftMatchmakingConfigurationSharedConfig(rName, queueName, ruleSetName, parameters, 10)
}

func testAccAWSGameliftMatchmakingConfigurationTags2(rName string, queueName string, ruleSetName string, tagName string, tagValue string, tagName2 string, tagValue2 string) string {
	parameters := fmt.Sprintf("tags = {\n%q = %q\n%q = %q\n}", tagName, tagValue, tagName2, tagValue2)
	
	return testAccAWSGameliftMatchmakingConfigurationSharedConfig(rName, queueName, ruleSetName, parameters, 10)
}

func testAccAWSGameliftMatchmakingConfigurationSharedConfig(rName string, queueName string, ruleSetName string, additionalParameters string, requestTimeoutSeconds int) string {
	backfillMode := gamelift.BackfillModeManual
	return fmt.Sprintf(`
resource "aws_gamelift_game_session_queue" "test" {
	name         = %[2]q
	destinations = []
	
	player_latency_policy {
		maximum_individual_player_latency_milliseconds = 3
		policy_duration_seconds                        = 7
	}
	
	player_latency_policy {
		maximum_individual_player_latency_milliseconds = 10
	}
	
	timeout_in_seconds = 25
}

resource "aws_gamelift_matchmaking_rule_set" "test" {
	name          = %[3]q
	rule_set_body = <<RULE_SET_BODY
	%[4]s
	RULE_SET_BODY	
}

resource "aws_gamelift_matchmaking_configuration" "test" {
	name          = %[1]q
	acceptance_required = false
	custom_event_data = "pvp"
	game_session_data = "game_session_data"
	backfill_mode = %[7]q
	request_timeout_seconds = %[6]d
	rule_set_name = aws_gamelift_matchmaking_rule_set.test.name
	game_session_queue_arns = [aws_gamelift_game_session_queue.test.arn]
	%[5]s
}
`, rName, queueName, ruleSetName, testAccAWSGameliftMatchMakingConfigurationRuleSetBody(), additionalParameters, requestTimeoutSeconds, backfillMode)
}
