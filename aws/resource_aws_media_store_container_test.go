package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSMediaStoreContainer_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaStoreContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaStoreContainerConfig(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaStoreContainerExists("aws_media_store_container.test"),
				),
			},
		},
	})
}

func TestAccAWSMediaStoreContainer_import(t *testing.T) {
	resourceName := "aws_media_store_container.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaStoreContainerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaStoreContainerConfig(acctest.RandString(5)),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsMediaStoreContainerDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).mediastoreconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_media_store_container" {
			continue
		}

		input := &mediastore.DescribeContainerInput{
			ContainerName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeContainer(input)
		if err != nil {
			if isAWSErr(err, mediastore.ErrCodeContainerNotFoundException, "") {
				return nil
			}
			return err
		}

		if *resp.Container.Status != mediastore.ContainerStatusDeleting {
			return fmt.Errorf("Media Store Container (%s) not deleted", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckAwsMediaStoreContainerExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).mediastoreconn

		input := &mediastore.DescribeContainerInput{
			ContainerName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeContainer(input)

		return err
	}
}

func testAccMediaStoreContainerConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_store_container" "test" {
  name = "tf_mediastore_%s"
}`, rName)
}
