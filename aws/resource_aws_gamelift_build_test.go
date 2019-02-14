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
		if !strings.HasPrefix(*build.Name, testAccGameliftBuildPrefix) {
			continue
		}

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

	buildName := fmt.Sprintf("%s_%s", testAccGameliftBuildPrefix, rString)
	uBuildName := fmt.Sprintf("%s_updated_%s", testAccGameliftBuildPrefix, rString)

	region := testAccGetRegion()
	g, err := testAccAWSGameliftSampleGame(region)
	if err != nil {
		t.Fatal(err)
	}

	loc := g.Location
	bucketName := *loc.Bucket
	roleArn := *loc.RoleArn
	key := *loc.Key

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGameliftBuildDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGameliftBuildBasicConfig(buildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftBuildExists("aws_gamelift_build.test", &conf),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "name", buildName),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "operating_system", "WINDOWS_2012"),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "storage_location.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "storage_location.0.bucket", bucketName),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "storage_location.0.key", key),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "storage_location.0.role_arn", roleArn),
				),
			},
			{
				Config: testAccAWSGameliftBuildBasicConfig(uBuildName, bucketName, key, roleArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGameliftBuildExists("aws_gamelift_build.test", &conf),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "name", uBuildName),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "operating_system", "WINDOWS_2012"),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "storage_location.#", "1"),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "storage_location.0.bucket", bucketName),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "storage_location.0.key", key),
					resource.TestCheckResourceAttr("aws_gamelift_build.test", "storage_location.0.role_arn", roleArn),
				),
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

func testAccAWSGameliftBuildBasicConfig(buildName, bucketName, key, roleArn string) string {
	return fmt.Sprintf(`
resource "aws_gamelift_build" "test" {
  name = "%s"
  operating_system = "WINDOWS_2012"
  storage_location {
    bucket = "%s"
    key = "%s"
    role_arn = "%s"
  }
}
`, buildName, bucketName, key, roleArn)
}
