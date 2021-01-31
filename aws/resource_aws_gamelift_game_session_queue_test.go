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

const testAccGameliftGameSessionQueuePrefix = "tfAccQueue-"

func init() {
	resource.AddTestSweepers("aws_gamelift_game_session_queue", &resource.Sweeper{
		Name: "aws_gamelift_game_session_queue",
		F:    testSweepGameliftGameSessionQueue,
	})
}

func testSweepGameliftGameSessionQueue(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).gameliftconn

	out, err := conn.DescribeGameSessionQueues(&gamelift.DescribeGameSessionQueuesInput{})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Gamelife Queue sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Gamelift Session Queue: %s", err)
	}

	if len(out.GameSessionQueues) == 0 {
		log.Print("[DEBUG] No Gamelift Session Queue to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d Gamelift Session Queue", len(out.GameSessionQueues))

	for _, queue := range out.GameSessionQueues {
		log.Printf("[INFO] Deleting Gamelift Session Queue %q", *queue.Name)
		_, err := conn.DeleteGameSessionQueue(&gamelift.DeleteGameSessionQueueInput{
			Name: aws.String(*queue.Name),
		})
		if err != nil {
			return fmt.Errorf("error deleting Gamelift Session Queue (%s): %s",
				*queue.Name, err)
		}
	}

	return nil
}

