package aws

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSOpsworksApplication_basic(t *testing.T) {
	var opsapp opsworks.App

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_opsworks_application.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(opsworks.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOpsworksApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOpsworksApplicationCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksApplicationExists(resourceName, &opsapp),
					testAccCheckAWSOpsworksCreateAppAttributes(&opsapp),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", "other"),
					resource.TestCheckResourceAttr(resourceName, "enable_ssl", "false"),
					resource.TestCheckNoResourceAttr(resourceName, "ssl_configuration"),
					resource.TestCheckNoResourceAttr(resourceName, "domains"),
					resource.TestCheckNoResourceAttr(resourceName, "app_source"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "environment.*", map[string]string{
						"key":    "key1",
						"value":  "value1",
						"secret": "",
					}),
					resource.TestCheckResourceAttr(resourceName, "document_root", "foo"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Environment variable import is not supported currently.
				ImportStateVerifyIgnore: []string{"environment"},
			},
			{
				Config: testAccAwsOpsworksApplicationUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksApplicationExists(resourceName, &opsapp),
					testAccCheckAWSOpsworksUpdateAppAttributes(&opsapp),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "type", "rails"),
					resource.TestCheckResourceAttr(resourceName, "enable_ssl", "true"),
					resource.TestCheckResourceAttr(resourceName, "ssl_configuration.0.certificate", "-----BEGIN CERTIFICATE-----\nMIIBkDCB+gIJALoScFD0sJq3MA0GCSqGSIb3DQEBBQUAMA0xCzAJBgNVBAYTAkRF\nMB4XDTE1MTIxOTIwMzU1MVoXDTE2MDExODIwMzU1MVowDTELMAkGA1UEBhMCREUw\ngZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAKKQKbTTH/Julz16xY7ArYlzJYCP\nedTCx1bopuryCx/+d1gC94MtRdlPSpQl8mfc9iBdtXbJppp73Qh/DzLzO9Ns25xZ\n+kUQMhbIyLsaCBzuEGLgAaVdGpNvRBw++UoYtd0U7QczFAreTGLH8n8+FIzuI5Mc\n+MJ1TKbbt5gFfRSzAgMBAAEwDQYJKoZIhvcNAQEFBQADgYEALARo96wCDmaHKCaX\nS0IGLGnZCfiIUfCmBxOXBSJxDBwter95QHR0dMGxYIujee5n4vvavpVsqZnfMC3I\nOZWPlwiUJbNIpK+04Bg2vd5m/NMMrvi75RfmyeMtSfq/NrIX2Q3+nyWI7DLq7yZI\nV/YEvOqdAiy5NEWBztHx8HvB9G4=\n-----END CERTIFICATE-----"),
					resource.TestCheckResourceAttr(resourceName, "ssl_configuration.0.private_key", "-----BEGIN RSA PRIVATE KEY-----\nMIICXQIBAAKBgQCikCm00x/ybpc9esWOwK2JcyWAj3nUwsdW6Kbq8gsf/ndYAveD\nLUXZT0qUJfJn3PYgXbV2yaaae90Ifw8y8zvTbNucWfpFEDIWyMi7Gggc7hBi4AGl\nXRqTb0QcPvlKGLXdFO0HMxQK3kxix/J/PhSM7iOTHPjCdUym27eYBX0UswIDAQAB\nAoGBAIYcrvuqDboguI8U4TUjCkfSAgds1pLLWk79wu8jXkA329d1IyNKT0y3WIye\nPbyoEzmidZmZROQ/+ZsPz8c12Y0DrX73WSVzKNyJeP7XMk9HSzA1D9RX0U0S+5Kh\nFAMc2NEVVFIfQtVtoVmHdKDpnRYtOCHLW9rRpvqOOjd4mYk5AkEAzeiFr1mtlnsa\n67shMxzDaOTAFMchRz6G7aSovvCztxcB63ulFI/w9OTUMdTQ7ff7pet+lVihLc2W\nefIL0HvsjQJBAMocNTKaR/TnsV5GSk2kPAdR+zFP5sQy8sfMy0lEXTylc7zN4ajX\nMeHVoxp+GZgpfDcZ3ya808H1umyXh+xA1j8CQE9x9ZKQYT98RAjL7KVR5btk9w+N\nPTPF1j1+mHUDXfO4ds8qp6jlWKzEVXLcj7ghRADiebaZuaZ4eiSW1SQdjEkCQQC4\nwDhQ3X9RfEpCp3ZcqvjEqEg6t5N3XitYQPjDLN8eBRBbUsgpEy3iBuxl10eGNMX7\niIbYXlwkPYAArDPv3wT5AkAwp4vym+YKmDqh6gseKfRDuJqRiW9yD5A8VGr/w88k\n5rkuduVGP7tK3uIp00Its3aEyKF8mLGWYszVGeeLxAMH\n-----END RSA PRIVATE KEY-----"),
					resource.TestCheckResourceAttr(resourceName, "domains.0", "example.com"),
					resource.TestCheckResourceAttr(resourceName, "domains.1", "sub.example.com"),
					resource.TestCheckResourceAttr(resourceName, "app_source.0.password", ""),
					resource.TestCheckResourceAttr(resourceName, "app_source.0.revision", "master"),
					resource.TestCheckResourceAttr(resourceName, "app_source.0.ssh_key", ""),
					resource.TestCheckResourceAttr(resourceName, "app_source.0.type", "git"),
					resource.TestCheckResourceAttr(resourceName, "app_source.0.url", "https://github.com/aws/example.git"),
					resource.TestCheckResourceAttr(resourceName, "app_source.0.username", ""),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "environment.*", map[string]string{
						"key":    "key2",
						"value":  "value2",
						"secure": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "environment.*", map[string]string{
						"key":    "key1",
						"value":  "value1",
						"secret": "",
					}),
					resource.TestCheckResourceAttr(resourceName, "document_root", "root"),
					resource.TestCheckResourceAttr(resourceName, "auto_bundle_on_deploy", "true"),
					resource.TestCheckResourceAttr(resourceName, "rails_env", "staging"),
				),
			},
		},
	})
}

