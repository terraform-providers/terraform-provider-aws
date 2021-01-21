package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceAWSS3BucketObjects_basic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3ObjectsConfigResources(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccAWSDataSourceS3ObjectsConfigBasic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsS3ObjectsDataSourceExists("data.aws_s3_bucket_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.#", "2"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.0", "arch/navajo/north_window"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.1", "arch/navajo/sand_dune"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSS3BucketObjects_basicViaAccessPoint(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3ObjectsConfigResourcesPlusAccessPoint(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccAWSDataSourceS3ObjectsConfigBasicViaAccessPoint(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsS3ObjectsDataSourceExists("data.aws_s3_bucket_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.#", "2"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.0", "arch/navajo/north_window"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.1", "arch/navajo/sand_dune"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSS3BucketObjects_all(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3ObjectsConfigResources(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccAWSDataSourceS3ObjectsConfigAll(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsS3ObjectsDataSourceExists("data.aws_s3_bucket_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.#", "7"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.0", "arch/courthouse_towers/landscape"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.1", "arch/navajo/north_window"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.2", "arch/navajo/sand_dune"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.3", "arch/partition/park_avenue"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.4", "arch/rubicon"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.5", "arch/three_gossips/broken"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.6", "arch/three_gossips/turret"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSS3BucketObjects_prefixes(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3ObjectsConfigResources(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccAWSDataSourceS3ObjectsConfigPrefixes(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsS3ObjectsDataSourceExists("data.aws_s3_bucket_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.#", "1"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.0", "arch/rubicon"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "common_prefixes.#", "4"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "common_prefixes.0", "arch/courthouse_towers/"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "common_prefixes.1", "arch/navajo/"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "common_prefixes.2", "arch/partition/"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "common_prefixes.3", "arch/three_gossips/"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSS3BucketObjects_encoded(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3ObjectsConfigExtraResource(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccAWSDataSourceS3ObjectsConfigEncoded(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsS3ObjectsDataSourceExists("data.aws_s3_bucket_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.#", "2"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.0", "arch/ru+b+ic+on"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.1", "arch/rubicon"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSS3BucketObjects_maxKeys(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3ObjectsConfigResources(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccAWSDataSourceS3ObjectsConfigMaxKeys(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsS3ObjectsDataSourceExists("data.aws_s3_bucket_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.#", "2"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.0", "arch/courthouse_towers/landscape"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.1", "arch/navajo/north_window"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSS3BucketObjects_startAfter(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3ObjectsConfigResources(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccAWSDataSourceS3ObjectsConfigStartAfter(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsS3ObjectsDataSourceExists("data.aws_s3_bucket_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.#", "1"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.0", "arch/three_gossips/turret"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSS3BucketObjects_fetchOwner(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3ObjectsConfigResources(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccAWSDataSourceS3ObjectsConfigOwners(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsS3ObjectsDataSourceExists("data.aws_s3_bucket_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.#", "2"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "owners.#", "2"),
				),
			},
		},
	})
}

func testAccCheckAwsS3ObjectsDataSourceExists(addr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("Can't find S3 objects data source: %s", addr)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("S3 objects data source ID not set")
		}

		return nil
	}
}

func testAccAWSDataSourceS3ObjectsConfigResources(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "objects_bucket" {
  bucket = "tf-acc-objects-test-bucket-%d"
}

resource "aws_s3_bucket_object" "object1" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/three_gossips/turret"
  content = "Delicate"
}

resource "aws_s3_bucket_object" "object2" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/three_gossips/broken"
  content = "Dark Angel"
}

resource "aws_s3_bucket_object" "object3" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/navajo/north_window"
  content = "Balanced Rock"
}

resource "aws_s3_bucket_object" "object4" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/navajo/sand_dune"
  content = "Queen Victoria Rock"
}

resource "aws_s3_bucket_object" "object5" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/partition/park_avenue"
  content = "Double-O"
}

resource "aws_s3_bucket_object" "object6" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/courthouse_towers/landscape"
  content = "Fiery Furnace"
}

resource "aws_s3_bucket_object" "object7" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/rubicon"
  content = "Devils Garden"
}
`, randInt)
}

func testAccAWSDataSourceS3ObjectsConfigResourcesPlusAccessPoint(randInt int) string {
	return testAccAWSDataSourceS3ObjectsConfigResources(randInt) + fmt.Sprintf(`
resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.objects_bucket.bucket
  name   = "tf-objects-test-access-point-%[1]d"
}
`, randInt)
}

func testAccAWSDataSourceS3ObjectsConfigBasic(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_bucket_objects" "yesh" {
  bucket    = aws_s3_bucket.objects_bucket.id
  prefix    = "arch/navajo/"
  delimiter = "/"
}
`, testAccAWSDataSourceS3ObjectsConfigResources(randInt))
}

func testAccAWSDataSourceS3ObjectsConfigBasicViaAccessPoint(randInt int) string {
	return testAccAWSDataSourceS3ObjectsConfigResourcesPlusAccessPoint(randInt) + `
data "aws_s3_bucket_objects" "yesh" {
  bucket    = aws_s3_access_point.test.arn
  prefix    = "arch/navajo/"
  delimiter = "/"
}
`
}

func testAccAWSDataSourceS3ObjectsConfigAll(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_bucket_objects" "yesh" {
  bucket = aws_s3_bucket.objects_bucket.id
}
`, testAccAWSDataSourceS3ObjectsConfigResources(randInt))
}

func testAccAWSDataSourceS3ObjectsConfigPrefixes(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_bucket_objects" "yesh" {
  bucket    = aws_s3_bucket.objects_bucket.id
  prefix    = "arch/"
  delimiter = "/"
}
`, testAccAWSDataSourceS3ObjectsConfigResources(randInt))
}

func testAccAWSDataSourceS3ObjectsConfigExtraResource(randInt int) string {
	return fmt.Sprintf(`
%s

resource "aws_s3_bucket_object" "object8" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/ru b ic on"
  content = "Goose Island"
}
`, testAccAWSDataSourceS3ObjectsConfigResources(randInt))
}

func testAccAWSDataSourceS3ObjectsConfigEncoded(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_bucket_objects" "yesh" {
  bucket        = aws_s3_bucket.objects_bucket.id
  encoding_type = "url"
  prefix        = "arch/ru"
}
`, testAccAWSDataSourceS3ObjectsConfigExtraResource(randInt))
}

func testAccAWSDataSourceS3ObjectsConfigMaxKeys(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_bucket_objects" "yesh" {
  bucket   = aws_s3_bucket.objects_bucket.id
  max_keys = 2
}
`, testAccAWSDataSourceS3ObjectsConfigResources(randInt))
}

func testAccAWSDataSourceS3ObjectsConfigStartAfter(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_bucket_objects" "yesh" {
  bucket      = aws_s3_bucket.objects_bucket.id
  start_after = "arch/three_gossips/broken"
}
`, testAccAWSDataSourceS3ObjectsConfigResources(randInt))
}

func testAccAWSDataSourceS3ObjectsConfigOwners(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_bucket_objects" "yesh" {
  bucket      = aws_s3_bucket.objects_bucket.id
  prefix      = "arch/three_gossips/"
  fetch_owner = true
}
`, testAccAWSDataSourceS3ObjectsConfigResources(randInt))
}
