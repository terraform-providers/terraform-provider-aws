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

func TestAccAWSMediaStoreContainerLifecyclePolicy_basic(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_media_store_container_lifecycle_policy.test"
	containerResourceName := "aws_media_store_container.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMediaStore(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaStoreContainerLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaStoreContainerLifecyclePolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaStoreContainerLifecyclePolicyExists(resourceName),
					resource.TestCheckResourceAttrPair(containerResourceName, "name", resourceName, "container_name"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
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

func TestAccAWSMediaStoreContainerLifecyclePolicy_disappears(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_media_store_container_lifecycle_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMediaStore(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsMediaStoreContainerLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaStoreContainerLifecyclePolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsMediaStoreContainerLifecyclePolicyExists(resourceName),
					testAccCheckAwsMediaStoreContainerLifecyclePolicyDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsMediaStoreContainerLifecyclePolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).mediastoreconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_media_store_container_lifecycle_policy" {
			continue
		}

		input := &mediastore.GetLifecyclePolicyInput{
			ContainerName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetLifecyclePolicy(input)
		if err != nil {
			if isAWSErr(err, mediastore.ErrCodeContainerNotFoundException, "") {
				continue
			}
			if isAWSErr(err, mediastore.ErrCodePolicyNotFoundException, "") {
				continue
			}
			return err
		}

		return fmt.Errorf("Expected MediaStore Container Lifecycle Policy to be destroyed, %s found", rs.Primary.ID)
	}
	return nil
}

func testAccCheckAwsMediaStoreContainerLifecyclePolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).mediastoreconn

		input := &mediastore.GetLifecyclePolicyInput{
			ContainerName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetLifecyclePolicy(input)

		return err
	}
}

func testAccCheckAwsMediaStoreContainerLifecyclePolicyDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).mediastoreconn

		input := &mediastore.DeleteLifecyclePolicyInput{
			ContainerName: aws.String(rs.Primary.ID),
		}

		if _, err := conn.DeleteLifecyclePolicy(input); err != nil {
			return err
		}

		return nil
	}
}

func testAccMediaStoreContainerLifecyclePolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_media_store_container" "test" {
  name = "tf_mediastore_%s"
}

resource "aws_media_store_container_lifecycle_policy" "test" {
  container_name = aws_media_store_container.test.id
  policy         = file("test-fixtures/mediastore_lifecycle_policy.json")
}

`, rName)
}
