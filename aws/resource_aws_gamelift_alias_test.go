package aws

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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
			if !strings.HasPrefix(*alias.Name, "tf_acc_alias_") {
				continue
			}

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

	aliasName := fmt.Sprintf("tf_acc_alias_%s", rString)
	description := fmt.Sprintf("tf test description %s", rString)
	message := fmt.Sprintf("tf test message %s", rString)

	uAliasName := fmt.Sprintf("tf_acc_alias_upd_%s", rString)
	uDescription := fmt.Sprintf("tf test updated description %s", rString)
	uMessage := fmt.Sprintf("tf test updated message %s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftAliasBasicConfig(aliasName, description, message),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftAliasExists("aws_gamelift_alias.test", &conf),
					resource.TestCheckResourceAttrSet("aws_gamelift_alias.test", "arn"),
					resource.TestCheckResourceAttr("aws_gamelift_alias.test", "routing_strategy.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_alias.test", "routing_strategy.0.message", message),
					resource.TestCheckResourceAttr("aws_gamelift_alias.test", "routing_strategy.0.type", "TERMINAL"),
					resource.TestCheckResourceAttr("aws_gamelift_alias.test", "name", aliasName),
					resource.TestCheckResourceAttr("aws_gamelift_alias.test", "description", description),
				),
			},
			{
				Config: testAccAWSGameliftAliasBasicConfig(uAliasName, uDescription, uMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftAliasExists("aws_gamelift_alias.test", &conf),
					resource.TestCheckResourceAttrSet("aws_gamelift_alias.test", "arn"),
					resource.TestCheckResourceAttr("aws_gamelift_alias.test", "routing_strategy.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_alias.test", "routing_strategy.0.message", uMessage),
					resource.TestCheckResourceAttr("aws_gamelift_alias.test", "routing_strategy.0.type", "TERMINAL"),
					resource.TestCheckResourceAttr("aws_gamelift_alias.test", "name", uAliasName),
					resource.TestCheckResourceAttr("aws_gamelift_alias.test", "description", uDescription),
				),
			},
		},
	})
}

func TestAccAWSGameliftAlias_importBasic(t *testing.T) {
	rString := acctest.RandString(8)

	aliasName := fmt.Sprintf("tf_acc_alias_%s", rString)
	description := fmt.Sprintf("tf test description %s", rString)
	message := fmt.Sprintf("tf test message %s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftAliasBasicConfig(aliasName, description, message),
			},
			{
				ResourceName:      "aws_gamelift_alias.test",
				ImportState:       true,
				ImportStateVerify: true,
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
	if err != nil {
		t.Fatal(err)
	}

	loc := g.Location
	bucketName := *loc.Bucket
	roleArn := *loc.RoleArn
	key := *loc.Key

	launchPath := g.LaunchPath
	params := g.Parameters(33435)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftAliasDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftAliasAllFieldsConfig(aliasName, description,
					fleetName, launchPath, params, buildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftAliasExists("aws_gamelift_alias.test", &conf),
					resource.TestCheckResourceAttrSet("aws_gamelift_alias.test", "arn"),
					resource.TestCheckResourceAttr("aws_gamelift_alias.test", "routing_strategy.#", "1"),
					resource.TestCheckResourceAttrSet("aws_gamelift_alias.test", "routing_strategy.0.fleet_id"),
					resource.TestCheckResourceAttr("aws_gamelift_alias.test", "routing_strategy.0.type", "SIMPLE"),
					resource.TestCheckResourceAttr("aws_gamelift_alias.test", "name", aliasName),
					resource.TestCheckResourceAttr("aws_gamelift_alias.test", "description", description),
				),
			},
		},
	})
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
  name = "%s"
  description = "%s"
  routing_strategy {
    message = "%s"
    type = "TERMINAL"
  }
}
`, aliasName, description, message)
}

func testAccAWSGameliftAliasAllFieldsConfig(aliasName, description,
	fleetName, launchPath, params, buildName, bucketName, key, roleArn string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_alias" "test" {
  name = "%s"
  description = "%s"
  routing_strategy {
    fleet_id = "${aws_gamelift_fleet.test.id}"
    type = "SIMPLE"
  }
}
%s
`, aliasName, description,
		testAccAWSGameliftFleetBasicConfig(fleetName, launchPath, params, buildName, bucketName, key, roleArn))
}
