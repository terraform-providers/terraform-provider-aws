package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_gamelift_alias", &resource.Sweeper{
		Name: "aws_gamelift_alias",
		Dependencies: []string{
			"aws_gamelift_fleet",
		},
		F: testSweepGameliftAliases,
	})
}

func testSweepGameliftAliases(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).gameliftconn

	err = listGameliftAliases(&gamelift.ListAliasesInput{}, conn, func(resp *gamelift.ListAliasesOutput) error {
		if len(resp.Aliases) == 0 {
			log.Print("[DEBUG] No Gamelift Aliases to sweep")
			return nil
		}

		log.Printf("[INFO] Found %d Gamelift Aliases", len(resp.Aliases))

		for _, alias := range resp.Aliases {
			log.Printf("[INFO] Deleting Gamelift Alias %q", *alias.AliasId)
			_, err := conn.DeleteAlias(&gamelift.DeleteAliasInput{
				AliasId: alias.AliasId,
			})
			if err != nil {
				return fmt.Errorf("Error deleting Gamelift Alias (%s): %s",
					*alias.AliasId, err)
			}
		}
		return nil
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Gamelift Alias sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing Gamelift Aliases: %s", err)
	}

	return nil
}

func listGameliftAliases(input *gamelift.ListAliasesInput, conn *gamelift.GameLift, f func(*gamelift.ListAliasesOutput) error) error {
	resp, err := conn.ListAliases(input)
	if err != nil {
		return err
	}
	err = f(resp)
	if err != nil {
		return err
	}

	if resp.NextToken != nil {
		return listGameliftAliases(input, conn, f)
	}
	return nil
}

