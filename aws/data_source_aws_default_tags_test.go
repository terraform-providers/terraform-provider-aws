package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccAWSDefaultTagsDataSource_basic(t *testing.T) {
	var providers []*schema.Provider

	dataSourceName := "data.aws_default_tags.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("first", "value"),
					testAccAWSDefaultTagsDataSource(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.first", "value"),
				),
			},
		},
	})
}

func TestAccAWSDefaultTagsDataSource_empty(t *testing.T) {
	var providers []*schema.Provider

	dataSourceName := "data.aws_default_tags.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags0(),
					testAccAWSDefaultTagsDataSource(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSDefaultTagsDataSource_multiple(t *testing.T) {
	var providers []*schema.Provider

	dataSourceName := "data.aws_default_tags.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags2("nuera", "hijo", "escalofrios", "calambres"),
					testAccAWSDefaultTagsDataSource(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.nuera", "hijo"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.escalofrios", "calambres"),
				),
			},
		},
	})
}

func TestAccAWSDefaultTagsDataSource_ignore(t *testing.T) {
	var providers []*schema.Provider

	dataSourceName := "data.aws_default_tags.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesInternal(&providers),
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSProviderConfigDefaultTags_Tags1("Tabac", "Louis Chiron"),
					testAccAWSDefaultTagsDataSource(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.Tabac", "Louis Chiron"),
				),
			},
			{
				Config: composeConfig(
					testAccProviderConfigDefaultAndIgnoreTagsKeys1("Tabac", "Louis Chiron"),
					testAccAWSDefaultTagsDataSource(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccAWSDefaultTagsDataSource() string {
	return `data "aws_default_tags" "test" {}`
}