func testAccCheckAWSOpsworksApplicationExists(
	n string, opsapp *opsworks.App) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).opsworksconn

		params := &opsworks.DescribeAppsInput{
			AppIds: []*string{&rs.Primary.ID},
		}
		resp, err := conn.DescribeApps(params)

		if err != nil {
			return err
		}

		if v := len(resp.Apps); v != 1 {
			return fmt.Errorf("Expected 1 response returned, got %d", v)
		}

		*opsapp = *resp.Apps[0]

		return nil
	}
}

func testAccCheckAWSOpsworksCreateAppAttributes(
	opsapp *opsworks.App) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *opsapp.EnableSsl {
			return fmt.Errorf("Unexpected enable ssl: %t", *opsapp.EnableSsl)
		}

		if *opsapp.Attributes["DocumentRoot"] != "foo" {
			return fmt.Errorf("Unnexpected document root: %s", *opsapp.Attributes["DocumentRoot"])
		}

		if *opsapp.Type != opsworks.AppTypeOther {
			return fmt.Errorf("Unnexpected type: %s", *opsapp.Type)
		}

		if *opsapp.AppSource.Type != "other" {
			return fmt.Errorf("Unnexpected appsource type: %s", *opsapp.AppSource.Type)
		}

		expectedEnv := []*opsworks.EnvironmentVariable{
			{
				Key:    aws.String("key1"),
				Value:  aws.String("value1"),
				Secure: aws.Bool(false),
			},
		}

		if !reflect.DeepEqual(expectedEnv, opsapp.Environment) {
			return fmt.Errorf("Unnexpected environment: %s", opsapp.Environment)
		}

		if v := len(opsapp.Domains); v != 0 {
			return fmt.Errorf("Expected 0 domains returned, got %d", v)
		}

		return nil
	}
}

