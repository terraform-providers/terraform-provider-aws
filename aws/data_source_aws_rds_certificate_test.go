package aws

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSRDSCertificateDataSource_Id(t *testing.T) {
	dataSourceName := "data.aws_rds_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRDSCertificatePreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSCertificateDataSourceConfigId(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", "data.aws_rds_certificate.latest", "id"),
				),
			},
		},
	})
}

func TestAccAWSRDSCertificateDataSource_LatestValidTill(t *testing.T) {
	dataSourceName := "data.aws_rds_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccAWSRDSCertificatePreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRDSCertificateDataSourceConfigLatestValidTill(),
				Check: resource.ComposeTestCheckFunc(
					testAccMatchResourceAttrRegionalARNNoAccount(dataSourceName, "arn", "rds", regexp.MustCompile(`cert:rds-ca-[0-9]{4}`)),
					resource.TestCheckResourceAttr(dataSourceName, "certificate_type", "CA"),
					resource.TestCheckResourceAttr(dataSourceName, "customer_override", "false"),
					resource.TestCheckNoResourceAttr(dataSourceName, "customer_override_valid_till"),
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile(`rds-ca-[0-9]{4}`)),
					resource.TestMatchResourceAttr(dataSourceName, "thumbprint", regexp.MustCompile(`[0-9a-f]+`)),
					resource.TestMatchResourceAttr(dataSourceName, "valid_from", regexp.MustCompile(rfc3339RegexPattern)),
					resource.TestMatchResourceAttr(dataSourceName, "valid_till", regexp.MustCompile(rfc3339RegexPattern)),
				),
			},
		},
	})
}

func testAccAWSRDSCertificatePreCheck(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	input := &rds.DescribeCertificatesInput{}

	_, err := conn.DescribeCertificates(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSRDSCertificateDataSourceConfigId() string {
	return `
data "aws_rds_certificate" "latest" {
  latest_valid_till = true
}

data "aws_rds_certificate" "test" {
  id = data.aws_rds_certificate.latest.id
}
`
}

func testAccAWSRDSCertificateDataSourceConfigLatestValidTill() string {
	return `
data "aws_rds_certificate" "test" {
  latest_valid_till = true
}
`
}
