package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_directory_service_directory", &resource.Sweeper{
		Name:         "aws_directory_service_directory",
		F:            testSweepDirectoryServiceDirectories,
		Dependencies: []string{"aws_fsx_windows_file_system", "aws_workspaces_directory"},
	})
}

func testSweepDirectoryServiceDirectories(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).dsconn

	input := &directoryservice.DescribeDirectoriesInput{}
	for {
		resp, err := conn.DescribeDirectories(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Directory Service Directory sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Directory Service Directories: %s", err)
		}

		for _, directory := range resp.DirectoryDescriptions {
			id := aws.StringValue(directory.DirectoryId)

			deleteDirectoryInput := directoryservice.DeleteDirectoryInput{
				DirectoryId: directory.DirectoryId,
			}

			log.Printf("[INFO] Deleting Directory Service Directory: %s", deleteDirectoryInput)
			_, err := conn.DeleteDirectory(&deleteDirectoryInput)
			if err != nil {
				return fmt.Errorf("error deleting Directory Service Directory (%s): %s", id, err)
			}

			log.Printf("[INFO] Waiting for Directory Service Directory (%q) to be deleted", id)
			err = waitForDirectoryServiceDirectoryDeletion(conn, id)
			if err != nil {
				return fmt.Errorf("error waiting for Directory Service (%s) to be deleted: %s", id, err)
			}
		}

		if resp.NextToken == nil {
			break
		}

		input.NextToken = resp.NextToken
	}

	return nil
}

func TestAccAWSDirectoryServiceDirectory_basic(t *testing.T) {
	resourceName := "aws_directory_service_directory.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSDirectoryService(t)
			testAccPreCheckAWSDirectoryServiceSimpleDirectory(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists("aws_directory_service_directory.test"),
					resource.TestCheckResourceAttrSet("aws_directory_service_directory.test", "security_group_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccAWSDirectoryServiceDirectory_tags(t *testing.T) {
	resourceName := "aws_directory_service_directory.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSDirectoryService(t)
			testAccPreCheckAWSDirectoryServiceSimpleDirectory(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryTagsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists("aws_directory_service_directory.test"),
					resource.TestCheckResourceAttr("aws_directory_service_directory.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_directory_service_directory.test", "tags.foo", "test"),
					resource.TestCheckResourceAttr("aws_directory_service_directory.test", "tags.project", "test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
			{
				Config: testAccDirectoryServiceDirectoryUpdateTagsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists("aws_directory_service_directory.test"),
					resource.TestCheckResourceAttr("aws_directory_service_directory.test", "tags.%", "3"),
					resource.TestCheckResourceAttr("aws_directory_service_directory.test", "tags.foo", "test"),
					resource.TestCheckResourceAttr("aws_directory_service_directory.test", "tags.project", "test2"),
					resource.TestCheckResourceAttr("aws_directory_service_directory.test", "tags.fizz", "buzz"),
				),
			},
			{
				Config: testAccDirectoryServiceDirectoryRemoveTagsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists("aws_directory_service_directory.test"),
					resource.TestCheckResourceAttr("aws_directory_service_directory.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_directory_service_directory.test", "tags.foo", "test"),
				),
			},
		},
	})
}

