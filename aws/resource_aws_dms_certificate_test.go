package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSDmsCertificate_basic(t *testing.T) {
	resourceName := "aws_dms_certificate.dms_certificate"
	randId := acctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, dms.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDmsCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDmsCertificateConfig(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSDmsCertificateExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_arn"),
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

func TestAccAWSDmsCertificate_disappears(t *testing.T) {
	resourceName := "aws_dms_certificate.dms_certificate"
	randId := acctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, dms.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDmsCertificateConfig(randId),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSDmsCertificateExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsDmsCertificate(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDmsCertificate_CertificateWallet(t *testing.T) {
	resourceName := "aws_dms_certificate.dms_certificate"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, dms.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDmsCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDmsCertificateConfigCertificateWallet(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccAWSDmsCertificateExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_wallet"),
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

func TestAccAWSDmsCertificate_tags(t *testing.T) {
	resourceName := "aws_dms_certificate.dms_certificate"
	randId := acctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, dms.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDmsCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDmsCertificateConfigTags1(randId, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAWSDmsCertificateExists(resourceName),
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
				Config: testAccAWSDmsCertificateConfigTags2(randId, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAWSDmsCertificateExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSDmsCertificateConfigTags1(randId, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAWSDmsCertificateExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSDmsCertificateDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dms_certificate" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).dmsconn

		output, err := conn.DescribeCertificates(&dms.DescribeCertificatesInput{
			Filters: []*dms.Filter{
				{
					Name:   aws.String("certificate-id"),
					Values: []*string{aws.String(rs.Primary.ID)},
				},
			},
		})

		if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading DMS Certificate (%s): %w", rs.Primary.ID, err)
		}

		if output != nil && len(output.Certificates) != 0 {
			return fmt.Errorf("DMS Certificate (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccAWSDmsCertificateExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).dmsconn

		output, err := conn.DescribeCertificates(&dms.DescribeCertificatesInput{
			Filters: []*dms.Filter{
				{
					Name:   aws.String("certificate-id"),
					Values: []*string{aws.String(rs.Primary.ID)},
				},
			},
		})

		if err != nil {
			return fmt.Errorf("error reading DMS Certificate (%s): %w", rs.Primary.ID, err)
		}

		if output == nil || len(output.Certificates) == 0 {
			return fmt.Errorf("DMS Certificate (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAWSDmsCertificateConfig(randId string) string {
	return fmt.Sprintf(`
resource "aws_dms_certificate" "dms_certificate" {
  certificate_id  = "tf-test-dms-certificate-%[1]s"
  certificate_pem = "-----BEGIN CERTIFICATE-----\nMIID2jCCAsKgAwIBAgIJAJ58TJVjU7G1MA0GCSqGSIb3DQEBBQUAMFExCzAJBgNV\nBAYTAlVTMREwDwYDVQQIEwhDb2xvcmFkbzEPMA0GA1UEBxMGRGVudmVyMRAwDgYD\nVQQKEwdDaGFydGVyMQwwCgYDVQQLEwNDU0UwHhcNMTcwMTMwMTkyMDA4WhcNMjYx\nMjA5MTkyMDA4WjBRMQswCQYDVQQGEwJVUzERMA8GA1UECBMIQ29sb3JhZG8xDzAN\nBgNVBAcTBkRlbnZlcjEQMA4GA1UEChMHQ2hhcnRlcjEMMAoGA1UECxMDQ1NFMIIB\nIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAv6dq6VLIImlAaTrckb5w3X6J\nWP7EGz2ChGAXlkEYto6dPCba0v5+f+8UlMOpeB25XGoai7gdItqNWVFpYsgmndx3\nvTad3ukO1zeElKtw5oHPH2plOaiv/gVJaDa9NTeINj0EtGZs74fCOclAzGFX5vBc\nb08ESWBceRgGjGv3nlij4JzHfqTkCKQz6P6pBivQBfk62rcOkkH5rKoaGltRHROS\nMbkwOhu2hN0KmSYTXRvts0LXnZU4N0l2ms39gmr7UNNNlKYINL2JoTs9dNBc7APD\ndZvlEHd+/FjcLCI8hC3t4g4AbfW0okIBCNG0+oVjqGb2DeONSJKsThahXt89MQID\nAQABo4G0MIGxMB0GA1UdDgQWBBQKq8JxjY1GmeZXJjfOMfW0kBIzPDCBgQYDVR0j\nBHoweIAUCqvCcY2NRpnmVyY3zjH1tJASMzyhVaRTMFExCzAJBgNVBAYTAlVTMREw\nDwYDVQQIEwhDb2xvcmFkbzEPMA0GA1UEBxMGRGVudmVyMRAwDgYDVQQKEwdDaGFy\ndGVyMQwwCgYDVQQLEwNDU0WCCQCefEyVY1OxtTAMBgNVHRMEBTADAQH/MA0GCSqG\nSIb3DQEBBQUAA4IBAQAWifoMk5kbv+yuWXvFwHiB4dWUUmMlUlPU/E300yVTRl58\np6DfOgJs7MMftd1KeWqTO+uW134QlTt7+jwI8Jq0uyKCu/O2kJhVtH/Ryog14tGl\n+wLcuIPLbwJI9CwZX4WMBrq4DnYss+6F47i8NCc+Z3MAiG4vtq9ytBmaod0dj2bI\ng4/Lac0e00dql9RnqENh1+dF0V+QgTJCoPkMqDNAlSB8vOodBW81UAb2z12t+IFi\n3X9J3WtCK2+T5brXL6itzewWJ2ALvX3QpmZx7fMHJ3tE+SjjyivE1BbOlzYHx83t\nTeYnm7pS9un7A/UzTDHbs7hPUezLek+H3xTPAnnq\n-----END CERTIFICATE-----\n"
}
`, randId)
}

func testAccAWSDmsCertificateConfigCertificateWallet(rName string) string {
	return fmt.Sprintf(`
resource "aws_dms_certificate" "dms_certificate" {
  certificate_id     = %q
  certificate_wallet = filebase64("testdata/service/dms/oracle_wallet_certificate.pem")
}
`, rName)
}

func testAccAWSDmsCertificateConfigTags1(randId, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_dms_certificate" "dms_certificate" {
  certificate_id  = "tf-test-dms-certificate-%[1]s"
  certificate_pem = "-----BEGIN CERTIFICATE-----\nMIID2jCCAsKgAwIBAgIJAJ58TJVjU7G1MA0GCSqGSIb3DQEBBQUAMFExCzAJBgNV\nBAYTAlVTMREwDwYDVQQIEwhDb2xvcmFkbzEPMA0GA1UEBxMGRGVudmVyMRAwDgYD\nVQQKEwdDaGFydGVyMQwwCgYDVQQLEwNDU0UwHhcNMTcwMTMwMTkyMDA4WhcNMjYx\nMjA5MTkyMDA4WjBRMQswCQYDVQQGEwJVUzERMA8GA1UECBMIQ29sb3JhZG8xDzAN\nBgNVBAcTBkRlbnZlcjEQMA4GA1UEChMHQ2hhcnRlcjEMMAoGA1UECxMDQ1NFMIIB\nIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAv6dq6VLIImlAaTrckb5w3X6J\nWP7EGz2ChGAXlkEYto6dPCba0v5+f+8UlMOpeB25XGoai7gdItqNWVFpYsgmndx3\nvTad3ukO1zeElKtw5oHPH2plOaiv/gVJaDa9NTeINj0EtGZs74fCOclAzGFX5vBc\nb08ESWBceRgGjGv3nlij4JzHfqTkCKQz6P6pBivQBfk62rcOkkH5rKoaGltRHROS\nMbkwOhu2hN0KmSYTXRvts0LXnZU4N0l2ms39gmr7UNNNlKYINL2JoTs9dNBc7APD\ndZvlEHd+/FjcLCI8hC3t4g4AbfW0okIBCNG0+oVjqGb2DeONSJKsThahXt89MQID\nAQABo4G0MIGxMB0GA1UdDgQWBBQKq8JxjY1GmeZXJjfOMfW0kBIzPDCBgQYDVR0j\nBHoweIAUCqvCcY2NRpnmVyY3zjH1tJASMzyhVaRTMFExCzAJBgNVBAYTAlVTMREw\nDwYDVQQIEwhDb2xvcmFkbzEPMA0GA1UEBxMGRGVudmVyMRAwDgYDVQQKEwdDaGFy\ndGVyMQwwCgYDVQQLEwNDU0WCCQCefEyVY1OxtTAMBgNVHRMEBTADAQH/MA0GCSqG\nSIb3DQEBBQUAA4IBAQAWifoMk5kbv+yuWXvFwHiB4dWUUmMlUlPU/E300yVTRl58\np6DfOgJs7MMftd1KeWqTO+uW134QlTt7+jwI8Jq0uyKCu/O2kJhVtH/Ryog14tGl\n+wLcuIPLbwJI9CwZX4WMBrq4DnYss+6F47i8NCc+Z3MAiG4vtq9ytBmaod0dj2bI\ng4/Lac0e00dql9RnqENh1+dF0V+QgTJCoPkMqDNAlSB8vOodBW81UAb2z12t+IFi\n3X9J3WtCK2+T5brXL6itzewWJ2ALvX3QpmZx7fMHJ3tE+SjjyivE1BbOlzYHx83t\nTeYnm7pS9un7A/UzTDHbs7hPUezLek+H3xTPAnnq\n-----END CERTIFICATE-----\n"

  tags = {
    %[2]q = %[3]q
  }
}
`, randId, tagKey1, tagValue1)
}

func testAccAWSDmsCertificateConfigTags2(randId, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_dms_certificate" "dms_certificate" {
  certificate_id  = "tf-test-dms-certificate-%[1]s"
  certificate_pem = "-----BEGIN CERTIFICATE-----\nMIID2jCCAsKgAwIBAgIJAJ58TJVjU7G1MA0GCSqGSIb3DQEBBQUAMFExCzAJBgNV\nBAYTAlVTMREwDwYDVQQIEwhDb2xvcmFkbzEPMA0GA1UEBxMGRGVudmVyMRAwDgYD\nVQQKEwdDaGFydGVyMQwwCgYDVQQLEwNDU0UwHhcNMTcwMTMwMTkyMDA4WhcNMjYx\nMjA5MTkyMDA4WjBRMQswCQYDVQQGEwJVUzERMA8GA1UECBMIQ29sb3JhZG8xDzAN\nBgNVBAcTBkRlbnZlcjEQMA4GA1UEChMHQ2hhcnRlcjEMMAoGA1UECxMDQ1NFMIIB\nIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAv6dq6VLIImlAaTrckb5w3X6J\nWP7EGz2ChGAXlkEYto6dPCba0v5+f+8UlMOpeB25XGoai7gdItqNWVFpYsgmndx3\nvTad3ukO1zeElKtw5oHPH2plOaiv/gVJaDa9NTeINj0EtGZs74fCOclAzGFX5vBc\nb08ESWBceRgGjGv3nlij4JzHfqTkCKQz6P6pBivQBfk62rcOkkH5rKoaGltRHROS\nMbkwOhu2hN0KmSYTXRvts0LXnZU4N0l2ms39gmr7UNNNlKYINL2JoTs9dNBc7APD\ndZvlEHd+/FjcLCI8hC3t4g4AbfW0okIBCNG0+oVjqGb2DeONSJKsThahXt89MQID\nAQABo4G0MIGxMB0GA1UdDgQWBBQKq8JxjY1GmeZXJjfOMfW0kBIzPDCBgQYDVR0j\nBHoweIAUCqvCcY2NRpnmVyY3zjH1tJASMzyhVaRTMFExCzAJBgNVBAYTAlVTMREw\nDwYDVQQIEwhDb2xvcmFkbzEPMA0GA1UEBxMGRGVudmVyMRAwDgYDVQQKEwdDaGFy\ndGVyMQwwCgYDVQQLEwNDU0WCCQCefEyVY1OxtTAMBgNVHRMEBTADAQH/MA0GCSqG\nSIb3DQEBBQUAA4IBAQAWifoMk5kbv+yuWXvFwHiB4dWUUmMlUlPU/E300yVTRl58\np6DfOgJs7MMftd1KeWqTO+uW134QlTt7+jwI8Jq0uyKCu/O2kJhVtH/Ryog14tGl\n+wLcuIPLbwJI9CwZX4WMBrq4DnYss+6F47i8NCc+Z3MAiG4vtq9ytBmaod0dj2bI\ng4/Lac0e00dql9RnqENh1+dF0V+QgTJCoPkMqDNAlSB8vOodBW81UAb2z12t+IFi\n3X9J3WtCK2+T5brXL6itzewWJ2ALvX3QpmZx7fMHJ3tE+SjjyivE1BbOlzYHx83t\nTeYnm7pS9un7A/UzTDHbs7hPUezLek+H3xTPAnnq\n-----END CERTIFICATE-----\n"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, randId, tagKey1, tagValue1, tagKey2, tagValue2)
}
