package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSEbsSnapshotCopy_basic(t *testing.T) {
	var v ec2.Snapshot
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotCopyConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExists("aws_ebs_snapshot_copy.test", &v),
					testAccCheckTags(&v.Tags, "Name", "testAccAwsEbsSnapshotCopyConfig"),
				),
			},
		},
	})
}

func TestAccAWSEbsSnapshotCopy_withDescription(t *testing.T) {
	var v ec2.Snapshot
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotCopyConfigWithDescription,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExists("aws_ebs_snapshot_copy.description_test", &v),
					resource.TestCheckResourceAttr("aws_ebs_snapshot_copy.description_test", "description", "Copy Snapshot Acceptance Test"),
				),
			},
		},
	})
}

func TestAccAWSEbsSnapshotCopy_withRegions(t *testing.T) {
	var v ec2.Snapshot

	// record the initialized providers so that we can use them to
	// check for the instances in each region
	var providers []*schema.Provider
	providerFactories := map[string]terraform.ResourceProviderFactory{
		"aws": func() (terraform.ResourceProvider, error) {
			p := Provider()
			providers = append(providers, p.(*schema.Provider))
			return p, nil
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotCopyConfigWithRegions,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExistsWithProviders("aws_ebs_snapshot_copy.region_test", &v, &providers),
				),
			},
		},
	})

}

func TestAccAWSEbsSnapshotCopy_withKms(t *testing.T) {
	var v ec2.Snapshot
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEbsSnapshotCopyConfigWithKms,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEbsSnapshotCopyExists("aws_ebs_snapshot_copy.kms_test", &v),
					resource.TestMatchResourceAttr("aws_ebs_snapshot_copy.kms_test", "kms_key_id",
						regexp.MustCompile(`^arn:aws:kms:[a-z]{2}-[a-z]+-\d{1}:[0-9]{12}:key/[a-z0-9-]{36}$`)),
				),
			},
		},
	})
}

func testAccCheckEbsSnapshotCopyExists(n string, v *ec2.Snapshot) resource.TestCheckFunc {
	providers := []*schema.Provider{testAccProvider}
	return testAccCheckEbsSnapshotCopyExistsWithProviders(n, v, &providers)
}

func testAccCheckEbsSnapshotCopyExistsWithProviders(n string, v *ec2.Snapshot, providers *[]*schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		for _, provider := range *providers {
			// Ignore if Meta is empty, this can happen for validation providers
			if provider.Meta() == nil {
				continue
			}

			conn := provider.Meta().(*AWSClient).ec2conn

			request := &ec2.DescribeSnapshotsInput{
				SnapshotIds: []*string{aws.String(rs.Primary.ID)},
			}

			response, err := conn.DescribeSnapshots(request)
			if err == nil {
				if response.Snapshots != nil && len(response.Snapshots) > 0 {
					*v = *response.Snapshots[0]
					return nil
				}
			}
		}
		return fmt.Errorf("Error finding EC2 Snapshot %s", rs.Primary.ID)
	}
}

const testAccAwsEbsSnapshotCopyConfig = `
resource "aws_ebs_volume" "test" {
  availability_zone = "us-west-2a"
  size              = 1
}

resource "aws_ebs_snapshot" "test" {
  volume_id = "${aws_ebs_volume.test.id}"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfig"
  }
}

resource "aws_ebs_snapshot_copy" "test" {
  source_snapshot_id = "${aws_ebs_snapshot.test.id}"
	source_region      = "us-west-2"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfig"
  }
}
`

const testAccAwsEbsSnapshotCopyConfigWithDescription = `
resource "aws_ebs_volume" "description_test" {
	availability_zone = "us-west-2a"
	size              = 1

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfigWithDescription"
  }
}

resource "aws_ebs_snapshot" "description_test" {
	volume_id   = "${aws_ebs_volume.description_test.id}"
	description = "EBS Snapshot Acceptance Test"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfigWithDescription"
  }
}

resource "aws_ebs_snapshot_copy" "description_test" {
	description        = "Copy Snapshot Acceptance Test"
  source_snapshot_id = "${aws_ebs_snapshot.description_test.id}"
	source_region      = "us-west-2"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfigWithDescription"
  }
}
`

const testAccAwsEbsSnapshotCopyConfigWithRegions = `
provider "aws" {
  region = "us-west-2"
  alias  = "uswest2"
}

provider "aws" {
  region = "us-east-1"
  alias  = "useast1"
}

resource "aws_ebs_volume" "region_test" {
  provider          = "aws.uswest2"
  availability_zone = "us-west-2a"
  size              = 1

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfigWithRegions"
  }
}

resource "aws_ebs_snapshot" "region_test" {
  provider  = "aws.uswest2"
  volume_id = "${aws_ebs_volume.region_test.id}"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfigWithRegions"
  }
}

resource "aws_ebs_snapshot_copy" "region_test" {
  provider           = "aws.useast1"
  source_snapshot_id = "${aws_ebs_snapshot.region_test.id}"
  source_region      = "us-west-2"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfigWithRegions"
  }
}
`

const testAccAwsEbsSnapshotCopyConfigWithKms = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_kms_key" "kms_test" {
  description             = "testAccAwsEbsSnapshotCopyConfigWithKms"
  deletion_window_in_days = 7
}

resource "aws_ebs_volume" "kms_test" {
  availability_zone = "us-west-2a"
  size              = 1

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfigWithKms"
  }
}

resource "aws_ebs_snapshot" "kms_test" {
  volume_id = "${aws_ebs_volume.kms_test.id}"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfigWithKms"
  }
}

resource "aws_ebs_snapshot_copy" "kms_test" {
  source_snapshot_id = "${aws_ebs_snapshot.kms_test.id}"
  source_region      = "us-west-2"
  encrypted          = true
  kms_key_id         = "${aws_kms_key.kms_test.arn}"

  tags = {
    Name = "testAccAwsEbsSnapshotCopyConfigWithKms"
  }
}
`