func TestAccAWSDirectoryServiceDirectory_microsoft(t *testing.T) {
	resourceName := "aws_directory_service_directory.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDirectoryService(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig_microsoft,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists("aws_directory_service_directory.test"),
					resource.TestCheckResourceAttr("aws_directory_service_directory.test", "edition", directoryservice.DirectoryEditionEnterprise),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccAWSDirectoryServiceDirectory_microsoftStandard(t *testing.T) {
	resourceName := "aws_directory_service_directory.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDirectoryService(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig_microsoftStandard,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists("aws_directory_service_directory.test"),
					resource.TestCheckResourceAttr("aws_directory_service_directory.test", "edition", directoryservice.DirectoryEditionStandard),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccAWSDirectoryServiceDirectory_connector(t *testing.T) {
	resourceName := "aws_directory_service_directory.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSDirectoryService(t)
			testAccPreCheckAWSDirectoryServiceSimpleDirectory(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig_connector,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists("aws_directory_service_directory.connector"),
					resource.TestCheckResourceAttrSet("aws_directory_service_directory.connector", "security_group_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
		},
	})
}

func TestAccAWSDirectoryServiceDirectory_withAliasAndSso(t *testing.T) {
	alias := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_directory_service_directory.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSDirectoryService(t)
			testAccPreCheckAWSDirectoryServiceSimpleDirectory(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig_withAlias(alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists("aws_directory_service_directory.test2"),
					testAccCheckServiceDirectoryAlias("aws_directory_service_directory.test2", alias),
					testAccCheckServiceDirectorySso("aws_directory_service_directory.test2", false),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"password",
				},
			},
			{
				Config: testAccDirectoryServiceDirectoryConfig_withSso(alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists("aws_directory_service_directory.test2"),
					testAccCheckServiceDirectoryAlias("aws_directory_service_directory.test2", alias),
					testAccCheckServiceDirectorySso("aws_directory_service_directory.test2", true),
				),
			},
			{
				Config: testAccDirectoryServiceDirectoryConfig_withSso_modified(alias),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists("aws_directory_service_directory.test2"),
					testAccCheckServiceDirectoryAlias("aws_directory_service_directory.test2", alias),
					testAccCheckServiceDirectorySso("aws_directory_service_directory.test2", false),
				),
			},
		},
	})
}

func testAccCheckDirectoryServiceDirectoryDestroy(s *terraform.State) error {
	dsconn := testAccProvider.Meta().(*AWSClient).dsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_directory_service_directory" {
			continue
		}

		input := directoryservice.DescribeDirectoriesInput{
			DirectoryIds: []*string{aws.String(rs.Primary.ID)},
		}
		out, err := dsconn.DescribeDirectories(&input)

		if isAWSErr(err, "EntityDoesNotExistException", "") {
			continue
		}

		if err != nil {
			return err
		}

		if out != nil && len(out.DirectoryDescriptions) > 0 {
			return fmt.Errorf("Expected AWS Directory Service Directory to be gone, but was still found")
		}
	}

	return nil
}

func testAccCheckServiceDirectoryExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		dsconn := testAccProvider.Meta().(*AWSClient).dsconn
		out, err := dsconn.DescribeDirectories(&directoryservice.DescribeDirectoriesInput{
			DirectoryIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if len(out.DirectoryDescriptions) < 1 {
			return fmt.Errorf("No DS directory found")
		}

		if *out.DirectoryDescriptions[0].DirectoryId != rs.Primary.ID {
			return fmt.Errorf("DS directory ID mismatch - existing: %q, state: %q",
				*out.DirectoryDescriptions[0].DirectoryId, rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckServiceDirectoryAlias(name, alias string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		dsconn := testAccProvider.Meta().(*AWSClient).dsconn
		out, err := dsconn.DescribeDirectories(&directoryservice.DescribeDirectoriesInput{
			DirectoryIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if *out.DirectoryDescriptions[0].Alias != alias {
			return fmt.Errorf("DS directory Alias mismatch - actual: %q, expected: %q",
				*out.DirectoryDescriptions[0].Alias, alias)
		}

		return nil
	}
}

func testAccCheckServiceDirectorySso(name string, ssoEnabled bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		dsconn := testAccProvider.Meta().(*AWSClient).dsconn
		out, err := dsconn.DescribeDirectories(&directoryservice.DescribeDirectoriesInput{
			DirectoryIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return err
		}

		if *out.DirectoryDescriptions[0].SsoEnabled != ssoEnabled {
			return fmt.Errorf("DS directory SSO mismatch - actual: %t, expected: %t",
				*out.DirectoryDescriptions[0].SsoEnabled, ssoEnabled)
		}

		return nil
	}
}

func testAccPreCheckAWSDirectoryService(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).dsconn

	input := &directoryservice.DescribeDirectoriesInput{}

	_, err := conn.DescribeDirectories(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

// Certain regions such as AWS GovCloud (US) do not support Simple AD directories
// and we do not have a good read-only way to determine this situation. Here we
// opt to perform a creation that will fail so we can determine Simple AD support.
func testAccPreCheckAWSDirectoryServiceSimpleDirectory(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).dsconn

	input := &directoryservice.CreateDirectoryInput{
		Name:     aws.String("corp.example.com"),
		Password: aws.String("PreCheck123"),
		Size:     aws.String(directoryservice.DirectorySizeSmall),
	}

	_, err := conn.CreateDirectory(input)

	if isAWSErr(err, directoryservice.ErrCodeClientException, "Simple AD directory creation is currently not supported in this region") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil && !isAWSErr(err, directoryservice.ErrCodeInvalidParameterException, "VpcSettings must be specified") {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

const testAccDirectoryServiceDirectoryConfig = `
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_directory_service_directory" "test" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size = "Small"

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.test.id}"]
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-directory-service-directory"
	}
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-foo"
  }
}
resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
  cidr_block = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-test"
  }
}
`

const testAccDirectoryServiceDirectoryTagsConfig = `
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_directory_service_directory" "test" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size = "Small"

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.test.id}"]
  }

	tags = {
		foo = "test"
		project = "test"
	}
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-directory-service-directory-tags"
	}
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-tags-foo"
  }
}
resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
  cidr_block = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-tags-test"
  }
}
`

const testAccDirectoryServiceDirectoryUpdateTagsConfig = `
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_directory_service_directory" "test" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size = "Small"

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.test.id}"]
  }

	tags = {
		foo = "test"
		project = "test2"
		fizz = "buzz"
	}
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-directory-service-directory-tags"
	}
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-tags-foo"
  }
}
resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
  cidr_block = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-tags-test"
  }
}
`

const testAccDirectoryServiceDirectoryRemoveTagsConfig = `
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_directory_service_directory" "test" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size = "Small"

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.test.id}"]
  }

	tags = {
		foo = "test"
	}
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-directory-service-directory-tags"
	}
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-tags-foo"
  }
}
resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
  cidr_block = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-tags-test"
  }
}
`

const testAccDirectoryServiceDirectoryConfig_connector = `
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_directory_service_directory" "test" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size = "Small"

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.test.id}"]
  }
}

resource "aws_directory_service_directory" "connector" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size = "Small"
  type = "ADConnector"

  connect_settings {
    customer_dns_ips = aws_directory_service_directory.test.dns_ip_addresses
    customer_username = "Administrator"
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.test.id}"]
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-directory-service-directory-connector"
	}
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-connector-foo"
  }
}
resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
  cidr_block = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-connector-test"
  }
}
`

const testAccDirectoryServiceDirectoryConfig_microsoft = `
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_directory_service_directory" "test" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type = "MicrosoftAD"

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.test.id}"]
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-directory-service-directory-microsoft"
	}
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-microsoft-foo"
  }
}
resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
  cidr_block = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-microsoft-test"
  }
}
`

const testAccDirectoryServiceDirectoryConfig_microsoftStandard = `
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_directory_service_directory" "test" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type = "MicrosoftAD"
  edition = "Standard"

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.test.id}"]
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-directory-service-directory-microsoft"
	}
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-microsoft-foo"
  }
}
resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
  cidr_block = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-microsoft-test"
  }
}
`

func testAccDirectoryServiceDirectoryConfig_withAlias(alias string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_directory_service_directory" "test2" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size = "Small"
  alias = %[1]q

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.test.id}"]
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-directory-service-directory-with-alias"
	}
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-with-alias-foo"
  }
}
resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
  cidr_block = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-with-alias-test"
  }
}
`, alias)
}

func testAccDirectoryServiceDirectoryConfig_withSso(alias string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_directory_service_directory" "test2" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size = "Small"
  alias = %[1]q
  enable_sso = true

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.test.id}"]
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-directory-service-directory-with-sso"
	}
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-with-sso-foo"
  }
}
resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
  cidr_block = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-with-sso-test"
  }
}
`, alias)
}

func testAccDirectoryServiceDirectoryConfig_withSso_modified(alias string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_directory_service_directory" "test2" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size = "Small"
  alias = %[1]q
  enable_sso = false

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.test.id}"]
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
	tags = {
		Name = "terraform-testacc-directory-service-directory-with-sso-modified"
	}
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-with-sso-foo"
  }
}
resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
  cidr_block = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-with-sso-test"
  }
}
`, alias)
}