func testAccCheckAWSOpsworksUpdateAppAttributes(
	opsapp *opsworks.App) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *opsapp.Type != "rails" {
			return fmt.Errorf("Unnexpected type: %s", *opsapp.Type)
		}

		if !*opsapp.EnableSsl {
			return fmt.Errorf("Unexpected enable ssl: %t", *opsapp.EnableSsl)
		}

		if *opsapp.SslConfiguration.Certificate != "-----BEGIN CERTIFICATE-----\nMIIBkDCB+gIJALoScFD0sJq3MA0GCSqGSIb3DQEBBQUAMA0xCzAJBgNVBAYTAkRF\nMB4XDTE1MTIxOTIwMzU1MVoXDTE2MDExODIwMzU1MVowDTELMAkGA1UEBhMCREUw\ngZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAKKQKbTTH/Julz16xY7ArYlzJYCP\nedTCx1bopuryCx/+d1gC94MtRdlPSpQl8mfc9iBdtXbJppp73Qh/DzLzO9Ns25xZ\n+kUQMhbIyLsaCBzuEGLgAaVdGpNvRBw++UoYtd0U7QczFAreTGLH8n8+FIzuI5Mc\n+MJ1TKbbt5gFfRSzAgMBAAEwDQYJKoZIhvcNAQEFBQADgYEALARo96wCDmaHKCaX\nS0IGLGnZCfiIUfCmBxOXBSJxDBwter95QHR0dMGxYIujee5n4vvavpVsqZnfMC3I\nOZWPlwiUJbNIpK+04Bg2vd5m/NMMrvi75RfmyeMtSfq/NrIX2Q3+nyWI7DLq7yZI\nV/YEvOqdAiy5NEWBztHx8HvB9G4=\n-----END CERTIFICATE-----" {
			return fmt.Errorf("Unexpected ssl configuration certificate: %s", *opsapp.SslConfiguration.Certificate)
		}

		if *opsapp.SslConfiguration.PrivateKey != "-----BEGIN RSA PRIVATE KEY-----\nMIICXQIBAAKBgQCikCm00x/ybpc9esWOwK2JcyWAj3nUwsdW6Kbq8gsf/ndYAveD\nLUXZT0qUJfJn3PYgXbV2yaaae90Ifw8y8zvTbNucWfpFEDIWyMi7Gggc7hBi4AGl\nXRqTb0QcPvlKGLXdFO0HMxQK3kxix/J/PhSM7iOTHPjCdUym27eYBX0UswIDAQAB\nAoGBAIYcrvuqDboguI8U4TUjCkfSAgds1pLLWk79wu8jXkA329d1IyNKT0y3WIye\nPbyoEzmidZmZROQ/+ZsPz8c12Y0DrX73WSVzKNyJeP7XMk9HSzA1D9RX0U0S+5Kh\nFAMc2NEVVFIfQtVtoVmHdKDpnRYtOCHLW9rRpvqOOjd4mYk5AkEAzeiFr1mtlnsa\n67shMxzDaOTAFMchRz6G7aSovvCztxcB63ulFI/w9OTUMdTQ7ff7pet+lVihLc2W\nefIL0HvsjQJBAMocNTKaR/TnsV5GSk2kPAdR+zFP5sQy8sfMy0lEXTylc7zN4ajX\nMeHVoxp+GZgpfDcZ3ya808H1umyXh+xA1j8CQE9x9ZKQYT98RAjL7KVR5btk9w+N\nPTPF1j1+mHUDXfO4ds8qp6jlWKzEVXLcj7ghRADiebaZuaZ4eiSW1SQdjEkCQQC4\nwDhQ3X9RfEpCp3ZcqvjEqEg6t5N3XitYQPjDLN8eBRBbUsgpEy3iBuxl10eGNMX7\niIbYXlwkPYAArDPv3wT5AkAwp4vym+YKmDqh6gseKfRDuJqRiW9yD5A8VGr/w88k\n5rkuduVGP7tK3uIp00Its3aEyKF8mLGWYszVGeeLxAMH\n-----END RSA PRIVATE KEY-----" {
			return fmt.Errorf("Unexpected ssl configuration private key: %s", *opsapp.SslConfiguration.PrivateKey)
		}

		expectedAttrs := map[string]*string{
			"DocumentRoot":        aws.String("root"),
			"RailsEnv":            aws.String("staging"),
			"AutoBundleOnDeploy":  aws.String("true"),
			"AwsFlowRubySettings": nil,
		}

		if !reflect.DeepEqual(expectedAttrs, opsapp.Attributes) {
			return fmt.Errorf("Unnexpected Attributes: %v", aws.StringValueMap(opsapp.Attributes))
		}

		expectedAppSource := &opsworks.Source{
			Type:     aws.String("git"),
			Revision: aws.String("master"),
			Url:      aws.String("https://github.com/aws/example.git"),
		}

		if !reflect.DeepEqual(expectedAppSource, opsapp.AppSource) {
			return fmt.Errorf("Unnexpected appsource: %s", opsapp.AppSource)
		}

		expectedEnv := []*opsworks.EnvironmentVariable{
			{
				Key:    aws.String("key2"),
				Value:  aws.String("*****FILTERED*****"),
				Secure: aws.Bool(true),
			},
			{
				Key:    aws.String("key1"),
				Value:  aws.String("value1"),
				Secure: aws.Bool(false),
			},
		}

		if !reflect.DeepEqual(expectedEnv, opsapp.Environment) {
			return fmt.Errorf("Unnexpected environment: %s", opsapp.Environment)
		}

		expectedDomains := []*string{
			aws.String("example.com"),
			aws.String("sub.example.com"),
		}

		if !reflect.DeepEqual(expectedDomains, opsapp.Domains) {
			return fmt.Errorf("Unnexpected Daomins : %v", aws.StringValueSlice(opsapp.Domains))
		}

		return nil
	}
}

func testAccCheckAwsOpsworksApplicationDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*AWSClient).opsworksconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_opsworks_application" {
			continue
		}

		req := &opsworks.DescribeAppsInput{
			AppIds: []*string{
				aws.String(rs.Primary.ID),
			},
		}

		resp, err := client.DescribeApps(req)
		if err == nil {
			if len(resp.Apps) > 0 {
				return fmt.Errorf("OpsWorks App still exist.")
			}
		}

		if !isAWSErr(err, opsworks.ErrCodeResourceNotFoundException, "") {
			return err
		}
	}

	return nil
}

