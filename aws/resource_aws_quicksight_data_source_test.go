package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/quicksight"
)

func TestAccAWSQuickSightDataSource_basic(t *testing.T) {
	var dataSource quicksight.DataSource
	resourceName := "aws_quicksight_data_source.default"
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rId1 := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckQuickSightDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQuickSightDataSourceConfig(rId1, rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSourceExists(resourceName, &dataSource),
					resource.TestCheckResourceAttr(resourceName, "data_source_id", rId1),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "quicksight", fmt.Sprintf("datasource/%s", rId1)),
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

func TestAccAWSQuickSightDataSource_disappears(t *testing.T) {
	var dataSource quicksight.DataSource
	resourceName := "aws_quicksight_data_source.default"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rId := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckQuickSightDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQuickSightDataSourceConfig(rId, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQuickSightDataSourceExists(resourceName, &dataSource),
					testAccCheckQuickSightDataSourceDisappears(&dataSource),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckQuickSightDataSourceExists(resourceName string, dataSource *quicksight.DataSource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		awsAccountID, dataSourceId, err := resourceAwsQuickSightDataSourceParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).quicksightconn

		input := &quicksight.DescribeDataSourceInput{
			AwsAccountId: aws.String(awsAccountID),
			DataSourceId: aws.String(dataSourceId),
		}

		output, err := conn.DescribeDataSource(input)

		if err != nil {
			return err
		}

		if output == nil || output.DataSource == nil {
			return fmt.Errorf("QuickSight Data Source (%s) not found", rs.Primary.ID)
		}

		*dataSource = *output.DataSource

		return nil
	}
}

func testAccCheckQuickSightDataSourceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).quicksightconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_quicksight_data_source" {
			continue
		}

		awsAccountID, dataSourceId, err := resourceAwsQuickSightDataSourceParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.DescribeDataSource(&quicksight.DescribeDataSourceInput{
			AwsAccountId: aws.String(awsAccountID),
			DataSourceId: aws.String(dataSourceId),
		})
		if isAWSErr(err, quicksight.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("QuickSight Data Source '%s' was not deleted properly", rs.Primary.ID)
	}

	return nil
}

func testAccCheckQuickSightDataSourceDisappears(v *quicksight.DataSource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).quicksightconn

		arn, err := arn.Parse(aws.StringValue(v.Arn))
		if err != nil {
			return err
		}

		input := &quicksight.DeleteDataSourceInput{
			AwsAccountId: aws.String(arn.AccountID),
			DataSourceId: v.DataSourceId,
		}

		if _, err := conn.DeleteDataSource(input); err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSQuickSightDataSourceConfig(rId string, rName string) string {
	manifestKey := acctest.RandomWithPrefix("tf-acc-test")

	return fmt.Sprintf(`
resource "aws_quicksight_data_source" "default" {
  data_source_id = %[1]q
  name           = %[2]q

  parameters {
    s3 {
      manifest_file_location {
        bucket = aws_s3_bucket.bucket.bucket
        key = aws_s3_bucket_object.manifest.key
      }
    }
  }
}

resource "aws_s3_bucket" "bucket" {
  acl = "public-read"
}

resource "aws_s3_bucket_object" "manifest" {
  bucket  = aws_s3_bucket.bucket.bucket
  key     = %[3]q
  content = <<EOF
{
  "fileLocations": [
      {
          "URIs": [
              "https://${aws_s3_bucket.bucket.bucket}.s3.amazonaws.com/%s"
          ]
      }
  ],
  "globalUploadSettings": {
      "format": "JSON"
  }
}
EOF

  acl = "public-read"
}

`, rId, rName, manifestKey, manifestKey)
}
