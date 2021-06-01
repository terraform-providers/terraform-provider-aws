package aws

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSOutpostsOutpostDataSource_Id(t *testing.T) {
	dataSourceName := "data.aws_outposts_outpost.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSOutpostsOutpostDataSourceConfigId(),
				Check: resource.ComposeTestCheckFunc(
					testAccMatchResourceAttrRegionalARN(dataSourceName, "arn", "outposts", regexp.MustCompile(`outpost/op-.+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "availability_zone", regexp.MustCompile(`^.+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "availability_zone_id", regexp.MustCompile(`^.+$`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "description"),
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile(`^op-.+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "name", regexp.MustCompile(`^.+$`)),
					testAccCheckResourceAttrAccountID(dataSourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccAWSOutpostsOutpostDataSource_Name(t *testing.T) {
	sourceDataSourceName := "data.aws_outposts_outpost.source"
	dataSourceName := "data.aws_outposts_outpost.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSOutpostsOutpostDataSourceConfigName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", sourceDataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone", sourceDataSourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_id", sourceDataSourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", sourceDataSourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", sourceDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", sourceDataSourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", sourceDataSourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccAWSOutpostsOutpostDataSource_Arn(t *testing.T) {
	sourceDataSourceName := "data.aws_outposts_outpost.source"
	dataSourceName := "data.aws_outposts_outpost.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSOutpostsOutpostDataSourceConfigArn(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", sourceDataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone", sourceDataSourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_id", sourceDataSourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", sourceDataSourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", sourceDataSourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", sourceDataSourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", sourceDataSourceName, "owner_id"),
				),
			},
		},
	})
}

func testAccAWSOutpostsOutpostDataSourceConfigId() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}
`
}

func testAccAWSOutpostsOutpostDataSourceConfigName() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "source" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

data "aws_outposts_outpost" "test" {
  name = data.aws_outposts_outpost.source.name
}
`
}

func testAccAWSOutpostsOutpostDataSourceConfigArn() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "source" {
  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

data "aws_outposts_outpost" "test" {
  arn = data.aws_outposts_outpost.source.arn
}
`
}
