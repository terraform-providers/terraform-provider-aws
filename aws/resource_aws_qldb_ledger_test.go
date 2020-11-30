package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qldb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_qldb_ledger", &resource.Sweeper{
		Name: "aws_qldb_ledger",
		F:    testSweepQLDBLedgers,
	})
}

func testSweepQLDBLedgers(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).qldbconn
	input := &qldb.ListLedgersInput{}
	page, err := conn.ListLedgers(input)

	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping QLDB Ledger sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error listing QLDB Ledgers: %s", err)
	}

	for _, item := range page.Ledgers {
		input := &qldb.DeleteLedgerInput{
			Name: item.Name,
		}
		name := aws.StringValue(item.Name)

		log.Printf("[INFO] Deleting QLDB Ledger: %s", name)
		_, err = conn.DeleteLedger(input)

		if err != nil {
			log.Printf("[ERROR] Failed to delete QLDB Ledger %s: %s", name, err)
			continue
		}

		if err := waitForQLDBLedgerDeletion(conn, name); err != nil {
			log.Printf("[ERROR] Error waiting for QLDB Ledger (%s) deletion: %s", name, err)
		}
	}

	return nil
}

func TestAccAWSQLDBLedger_basic(t *testing.T) {
	var qldbCluster qldb.DescribeLedgerOutput
	rInt := acctest.RandInt()
	resourceName := "aws_qldb_ledger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(qldb.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSQLDBLedgerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQLDBLedgerConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSQLDBLedgerExists(resourceName, &qldbCluster),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "qldb", regexp.MustCompile(`ledger/.+`)),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("test-ledger-[0-9]+")),
					resource.TestCheckResourceAttr(resourceName, "deletion_protection", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func testAccCheckAWSQLDBLedgerDestroy(s *terraform.State) error {
	return testAccCheckAWSLedgerDestroyWithProvider(s, testAccProvider)
}

func testAccCheckAWSLedgerDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AWSClient).qldbconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_qldb_ledger" {
			continue
		}

		// Try to find the Group
		var err error
		resp, err := conn.DescribeLedger(
			&qldb.DescribeLedgerInput{
				Name: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if len(aws.StringValue(resp.Name)) != 0 && aws.StringValue(resp.Name) == rs.Primary.ID {
				return fmt.Errorf("QLDB Ledger %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the cluster is already destroyed
		if isAWSErr(err, qldb.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckAWSQLDBLedgerExists(n string, v *qldb.DescribeLedgerOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No QLDB Ledger ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).qldbconn
		resp, err := conn.DescribeLedger(&qldb.DescribeLedgerInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.Name == rs.Primary.ID {
			*v = *resp
			return nil
		}

		return fmt.Errorf("QLDB Ledger (%s) not found", rs.Primary.ID)
	}
}

func testAccAWSQLDBLedgerConfig(n int) string {
	return fmt.Sprintf(`
resource "aws_qldb_ledger" "test" {
  name                = "test-ledger-%d"
  deletion_protection = false
}
`, n)
}

func TestAccAWSQLDBLedger_Tags(t *testing.T) {
	var cluster1, cluster2, cluster3 qldb.DescribeLedgerOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_qldb_ledger.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(qldb.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSQLDBLedgerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSQLDBLedgerConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSQLDBLedgerExists(resourceName, &cluster1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSQLDBLedgerConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSQLDBLedgerExists(resourceName, &cluster2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSQLDBLedgerConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSQLDBLedgerExists(resourceName, &cluster3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccAWSQLDBLedgerConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_qldb_ledger" "test" {
  name                = %[1]q
  deletion_protection = false

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSQLDBLedgerConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_qldb_ledger" "test" {
  name                = %[1]q
  deletion_protection = false

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
