package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_cloudfront_key_group", &resource.Sweeper{
		Name: "aws_cloudfront_key_group",
		F:    testSweepCloudFrontKeyGroup,
	})
}

func testSweepCloudFrontKeyGroup(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*AWSClient).cloudfrontconn
	var sweeperErrs *multierror.Error

	input := &cloudfront.ListKeyGroupsInput{}

	for {
		output, err := conn.ListKeyGroups(input)
		if err != nil {
			if testSweepSkipSweepError(err) {
				log.Printf("[WARN] Skipping CloudFront key group sweep for %s: %s", region, err)
				return nil
			}
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving CloudFront key group: %w", err))
			return sweeperErrs.ErrorOrNil()
		}

		if output == nil || output.KeyGroupList == nil || len(output.KeyGroupList.Items) == 0 {
			log.Print("[DEBUG] No CloudFront key group to sweep")
			return nil
		}

		for _, item := range output.KeyGroupList.Items {
			strId := aws.StringValue(item.KeyGroup.Id)
			log.Printf("[INFO] CloudFront key group %s", strId)
			_, err := conn.DeleteKeyGroup(&cloudfront.DeleteKeyGroupInput{
				Id: item.KeyGroup.Id,
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting CloudFront key group %s: %w", strId, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if output.KeyGroupList.NextMarker == nil {
			break
		}
		input.Marker = output.KeyGroupList.NextMarker
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSCloudFrontKeyGroup_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudfront_key_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontKeyGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontKeyGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontKeyGroupExistence(resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "comment", "test key group"),
					resource.TestCheckResourceAttrSet("aws_cloudfront_key_group.test", "etag"),
					resource.TestCheckResourceAttrSet("aws_cloudfront_key_group.test", "id"),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "items.#", "1"),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "name", rName),
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

func TestAccAWSCloudFrontKeyGroup_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudfront_key_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontKeyGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontKeyGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontKeyGroupExistence(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudFrontKeyGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudFrontKeyGroup_Comment(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudfront_key_group.test"

	firstComment := "first comment"
	secondComment := "second comment"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontKeyGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontKeyGroupConfigComment(rName, firstComment),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontKeyGroupExistence(resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "comment", firstComment),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudFrontKeyGroupConfigComment(rName, secondComment),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontKeyGroupExistence(resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "comment", secondComment),
				),
			},
		},
	})
}

func TestAccAWSCloudFrontKeyGroup_Items(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudfront_key_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, cloudfront.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontKeyGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontKeyGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontKeyGroupExistence(resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "items.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudFrontKeyGroupConfigItems(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontKeyGroupExistence(resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "items.#", "2"),
				),
			},
		},
	})
}

func testAccCheckCloudFrontKeyGroupExistence(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("not found: %s", r)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Id is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

		input := &cloudfront.GetKeyGroupInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetKeyGroup(input)
		if err != nil {
			return fmt.Errorf("error retrieving CloudFront key group: %s", err)
		}
		return nil
	}
}

func testAccCheckCloudFrontKeyGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_key_group" {
			continue
		}

		input := &cloudfront.GetKeyGroupInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetKeyGroup(input)
		if isAWSErr(err, cloudfront.ErrCodeNoSuchResource, "") {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("CloudFront key group (%s) was not deleted", rs.Primary.ID)
	}

	return nil
}

func testAccAWSCloudFrontKeyGroupConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test" {
  comment     = "test key"
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
  name        = %q
}
`, rName)
}

func testAccAWSCloudFrontKeyGroupConfig(rName string) string {
	return testAccAWSCloudFrontKeyGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_cloudfront_key_group" "test" {
  comment = "test key group"
  items   = [aws_cloudfront_public_key.test.id]
  name    = %q
}
`, rName)
}

func testAccAWSCloudFrontKeyGroupConfigComment(rName string, comment string) string {
	return testAccAWSCloudFrontKeyGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_cloudfront_key_group" "test" {
  comment = %q
  items   = [aws_cloudfront_public_key.test.id]
  name    = %q
}
`, comment, rName)
}

func testAccAWSCloudFrontKeyGroupConfigItems(rName string) string {
	return testAccAWSCloudFrontKeyGroupConfigBase(rName) + fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test2" {
  comment     = "second test key"
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
  name        = "%[1]s-second"
}

resource "aws_cloudfront_key_group" "test" {
  comment = "test key group"
  items   = [aws_cloudfront_public_key.test.id, aws_cloudfront_public_key.test2.id]
  name    = %[1]q
}
`, rName)
}
