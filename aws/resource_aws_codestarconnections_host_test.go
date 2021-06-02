package aws

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCodeStarConnectionsHost_basic(t *testing.T) {
	var v codestarconnections.GetHostOutput
	resourceName := "aws_codestarconnections_host.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codestarconnections.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, codestarconnections.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeStarConnectionsHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeStarConnectionsHostConfigBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCodeStarConnectionsHostExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "id", "codestar-connections", regexp.MustCompile("host/.+")),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "codestar-connections", regexp.MustCompile("host/.+")),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "provider_endpoint", "https://test.com"),
					resource.TestCheckResourceAttr(resourceName, "provider_type", codestarconnections.ProviderTypeGitHubEnterpriseServer),
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

func TestAccAWSCodeStarConnectionsHost_disappears(t *testing.T) {
	var v codestarconnections.GetHostOutput
	resourceName := "aws_codestarconnections_host.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codestarconnections.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, codestarconnections.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeStarConnectionsHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeStarConnectionsHostConfigBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCodeStarConnectionsHostExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCodeStarConnectionsHost(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCodeStarConnectionsHost_vpcConfig(t *testing.T) {
	var v codestarconnections.GetHostOutput
	resourceName := "aws_codestarconnections_host.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codestarconnections.EndpointsID, t) },
		ErrorCheck:   testAccErrorCheck(t, codestarconnections.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeStarConnectionsHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeStarConnectionsHostConfigVpcConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSCodeStarConnectionsHostExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "id", "codestar-connections", regexp.MustCompile("host/.+")),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "codestar-connections", regexp.MustCompile("host/.+")),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "provider_endpoint", "https://test.com"),
					resource.TestCheckResourceAttr(resourceName, "provider_type", codestarconnections.ProviderTypeGitHubEnterpriseServer),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_configuration.0.tls_certificate", "-----BEGIN CERTIFICATE-----\nMIID2jCCAsKgAwIBAgIJAJ58TJVjU7G1MA0GCSqGSIb3DQEBBQUAMFExCzAJBgNV\nBAYTAlVTMREwDwYDVQQIEwhDb2xvcmFkbzEPMA0GA1UEBxMGRGVudmVyMRAwDgYD\nVQQKEwdDaGFydGVyMQwwCgYDVQQLEwNDU0UwHhcNMTcwMTMwMTkyMDA4WhcNMjYx\nMjA5MTkyMDA4WjBRMQswCQYDVQQGEwJVUzERMA8GA1UECBMIQ29sb3JhZG8xDzAN\nBgNVBAcTBkRlbnZlcjEQMA4GA1UEChMHQ2hhcnRlcjEMMAoGA1UECxMDQ1NFMIIB\nIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAv6dq6VLIImlAaTrckb5w3X6J\nWP7EGz2ChGAXlkEYto6dPCba0v5+f+8UlMOpeB25XGoai7gdItqNWVFpYsgmndx3\nvTad3ukO1zeElKtw5oHPH2plOaiv/gVJaDa9NTeINj0EtGZs74fCOclAzGFX5vBc\nb08ESWBceRgGjGv3nlij4JzHfqTkCKQz6P6pBivQBfk62rcOkkH5rKoaGltRHROS\nMbkwOhu2hN0KmSYTXRvts0LXnZU4N0l2ms39gmr7UNNNlKYINL2JoTs9dNBc7APD\ndZvlEHd+/FjcLCI8hC3t4g4AbfW0okIBCNG0+oVjqGb2DeONSJKsThahXt89MQID\nAQABo4G0MIGxMB0GA1UdDgQWBBQKq8JxjY1GmeZXJjfOMfW0kBIzPDCBgQYDVR0j\nBHoweIAUCqvCcY2NRpnmVyY3zjH1tJASMzyhVaRTMFExCzAJBgNVBAYTAlVTMREw\nDwYDVQQIEwhDb2xvcmFkbzEPMA0GA1UEBxMGRGVudmVyMRAwDgYDVQQKEwdDaGFy\ndGVyMQwwCgYDVQQLEwNDU0WCCQCefEyVY1OxtTAMBgNVHRMEBTADAQH/MA0GCSqG\nSIb3DQEBBQUAA4IBAQAWifoMk5kbv+yuWXvFwHiB4dWUUmMlUlPU/E300yVTRl58\np6DfOgJs7MMftd1KeWqTO+uW134QlTt7+jwI8Jq0uyKCu/O2kJhVtH/Ryog14tGl\n+wLcuIPLbwJI9CwZX4WMBrq4DnYss+6F47i8NCc+Z3MAiG4vtq9ytBmaod0dj2bI\ng4/Lac0e00dql9RnqENh1+dF0V+QgTJCoPkMqDNAlSB8vOodBW81UAb2z12t+IFi\n3X9J3WtCK2+T5brXL6itzewWJ2ALvX3QpmZx7fMHJ3tE+SjjyivE1BbOlzYHx83t\nTeYnm7pS9un7A/UzTDHbs7hPUezLek+H3xTPAnnq\n-----END CERTIFICATE-----\n"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_configuration.0.vpc_id"),
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

