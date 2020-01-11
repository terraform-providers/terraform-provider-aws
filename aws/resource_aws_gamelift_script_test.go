package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/gamelift"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_gamelift_script", &resource.Sweeper{
		Name: "aws_gamelift_script",
		F:    testSweepGameliftScripts,
	})
}

func testSweepGameliftScripts(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).gameliftconn

	resp, err := conn.ListScripts(&gamelift.ListScriptsInput{})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Gamelife Script sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing Gamelift Scripts: %s", err)
	}

	if len(resp.Scripts) == 0 {
		log.Print("[DEBUG] No Gamelift Scripts to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d Gamelift Scripts", len(resp.Scripts))

	for _, script := range resp.Scripts {
		log.Printf("[INFO] Deleting Gamelift Script %q", *script.ScriptId)
		_, err := conn.DeleteScript(&gamelift.DeleteScriptInput{
			ScriptId: script.ScriptId,
		})
		if err != nil {
			return fmt.Errorf("Error deleting Gamelift Script (%s): %s",
				*script.ScriptId, err)
		}
	}

	return nil
}

func TestAccAWSGameliftScript_basic(t *testing.T) {
	var conf gamelift.Script

	resourceName := "aws_gamelift_script.test"

	rName := acctest.RandomWithPrefix("acc-test-script")
	uScriptName := acctest.RandomWithPrefix("acc-test-script-upd")

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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGameliftScripts(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftScriptBasicConfig(rName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftScriptExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`script/script-.+`)),
					resource.TestCheckResourceAttr(resourceName, "storage_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.key", key),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.role_arn", roleArn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSGameliftScriptBasicConfig(uScriptName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftScriptExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", uScriptName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`script/script-.+`)),
					resource.TestCheckResourceAttr(resourceName, "storage_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.key", key),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.role_arn", roleArn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSGameliftScript_tags(t *testing.T) {
	var conf gamelift.Script

	resourceName := "aws_gamelift_script.test"

	rName := acctest.RandomWithPrefix("acc-test-script")
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGameliftScripts(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftScriptBasicConfigTags1(rName, bucketName, key, roleArn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftScriptExists(resourceName, &conf),
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
				Config: testAccAWSGameliftScriptBasicConfigTags2(rName, bucketName, key, roleArn, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftScriptExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSGameliftScriptBasicConfigTags1(rName, bucketName, key, roleArn, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftScriptExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSGameliftScript_disappears(t *testing.T) {
	var conf gamelift.Script

	resourceName := "aws_gamelift_script.test"

	rName := acctest.RandomWithPrefix("acc-test-script")

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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGameliftScripts(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftScriptDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftScriptBasicConfig(rName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftScriptExists(resourceName, &conf),
					testAccCheckAWSGameliftScriptDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSGameliftScriptExists(n string, res *gamelift.Script) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Gamelift Script ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).gameliftconn

		req := &gamelift.DescribeScriptInput{
			ScriptId: aws.String(rs.Primary.ID),
		}
		out, err := conn.DescribeScript(req)
		if err != nil {
			return err
		}

		b := out.Script

		if *b.ScriptId != rs.Primary.ID {
			return fmt.Errorf("Gamelift Script not found")
		}

		*res = *b

		return nil
	}
}

func testAccCheckAWSGameliftScriptDisappears(res *gamelift.Script) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).gameliftconn

		input := &gamelift.DeleteScriptInput{ScriptId: res.ScriptId}

		_, err := conn.DeleteScript(input)
		return err
	}
}

func testAccCheckAWSGameliftScriptDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).gameliftconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_gamelift_script" {
			continue
		}

		req := gamelift.DescribeScriptInput{
			ScriptId: aws.String(rs.Primary.ID),
		}
		out, err := conn.DescribeScript(&req)
		if err == nil {
			if *out.Script.ScriptId == rs.Primary.ID {
				return fmt.Errorf("Gamelift Script still exists")
			}
		}
		if isAWSErr(err, gamelift.ErrCodeNotFoundException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccPreCheckAWSGameliftScripts(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).gameliftconn

	input := &gamelift.ListScriptsInput{}

	_, err := conn.ListScripts(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSGameliftScriptBasicConfig(rName, bucketName, key, roleArn string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_script" "test" {
  name             = "%s"

  storage_location {
    bucket   = "%s"
    key      = "%s"
    role_arn = "%s"
  }
}
`, rName, bucketName, key, roleArn)
}

func testAccAWSGameliftScriptBasicConfigTags1(rName, bucketName, key, roleArn, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_script" "test" {
  name             = %[1]q

  storage_location {
    bucket   = %[2]q
    key      = %[3]q
    role_arn = %[4]q
  }

  tags = {
    %[5]q = %[6]q
  }
}
`, rName, bucketName, key, roleArn, tagKey1, tagValue1)
}

func testAccAWSGameliftScriptBasicConfigTags2(rName, bucketName, key, roleArn, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_script" "test" {
  name             = %[1]q

  storage_location {
    bucket   = %[2]q
    key      = %[3]q
    role_arn = %[4]q
  }

  tags = {
    %[5]q = %[6]q
    %[7]q = %[8]q
  }
}
`, rName, bucketName, key, roleArn, tagKey1, tagValue1, tagKey2, tagValue2)
}