func testAccAwsOpsworksApplicationCreate(name string) string {
	return testAccAwsOpsworksStackConfigVpcCreate(name) +
		fmt.Sprintf(`
resource "aws_opsworks_application" "test" {
  document_root = "foo"
  enable_ssl    = false
  name          = %q
  stack_id      = aws_opsworks_stack.tf-acc.id
  type          = "other"

  app_source {
    type = "other"
  }

  environment {
    key    = "key1"
    value  = "value1"
    secure = false
  }
}
`, name)
}

func testAccAwsOpsworksApplicationUpdate(name string) string {
	return testAccAwsOpsworksStackConfigVpcCreate(name) +
		fmt.Sprintf(`
resource "aws_opsworks_application" "test" {
  auto_bundle_on_deploy = "true"
  document_root         = "root"
  domains               = ["example.com", "sub.example.com"]
  enable_ssl            = true
  name                  = %q
  rails_env             = "staging"
  stack_id              = aws_opsworks_stack.tf-acc.id
  type                  = "rails"

  ssl_configuration {
    private_key = <<EOS
-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQCikCm00x/ybpc9esWOwK2JcyWAj3nUwsdW6Kbq8gsf/ndYAveD
LUXZT0qUJfJn3PYgXbV2yaaae90Ifw8y8zvTbNucWfpFEDIWyMi7Gggc7hBi4AGl
XRqTb0QcPvlKGLXdFO0HMxQK3kxix/J/PhSM7iOTHPjCdUym27eYBX0UswIDAQAB
AoGBAIYcrvuqDboguI8U4TUjCkfSAgds1pLLWk79wu8jXkA329d1IyNKT0y3WIye
PbyoEzmidZmZROQ/+ZsPz8c12Y0DrX73WSVzKNyJeP7XMk9HSzA1D9RX0U0S+5Kh
FAMc2NEVVFIfQtVtoVmHdKDpnRYtOCHLW9rRpvqOOjd4mYk5AkEAzeiFr1mtlnsa
67shMxzDaOTAFMchRz6G7aSovvCztxcB63ulFI/w9OTUMdTQ7ff7pet+lVihLc2W
efIL0HvsjQJBAMocNTKaR/TnsV5GSk2kPAdR+zFP5sQy8sfMy0lEXTylc7zN4ajX
MeHVoxp+GZgpfDcZ3ya808H1umyXh+xA1j8CQE9x9ZKQYT98RAjL7KVR5btk9w+N
PTPF1j1+mHUDXfO4ds8qp6jlWKzEVXLcj7ghRADiebaZuaZ4eiSW1SQdjEkCQQC4
wDhQ3X9RfEpCp3ZcqvjEqEg6t5N3XitYQPjDLN8eBRBbUsgpEy3iBuxl10eGNMX7
iIbYXlwkPYAArDPv3wT5AkAwp4vym+YKmDqh6gseKfRDuJqRiW9yD5A8VGr/w88k
5rkuduVGP7tK3uIp00Its3aEyKF8mLGWYszVGeeLxAMH
-----END RSA PRIVATE KEY-----
EOS

    certificate = <<EOS
-----BEGIN CERTIFICATE-----
MIIBkDCB+gIJALoScFD0sJq3MA0GCSqGSIb3DQEBBQUAMA0xCzAJBgNVBAYTAkRF
MB4XDTE1MTIxOTIwMzU1MVoXDTE2MDExODIwMzU1MVowDTELMAkGA1UEBhMCREUw
gZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAKKQKbTTH/Julz16xY7ArYlzJYCP
edTCx1bopuryCx/+d1gC94MtRdlPSpQl8mfc9iBdtXbJppp73Qh/DzLzO9Ns25xZ
+kUQMhbIyLsaCBzuEGLgAaVdGpNvRBw++UoYtd0U7QczFAreTGLH8n8+FIzuI5Mc
+MJ1TKbbt5gFfRSzAgMBAAEwDQYJKoZIhvcNAQEFBQADgYEALARo96wCDmaHKCaX
S0IGLGnZCfiIUfCmBxOXBSJxDBwter95QHR0dMGxYIujee5n4vvavpVsqZnfMC3I
OZWPlwiUJbNIpK+04Bg2vd5m/NMMrvi75RfmyeMtSfq/NrIX2Q3+nyWI7DLq7yZI
V/YEvOqdAiy5NEWBztHx8HvB9G4=
-----END CERTIFICATE-----
EOS

  }

  app_source {
    type     = "git"
    revision = "master"
    url      = "https://github.com/aws/example.git"
  }

  environment {
    key    = "key1"
    value  = "value1"
    secure = false
  }

  environment {
    key   = "key2"
    value = "value2"
  }
}
`, name)
}
