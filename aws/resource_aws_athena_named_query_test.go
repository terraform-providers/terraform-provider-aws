package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAthenaNamedQuery_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaNamedQueryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaNamedQueryConfig(acctest.RandInt(), acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAthenaNamedQueryExists("aws_athena_named_query.foo"),
				),
			},
		},
	})
}

func TestAccAWSAthenaNamedQuery_import(t *testing.T) {
	resourceName := "aws_athena_named_query.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAthenaNamedQueryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaNamedQueryConfig(acctest.RandInt(), acctest.RandString(5)),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSAthenaNamedQueryDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).athenaconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_athena_named_query" {
			continue
		}

		input := &athena.GetNamedQueryInput{
			NamedQueryId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetNamedQuery(input)
		if err != nil {
			if isAWSErr(err, athena.ErrCodeInvalidRequestException, rs.Primary.ID) {
				return nil
			}
			return err
		}
		if resp.NamedQuery != nil {
			return fmt.Errorf("Athena Named Query (%s) found", rs.Primary.ID)
		}
	}
	return nil
}

func testAccCheckAWSAthenaNamedQueryExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).athenaconn

		input := &athena.GetNamedQueryInput{
			NamedQueryId: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetNamedQuery(input)
		return err
	}
}

func testAccAthenaNamedQueryConfig(rInt int, rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "hoge" {
	bucket = "tf-athena-db-%s-%d"
	force_destroy = true
}

resource "aws_athena_database" "hoge" {
	name = "%s"
	bucket = "${aws_s3_bucket.hoge.bucket}"
}

resource "aws_athena_named_query" "foo" {
  name = "tf-athena-named-query-%s"
  database = "${aws_athena_database.hoge.name}"
  query = "SELECT * FROM ${aws_athena_database.hoge.name} limit 10;"
  description = "tf test"
}
		`, rName, rInt, rName, rName)
}
