package aws

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSEcrDataSource_ecrImage(t *testing.T) {
	registry, repo, tag := "137112412989", "amazonlinux", "latest"
	resourceByTag := "data.aws_ecr_image.by_tag"
	resourceByDigest := "data.aws_ecr_image.by_digest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEcrImageDataSourceConfig(registry, repo, tag),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceByTag, "image_digest"),
					resource.TestCheckResourceAttrSet(resourceByTag, "image_pushed_at"),
					resource.TestCheckResourceAttrSet(resourceByTag, "image_size_in_bytes"),
					testCheckTagInImageTags(resourceByTag, tag),
					resource.TestCheckResourceAttrSet(resourceByDigest, "image_pushed_at"),
					resource.TestCheckResourceAttrSet(resourceByDigest, "image_size_in_bytes"),
					testCheckTagInImageTags(resourceByDigest, tag),
				),
			},
		},
	})
}

func testAccCheckAwsEcrImageDataSourceConfig(reg, repo, tag string) string {
	return fmt.Sprintf(`
data "aws_ecr_image" "by_tag" {
  registry_id     = "%s"
  repository_name = "%s"
  image_tag       = "%s"
}

data "aws_ecr_image" "by_digest" {
  registry_id     = data.aws_ecr_image.by_tag.registry_id
  repository_name = data.aws_ecr_image.by_tag.repository_name
  image_digest    = data.aws_ecr_image.by_tag.image_digest
}
`, reg, repo, tag)
}

func testCheckTagInImageTags(name, expectedTag string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Resource not found: %s", name)
		}

		tagsLenStr, ok := rs.Primary.Attributes["image_tags.#"]
		if !ok {
			return fmt.Errorf("No attribute 'image_tags' in resource: %s", name)
		}
		tagsLen, _ := strconv.Atoi(tagsLenStr)

		for i := 0; i < tagsLen; i++ {
			tag := rs.Primary.Attributes[fmt.Sprintf("image_tags.%d", i)]
			if tag == expectedTag {
				return nil
			}
		}
		return fmt.Errorf("No tag '%s' in images_tags of resource %s", expectedTag, name)
	}
}
