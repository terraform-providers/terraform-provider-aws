package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/qldb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsQLDBLedger_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(7)) // QLDB name cannot be longer than 32 characters

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(qldb.EndpointsID, t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsQLDBLedgerConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.aws_qldb_ledger.by_name", "arn", "aws_qldb_ledger.tf_test", "arn"),
					resource.TestCheckResourceAttrPair("data.aws_qldb_ledger.by_name", "deletion_protection", "aws_qldb_ledger.tf_test", "deletion_protection"),
					resource.TestCheckResourceAttrPair("data.aws_qldb_ledger.by_name", "name", "aws_qldb_ledger.tf_test", "name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsQLDBLedgerConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_qldb_ledger" "tf_wrong1" {
  name                = "%[1]s1"
  deletion_protection = false
}

resource "aws_qldb_ledger" "tf_test" {
  name                = "%[1]s2"
  deletion_protection = false
}

resource "aws_qldb_ledger" "tf_wrong2" {
  name                = "%[1]s3"
  deletion_protection = false
}

data "aws_qldb_ledger" "by_name" {
  name = aws_qldb_ledger.tf_test.name
}
`, rName)
}
