package aws

import (
	"fmt"
	"regexp"
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestResourceSortByExpirationDate(t *testing.T) {
	certs := []*iam.ServerCertificateMetadata{
		{
			ServerCertificateName: aws.String("oldest"),
			Expiration:            aws.Time(time.Now()),
		},
		{
			ServerCertificateName: aws.String("latest"),
			Expiration:            aws.Time(time.Now().Add(3 * time.Hour)),
		},
		{
			ServerCertificateName: aws.String("in between"),
			Expiration:            aws.Time(time.Now().Add(2 * time.Hour)),
		},
	}
	sort.Sort(certificateByExpiration(certs))
	if *certs[0].ServerCertificateName != "latest" {
		t.Fatalf("Expected first item to be %q, but was %q", "latest", *certs[0].ServerCertificateName)
	}
}

func TestAccAWSDataSourceIAMServerCertificate_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")

	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMServerCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDataIAMServerCertConfig(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aws_iam_server_certificate.test_cert", "arn"),
					resource.TestCheckResourceAttrSet("data.aws_iam_server_certificate.test", "arn"),
					resource.TestCheckResourceAttrSet("data.aws_iam_server_certificate.test", "id"),
					resource.TestCheckResourceAttrSet("data.aws_iam_server_certificate.test", "name"),
					resource.TestCheckResourceAttrSet("data.aws_iam_server_certificate.test", "path"),
					resource.TestCheckResourceAttrSet("data.aws_iam_server_certificate.test", "upload_date"),
					resource.TestCheckResourceAttr("data.aws_iam_server_certificate.test", "certificate_chain", ""),
					resource.TestMatchResourceAttr("data.aws_iam_server_certificate.test", "certificate_body", regexp.MustCompile("^-----BEGIN CERTIFICATE-----")),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMServerCertificate_matchNamePrefix(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMServerCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsDataIAMServerCertConfigMatchNamePrefix,
				ExpectError: regexp.MustCompile(`Search for AWS IAM server certificate returned no results`),
			},
		},
	})
}

func TestAccAWSDataSourceIAMServerCertificate_path(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	path := "/test-path/"
	pathPrefix := "/test-path/"

	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, "example.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMServerCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDataIAMServerCertConfigPath(rName, path, pathPrefix, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_server_certificate.test", "path", path),
				),
			},
		},
	})
}

func testAccAwsDataIAMServerCertConfig(rName, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_iam_server_certificate" "test_cert" {
  name             = "%[1]s"
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

data "aws_iam_server_certificate" "test" {
  name   = aws_iam_server_certificate.test_cert.name
  latest = true
}
`, rName, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}

func testAccAwsDataIAMServerCertConfigPath(rName, path, pathPrefix, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_iam_server_certificate" "test_cert" {
  name             = "%[1]s"
  path             = "%[2]s"
  certificate_body = "%[3]s"
  private_key      = "%[4]s"
}

data "aws_iam_server_certificate" "test" {
  name        = aws_iam_server_certificate.test_cert.name
  path_prefix = "%[5]s"
  latest      = true
}
`, rName, path, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key), pathPrefix)
}

var testAccAwsDataIAMServerCertConfigMatchNamePrefix = `
data "aws_iam_server_certificate" "test" {
  name_prefix = "MyCert"
  latest      = true
}
`
