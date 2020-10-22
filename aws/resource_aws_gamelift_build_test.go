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

const testAccGameliftBuildPrefix = "tf_acc_build_"

func init() {
	resource.AddTestSweepers("aws_gamelift_build", &resource.Sweeper{
		Name: "aws_gamelift_build",
		F:    testSweepGameliftBuilds,
	})
}

func testSweepGameliftBuilds(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).gameliftconn

	resp, err := conn.ListBuilds(&gamelift.ListBuildsInput{})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Gamelife Build sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing Gamelift Builds: %s", err)
	}

	if len(resp.Builds) == 0 {
		log.Print("[DEBUG] No Gamelift Builds to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d Gamelift Builds", len(resp.Builds))

	for _, build := range resp.Builds {
		log.Printf("[INFO] Deleting Gamelift Build %q", *build.BuildId)
		_, err := conn.DeleteBuild(&gamelift.DeleteBuildInput{
			BuildId: build.BuildId,
		})
		if err != nil {
			return fmt.Errorf("Error deleting Gamelift Build (%s): %s",
				*build.BuildId, err)
		}
	}

	return nil
}

func TestAccAWSGameliftBuild_basic(t *testing.T) {
	var conf gamelift.Build

	rString := acctest.RandString(8)
	resourceName := "aws_gamelift_build.test"

	buildName := fmt.Sprintf("%s_%s", testAccGameliftBuildPrefix, rString)
	uBuildName := fmt.Sprintf("%s_updated_%s", testAccGameliftBuildPrefix, rString)

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
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGamelift(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftBuildDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftBuildBasicConfig(buildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftBuildExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", buildName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`build/build-.+`)),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "WINDOWS_2012"),
					resource.TestCheckResourceAttr(resourceName, "storage_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.bucket", bucketName),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.key", key),
					resource.TestCheckResourceAttr(resourceName, "storage_location.0.role_arn", roleArn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSGameliftBuildBasicConfig(uBuildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftBuildExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "name", uBuildName),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "gamelift", regexp.MustCompile(`build/build-.+`)),
					resource.TestCheckResourceAttr(resourceName, "operating_system", "WINDOWS_2012"),
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

func TestAccAWSGameliftBuild_tags(t *testing.T) {
	var conf gamelift.Build

	rString := acctest.RandString(8)
	resourceName := "aws_gamelift_build.test"

	buildName := fmt.Sprintf("%s_%s", testAccGameliftBuildPrefix, rString)
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
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGamelift(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftBuildDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftBuildBasicConfigTags1(buildName, bucketName, key, roleArn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftBuildExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAWSGameliftBuildBasicConfigTags2(buildName, bucketName, key, roleArn, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftBuildExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSGameliftBuildBasicConfigTags1(buildName, bucketName, key, roleArn, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftBuildExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSGameliftBuild_disappears(t *testing.T) {
	var conf gamelift.Build

	rString := acctest.RandString(8)
	resourceName := "aws_gamelift_build.test"

	buildName := fmt.Sprintf("%s_%s", testAccGameliftBuildPrefix, rString)

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
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSGamelift(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftBuildDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftBuildBasicConfig(buildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftBuildExists(resourceName, &conf),
					testAccCheckAWSGameliftBuildDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSGameliftBuildExists(n string, res *gamelift.Build) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Gamelift Build ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).gameliftconn

		req := &gamelift.DescribeBuildInput{
			BuildId: aws.String(rs.Primary.ID),
		}
		out, err := conn.DescribeBuild(req)
		if err != nil {
			return err
		}

		b := out.Build

		if *b.BuildId != rs.Primary.ID {
			return fmt.Errorf("Gamelift Build not found")
		}

		*res = *b

		return nil
	}
}

func testAccCheckAWSGameliftBuildDisappears(res *gamelift.Build) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).gameliftconn

		input := &gamelift.DeleteBuildInput{BuildId: res.BuildId}

		_, err := conn.DeleteBuild(input)
		return err
	}
}

func testAccCheckAWSGameliftBuildDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).gameliftconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_gamelift_build" {
			continue
		}

		req := gamelift.DescribeBuildInput{
			BuildId: aws.String(rs.Primary.ID),
		}
		out, err := conn.DescribeBuild(&req)
		if err == nil {
			if *out.Build.BuildId == rs.Primary.ID {
				return fmt.Errorf("Gamelift Build still exists")
			}
		}
		if isAWSErr(err, gamelift.ErrCodeNotFoundException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccPreCheckAWSGamelift(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).gameliftconn

	input := &gamelift.ListBuildsInput{}

	_, err := conn.ListBuilds(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSGameliftBuildBasicConfig(buildName, bucketName, key, roleArn string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_build" "test" {
  name             = "%s"
  operating_system = "WINDOWS_2012"

  storage_location {
    bucket   = "%s"
    key      = "%s"
    role_arn = "%s"
  }
}
`, buildName, bucketName, key, roleArn)
}

func testAccAWSGameliftBuildBasicConfigTags1(buildName, bucketName, key, roleArn, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_build" "test" {
  name             = %[1]q
  operating_system = "WINDOWS_2012"

  storage_location {
    bucket   = %[2]q
    key      = %[3]q
    role_arn = %[4]q
  }

  tags = {
    %[5]q = %[6]q
  }
}
`, buildName, bucketName, key, roleArn, tagKey1, tagValue1)
}

func testAccAWSGameliftBuildBasicConfigTags2(buildName, bucketName, key, roleArn, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_build" "test" {
  name             = %[1]q
  operating_system = "WINDOWS_2012"

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
`, buildName, bucketName, key, roleArn, tagKey1, tagValue1, tagKey2, tagValue2)
}