func testAccCheckAWSCodeStarConnectionsHostExists(n string, v *codestarconnections.GetHostOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No CodeStar host ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).codestarconnectionsconn

		resp, err := conn.GetHost(&codestarconnections.GetHostInput{
			HostArn: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccCheckAWSCodeStarConnectionsHostDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codestarconnectionsconn

	for _, rs := range s.RootModule().Resources {
		switch rs.Type {
		case "aws_codestarconnections_host":
			_, err := conn.DeleteHost(&codestarconnections.DeleteHostInput{
				HostArn: aws.String(rs.Primary.ID),
			})

			if err != nil && !isAWSErr(err, codestarconnections.ErrCodeResourceNotFoundException, "") {
				return err
			}
		}
	}

	return nil
}

func testAccAWSCodeStarConnectionsHostVpcConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count             = 2
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index + 2)
  vpc_id            = aws_vpc.test.id
  tags = {
    Name = "%[1]s-${count.index}"
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSCodeStarConnectionsHostConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codestarconnections_host" "test" {
  name              = %[1]q
  provider_endpoint = "https://test.com"
  provider_type     = "GitHubEnterpriseServer"
}
`, rName)
}

func testAccAWSCodeStarConnectionsHostConfigVpcConfig(rName string) string {
	return testAccAWSCodeStarConnectionsHostVpcConfig(rName) + fmt.Sprintf(`
resource "aws_codestarconnections_host" "test" {
  name              = %[1]q
  provider_endpoint = "https://test.com"
  provider_type     = "GitHubEnterpriseServer"
  vpc_configuration {
    security_group_ids = [aws_security_group.test.id]
    subnet_ids         = aws_subnet.test[*].id
    tls_certificate    = "-----BEGIN CERTIFICATE-----\nMIID2jCCAsKgAwIBAgIJAJ58TJVjU7G1MA0GCSqGSIb3DQEBBQUAMFExCzAJBgNV\nBAYTAlVTMREwDwYDVQQIEwhDb2xvcmFkbzEPMA0GA1UEBxMGRGVudmVyMRAwDgYD\nVQQKEwdDaGFydGVyMQwwCgYDVQQLEwNDU0UwHhcNMTcwMTMwMTkyMDA4WhcNMjYx\nMjA5MTkyMDA4WjBRMQswCQYDVQQGEwJVUzERMA8GA1UECBMIQ29sb3JhZG8xDzAN\nBgNVBAcTBkRlbnZlcjEQMA4GA1UEChMHQ2hhcnRlcjEMMAoGA1UECxMDQ1NFMIIB\nIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAv6dq6VLIImlAaTrckb5w3X6J\nWP7EGz2ChGAXlkEYto6dPCba0v5+f+8UlMOpeB25XGoai7gdItqNWVFpYsgmndx3\nvTad3ukO1zeElKtw5oHPH2plOaiv/gVJaDa9NTeINj0EtGZs74fCOclAzGFX5vBc\nb08ESWBceRgGjGv3nlij4JzHfqTkCKQz6P6pBivQBfk62rcOkkH5rKoaGltRHROS\nMbkwOhu2hN0KmSYTXRvts0LXnZU4N0l2ms39gmr7UNNNlKYINL2JoTs9dNBc7APD\ndZvlEHd+/FjcLCI8hC3t4g4AbfW0okIBCNG0+oVjqGb2DeONSJKsThahXt89MQID\nAQABo4G0MIGxMB0GA1UdDgQWBBQKq8JxjY1GmeZXJjfOMfW0kBIzPDCBgQYDVR0j\nBHoweIAUCqvCcY2NRpnmVyY3zjH1tJASMzyhVaRTMFExCzAJBgNVBAYTAlVTMREw\nDwYDVQQIEwhDb2xvcmFkbzEPMA0GA1UEBxMGRGVudmVyMRAwDgYDVQQKEwdDaGFy\ndGVyMQwwCgYDVQQLEwNDU0WCCQCefEyVY1OxtTAMBgNVHRMEBTADAQH/MA0GCSqG\nSIb3DQEBBQUAA4IBAQAWifoMk5kbv+yuWXvFwHiB4dWUUmMlUlPU/E300yVTRl58\np6DfOgJs7MMftd1KeWqTO+uW134QlTt7+jwI8Jq0uyKCu/O2kJhVtH/Ryog14tGl\n+wLcuIPLbwJI9CwZX4WMBrq4DnYss+6F47i8NCc+Z3MAiG4vtq9ytBmaod0dj2bI\ng4/Lac0e00dql9RnqENh1+dF0V+QgTJCoPkMqDNAlSB8vOodBW81UAb2z12t+IFi\n3X9J3WtCK2+T5brXL6itzewWJ2ALvX3QpmZx7fMHJ3tE+SjjyivE1BbOlzYHx83t\nTeYnm7pS9un7A/UzTDHbs7hPUezLek+H3xTPAnnq\n-----END CERTIFICATE-----\n"
    vpc_id             = aws_vpc.test.id
  }
}
`, rName)
}