func TestAccAWSGameliftAlias_basic(t *testing.T) {
	var conf gamelift.Alias

	rString := acctest.RandString(8)
	resourceName := "aws_gamelift_alias.test"

	aliasName := fmt.Sprintf("tf_acc_alias_%s", rString)
	description := fmt.Sprintf("tf test description %s", rString)
	message := fmt.Sprintf("tf test message %s", rString)

	uAliasName := fmt.Sprintf("tf_acc_alias_upd_%s", rString)
	uDescription := fmt.Sprintf("tf test updated description %s", rString)
	uMessage := fmt.Sprintf("tf test updated message %s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(gamelift.EndpointsID, t)
			testAccPreCheckAWSGamelift(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftAliasBasicConfig(aliasName, description, message),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftAliasExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "gamelift", regexp.MustCompile(`alias/alias-.+`)),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.0.message", message),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.0.type", "TERMINAL"),
					resource.TestCheckResourceAttr(resourceName, "name", aliasName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSGameliftAliasBasicConfig(uAliasName, uDescription, uMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftAliasExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "gamelift", regexp.MustCompile(`alias/.+`)),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.0.message", uMessage),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.0.type", "TERMINAL"),
					resource.TestCheckResourceAttr(resourceName, "name", uAliasName),
					resource.TestCheckResourceAttr(resourceName, "description", uDescription),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSGameliftAlias_tags(t *testing.T) {
	var conf gamelift.Alias

	resourceName := "aws_gamelift_alias.test"
	aliasName := acctest.RandomWithPrefix("tf-acc-alias")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(gamelift.EndpointsID, t)
			testAccPreCheckAWSGamelift(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftAliasBasicConfigTags1(aliasName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftAliasExists(resourceName, &conf),
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
				Config: testAccAWSGameliftAliasBasicConfigTags2(aliasName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftAliasExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSGameliftAliasBasicConfigTags1(aliasName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftAliasExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSGameliftAlias_fleetRouting(t *testing.T) {
	var conf gamelift.Alias

	rString := acctest.RandString(8)

	aliasName := fmt.Sprintf("tf_acc_alias_%s", rString)
	description := fmt.Sprintf("tf test description %s", rString)
	fleetName := fmt.Sprintf("tf_acc_fleet_%s", rString)
	buildName := fmt.Sprintf("tf_acc_build_%s", rString)

	region := testAccGetRegion()
	g, err := testAccAWSGameliftSampleGame(region)

	if isResourceNotFoundError(err) {
		t.Skip(err)
	}

	if err != nil {
		t.Fatal(err)
	}

	loc := g.Location
	bucketName := *loc.Bucket
	roleArn := *loc.RoleArn
	key := *loc.Key

	launchPath := g.LaunchPath
	params := g.Parameters(33435)
	resourceName := "aws_gamelift_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(gamelift.EndpointsID, t)
			testAccPreCheckAWSGamelift(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftAliasAllFieldsConfig(aliasName, description,
					fleetName, launchPath, params, buildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftAliasExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "gamelift", regexp.MustCompile(`alias/alias-.+`)),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "routing_strategy.0.fleet_id"),
					resource.TestCheckResourceAttr(resourceName, "routing_strategy.0.type", "SIMPLE"),
					resource.TestCheckResourceAttr(resourceName, "name", aliasName),
					resource.TestCheckResourceAttr(resourceName, "description", description),
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

func TestAccAWSGameliftAlias_disappears(t *testing.T) {
	var conf gamelift.Alias

	rString := acctest.RandString(8)
	resourceName := "aws_gamelift_alias.test"

	aliasName := fmt.Sprintf("tf_acc_alias_%s", rString)
	description := fmt.Sprintf("tf test description %s", rString)
	message := fmt.Sprintf("tf test message %s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPartitionHasServicePreCheck(gamelift.EndpointsID, t)
			testAccPreCheckAWSGamelift(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftAliasBasicConfig(aliasName, description, message),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftAliasExists(resourceName, &conf),
					testAccCheckAWSGameliftAliasDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSGameliftAliasDisappears(res *gamelift.Alias) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).gameliftconn

		input := &gamelift.DeleteAliasInput{AliasId: res.AliasId}

		_, err := conn.DeleteAlias(input)

		return err
	}
}

func testAccCheckAWSGameliftAliasExists(n string, res *gamelift.Alias) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Gamelift Alias ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).gameliftconn

		out, err := conn.DescribeAlias(&gamelift.DescribeAliasInput{
			AliasId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}
		a := out.Alias

		if *a.AliasId != rs.Primary.ID {
			return fmt.Errorf("Gamelift Alias not found")
		}

		*res = *a

		return nil
	}
}

func testAccCheckAWSGameliftAliasDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).gameliftconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_gamelift_alias" {
			continue
		}

		_, err := conn.DescribeAlias(&gamelift.DescribeAliasInput{
			AliasId: aws.String(rs.Primary.ID),
		})
		if err == nil {
			return fmt.Errorf("Gamelift Alias still exists")
		}

		if isAWSErr(err, gamelift.ErrCodeNotFoundException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccAWSGameliftAliasBasicConfig(aliasName, description, message string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_alias" "test" {
  name        = "%s"
  description = "%s"

  routing_strategy {
    message = "%s"
    type    = "TERMINAL"
  }
}
`, aliasName, description, message)
}

func testAccAWSGameliftAliasBasicConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_alias" "test" {
  name        = %[1]q
  description = "foo"

  routing_strategy {
    message = "bar"
    type    = "TERMINAL"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSGameliftAliasBasicConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_alias" "test" {
  name        = %[1]q
  description = "foo"

  routing_strategy {
    message = "bar"
    type    = "TERMINAL"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSGameliftAliasAllFieldsConfig(aliasName, description,
	fleetName, launchPath, params, buildName, bucketName, key, roleArn string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_alias" "test" {
  name        = "%s"
  description = "%s"

  routing_strategy {
    fleet_id = aws_gamelift_fleet.test.id
    type     = "SIMPLE"
  }
}
%s
`, aliasName, description,
		testAccAWSGameliftFleetBasicConfig(fleetName, launchPath, params, buildName, bucketName, key, roleArn))
}
