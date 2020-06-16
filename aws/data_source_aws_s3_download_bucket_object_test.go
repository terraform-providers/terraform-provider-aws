package aws

import (
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDataSourceAWSS3DownloadBucketObject_basic(t *testing.T) {
	dirName, _ := ioutil.TempDir("", "")
	fileName := path.Join(dirName, "testfile.txt")
	rInt := acctest.RandInt()
	resourceOnlyConf, conf := testAccAWSDataSourceS3DownloadObjectConfig_basic(rInt, fileName)

	var rObj s3.GetObjectOutput

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: resourceOnlyConf,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketObjectExists("aws_s3_bucket_object.object", &rObj),
				),
			},
			{
				Config: conf,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsS3DownloadedObjectExists(fileName, "Hello World"),
					resource.TestCheckResourceAttr("data.aws_s3_download_bucket_object.obj", "filename", fileName),
				),
			},
		},
	})
}

func testAccCheckAwsS3DownloadedObjectExists(fileName string, expectedContent string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		content, err := ioutil.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("Error while reading the downloaded file: %s", err)
		}

		gottenContent := string(content)
		if gottenContent != expectedContent {
			return fmt.Errorf("Gotten content was `%s`, not `%s` as expected", gottenContent, expectedContent)
		}

		return nil

	}

}

func testAccAWSDataSourceS3DownloadObjectConfig_basic(randInt int, fileName string) (string, string) {
	resources := fmt.Sprintf(`
resource "aws_s3_bucket" "object_bucket" {
	bucket = "tf-object-test-bucket-%d"
}
resource "aws_s3_bucket_object" "object" {
	bucket = "${aws_s3_bucket.object_bucket.bucket}"
	key = "tf-testing-obj-%d"
	content = "Hello World"
}
`, randInt, randInt)

	both := fmt.Sprintf(`%s
data "aws_s3_download_bucket_object" "obj" {
	bucket = "tf-object-test-bucket-%d"
	key = "tf-testing-obj-%d"
	filename = "%s"
}
`, resources, randInt, randInt, fileName)

	return resources, both
}
