package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSMediaStoreContainerCorsPolicy_basic(t *testing.T) {
	resourceName := "aws_media_store_container_cors_policy.test"
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaStoreContainerCorsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaStoreContainerCorsPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaStoreContainerCorsPolicyExists(resourceName),
					resource.TestCheckResourceAttrPair(
						resourceName, "container_name",
						"aws_media_store_container.test", "name"),
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

func TestAccAWSMediaStoreContainerCorsPolicy_optional(t *testing.T) {
	resourceName := "aws_media_store_container_cors_policy.test"
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaStoreContainerCorsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaStoreContainerCorsPolicyConfig_optional(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaStoreContainerCorsPolicyExists(resourceName),
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

func TestAccAWSMediaStoreContainerCorsPolicy_update(t *testing.T) {
	resourceName := "aws_media_store_container_cors_policy.test"
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaStoreContainerCorsPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaStoreContainerCorsPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaStoreContainerCorsPolicyExists(resourceName),
				),
			},
			{
				Config: testAccMediaStoreContainerCorsPolicyConfig_update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaStoreContainerCorsPolicyExists(resourceName),
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

func testAccCheckAwsMediaStoreContainerCorsPolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).mediastoreconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_media_store_container_cors_policy" {
			continue
		}

		input := &mediastore.GetCorsPolicyInput{
			ContainerName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetCorsPolicy(input)
		if err != nil {
			if isAWSErr(err, mediastore.ErrCodeCorsPolicyNotFoundException, "") || isAWSErr(err, mediastore.ErrCodeContainerNotFoundException, "") {
				return nil
			}
			return err
		}
		return fmt.Errorf("Media Store Container Cors Policy(%s) is stil exists.", rs.Primary.ID)
	}
	return nil
}

func testAccCheckAwsMediaStoreContainerCorsPolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).mediastoreconn

		input := &mediastore.GetCorsPolicyInput{
			ContainerName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetCorsPolicy(input)

		return err
	}
}

func testAccMediaStoreContainerCorsPolicyConfig(rName string) string {
	return fmt.Sprintf(`
%s

resource "aws_media_store_container_cors_policy" "test" {
	container_name = "${aws_media_store_container.test.name}"

	cors_policy {
		allowed_headers = ["*"]
		allowed_methods = ["GET"]
		allowed_origins = ["*"]
	}
}`, testAccMediaStoreContainerConfig(rName))
}

func testAccMediaStoreContainerCorsPolicyConfig_optional(rName string) string {
	return fmt.Sprintf(`
%s

resource "aws_media_store_container_cors_policy" "test" {
	container_name = "${aws_media_store_container.test.name}"

	cors_policy {
		allowed_headers = ["*"]
		allowed_methods = ["GET", "HEAD"]
		allowed_origins = ["http://aaa.example.com"]
		expose_headers 	= ["*"]
		max_age_seconds = 0
	}
}`, testAccMediaStoreContainerConfig(rName))
}

func testAccMediaStoreContainerCorsPolicyConfig_update(rName string) string {
	return fmt.Sprintf(`
%s

resource "aws_media_store_container_cors_policy" "test" {
	container_name = "${aws_media_store_container.test.name}"

	cors_policy {
		allowed_headers = ["Access-Control-Request-Headers"]
		allowed_methods = ["PUT", "GET"]
		allowed_origins = ["http://aaa.example.com", "http://bbb.example.com"]
		expose_headers 	= ["XMLHttpRequest"]
		max_age_seconds = 3600
	}
}`, testAccMediaStoreContainerConfig(rName))
}