func TestAccAWSGameliftGameSessionQueue_basic(t *testing.T) {
	var conf gamelift.GameSessionQueue

	resourceName := "aws_gamelift_game_session_queue.test"
	queueName := testAccGameliftGameSessionQueuePrefix + acctest.RandString(8)
	playerLatencyPolicies := []gamelift.PlayerLatencyPolicy{
		{
			MaximumIndividualPlayerLatencyMilliseconds: aws.Int64(100),
			PolicyDurationSeconds:                      aws.Int64(5),
		},
		{
			MaximumIndividualPlayerLatencyMilliseconds: aws.Int64(200),
			PolicyDurationSeconds:                      nil,
		},
	}
	timeoutInSeconds := int64(124)

	uQueueName := queueName + "-updated"
	uPlayerLatencyPolicies := []gamelift.PlayerLatencyPolicy{
		{
			MaximumIndividualPlayerLatencyMilliseconds: aws.Int64(150),
			PolicyDurationSeconds:                      aws.Int64(10),
		},
		{
			MaximumIndividualPlayerLatencyMilliseconds: aws.Int64(250),
			PolicyDurationSeconds:                      nil,
		},
	}
	uTimeoutInSeconds := int64(600)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(gamelift.EndpointsID, t)
			testAccPreCheckAWSGamelift(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftGameSessionQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftGameSessionQueueBasicConfig(queueName,
					playerLatencyPolicies, timeoutInSeconds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftGameSessionQueueExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`gamesessionqueue/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", queueName),
					resource.TestCheckResourceAttr(resourceName, "destinations.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.0.maximum_individual_player_latency_milliseconds",
						fmt.Sprintf("%d", *playerLatencyPolicies[0].MaximumIndividualPlayerLatencyMilliseconds)),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.0.policy_duration_seconds",
						fmt.Sprintf("%d", *playerLatencyPolicies[0].PolicyDurationSeconds)),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.1.maximum_individual_player_latency_milliseconds",
						fmt.Sprintf("%d", *playerLatencyPolicies[1].MaximumIndividualPlayerLatencyMilliseconds)),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.1.policy_duration_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout_in_seconds", fmt.Sprintf("%d", timeoutInSeconds)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSGameliftGameSessionQueueBasicConfig(uQueueName, uPlayerLatencyPolicies, uTimeoutInSeconds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftGameSessionQueueExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`gamesessionqueue/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", uQueueName),
					resource.TestCheckResourceAttr(resourceName, "destinations.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.0.maximum_individual_player_latency_milliseconds",
						fmt.Sprintf("%d", *uPlayerLatencyPolicies[0].MaximumIndividualPlayerLatencyMilliseconds)),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.0.policy_duration_seconds",
						fmt.Sprintf("%d", *uPlayerLatencyPolicies[0].PolicyDurationSeconds)),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.1.maximum_individual_player_latency_milliseconds",
						fmt.Sprintf("%d", *uPlayerLatencyPolicies[1].MaximumIndividualPlayerLatencyMilliseconds)),
					resource.TestCheckResourceAttr(resourceName, "player_latency_policy.1.policy_duration_seconds", "0"),
					resource.TestCheckResourceAttr(resourceName, "timeout_in_seconds", fmt.Sprintf("%d", uTimeoutInSeconds)),
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

func TestAccAWSGameliftGameSessionQueue_tags(t *testing.T) {
	var conf gamelift.GameSessionQueue

	resourceName := "aws_gamelift_game_session_queue.test"
	queueName := testAccGameliftGameSessionQueuePrefix + acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(gamelift.EndpointsID, t)
			testAccPreCheckAWSGamelift(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftGameSessionQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftGameSessionQueueBasicConfigTags1(queueName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftGameSessionQueueExists(resourceName, &conf),
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
				Config: testAccAWSGameliftGameSessionQueueBasicConfigTags2(queueName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftGameSessionQueueExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSGameliftGameSessionQueueBasicConfigTags1(queueName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftGameSessionQueueExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSGameliftGameSessionQueue_disappears(t *testing.T) {
	var conf gamelift.GameSessionQueue

	resourceName := "aws_gamelift_game_session_queue.test"
	queueName := testAccGameliftGameSessionQueuePrefix + acctest.RandString(8)
	playerLatencyPolicies := []gamelift.PlayerLatencyPolicy{
		{
			MaximumIndividualPlayerLatencyMilliseconds: aws.Int64(100),
			PolicyDurationSeconds:                      aws.Int64(5),
		},
		{
			MaximumIndividualPlayerLatencyMilliseconds: aws.Int64(200),
			PolicyDurationSeconds:                      nil,
		},
	}
	timeoutInSeconds := int64(124)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(gamelift.EndpointsID, t)
			testAccPreCheckAWSGamelift(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftGameSessionQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftGameSessionQueueBasicConfig(queueName,
					playerLatencyPolicies, timeoutInSeconds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftGameSessionQueueExists(resourceName, &conf),
					testAccCheckAWSGameliftGameSessionQueueDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSGameliftGameSessionQueueExists(n string, res *gamelift.GameSessionQueue) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Gamelift Session Queue Name is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).gameliftconn

		name := rs.Primary.Attributes["name"]
		limit := int64(1)
		out, err := conn.DescribeGameSessionQueues(&gamelift.DescribeGameSessionQueuesInput{
			Names: aws.StringSlice([]string{name}),
			Limit: &limit,
		})
		if err != nil {
			return err
		}
		attributes := out.GameSessionQueues
		if len(attributes) < 1 {
			return fmt.Errorf("gmelift Session Queue %q not found", name)
		}
		if len(attributes) != 1 {
			return fmt.Errorf("expected exactly 1 Gamelift Session Queue, found %d under %q",
				len(attributes), name)
		}
		queue := attributes[0]

		if *queue.Name != name {
			return fmt.Errorf("gamelift Session Queue not found")
		}

		*res = *queue

		return nil
	}
}

func testAccCheckAWSGameliftGameSessionQueueDisappears(res *gamelift.GameSessionQueue) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).gameliftconn

		input := &gamelift.DeleteGameSessionQueueInput{Name: res.Name}

		_, err := conn.DeleteGameSessionQueue(input)

		return err
	}
}

func testAccCheckAWSGameliftGameSessionQueueDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).gameliftconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_gamelift_game_session_queue" {
			continue
		}

		input := &gamelift.DescribeGameSessionQueuesInput{
			Names: aws.StringSlice([]string{rs.Primary.ID}),
			Limit: aws.Int64(1),
		}

		// Deletions can take a few seconds
		err := resource.Retry(30*time.Second, func() *resource.RetryError {
			out, err := conn.DescribeGameSessionQueues(input)

			if isAWSErr(err, gamelift.ErrCodeNotFoundException, "") {
				return nil
			}

			if err != nil {
				return resource.NonRetryableError(err)
			}

			attributes := out.GameSessionQueues

			if len(attributes) > 0 {
				return resource.RetryableError(fmt.Errorf("gamelift Session Queue still exists"))
			}

			return nil
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccAWSGameliftGameSessionQueueBasicConfig(queueName string,
	playerLatencyPolicies []gamelift.PlayerLatencyPolicy, timeoutInSeconds int64) string {
	return fmt.Sprintf(`
resource "aws_gamelift_game_session_queue" "test" {
  name         = "%s"
  destinations = []

  player_latency_policy {
    maximum_individual_player_latency_milliseconds = %d
    policy_duration_seconds                        = %d
  }

  player_latency_policy {
    maximum_individual_player_latency_milliseconds = %d
  }

  timeout_in_seconds = %d
}
`,
		queueName,
		*playerLatencyPolicies[0].MaximumIndividualPlayerLatencyMilliseconds,
		*playerLatencyPolicies[0].PolicyDurationSeconds,
		*playerLatencyPolicies[1].MaximumIndividualPlayerLatencyMilliseconds,
		timeoutInSeconds)
}

func testAccAWSGameliftGameSessionQueueBasicConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_game_session_queue" "test" {
  name         = %[1]q
  destinations = []

  player_latency_policy {
    maximum_individual_player_latency_milliseconds = 100000
    policy_duration_seconds                        = 10
  }

  player_latency_policy {
    maximum_individual_player_latency_milliseconds = 100000
  }

  timeout_in_seconds = 10

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSGameliftGameSessionQueueBasicConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_game_session_queue" "test" {
  name         = %[1]q
  destinations = []

  player_latency_policy {
    maximum_individual_player_latency_milliseconds = 100000
    policy_duration_seconds                        = 10
  }

  player_latency_policy {
    maximum_individual_player_latency_milliseconds = 100000
  }

  timeout_in_seconds = 10

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
