package aws

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/directoryservice"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_directory_service_directory", &resource.Sweeper{
		Name: "aws_directory_service_directory",
		F:    testSweepDirectoryServiceDirectories,
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
			name := aws.StringValue(directory.Name)

			if name != "corp.notexample.com" && name != "terraformtesting.com" {
				log.Printf("[INFO] Skipping Directory Service Directory: %s / %s", id, name)
				continue
			}

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

func TestDiffTagsDirectoryService(t *testing.T) {
	cases := []struct {
		Old, New       map[string]interface{}
		Create, Remove map[string]string
	}{
		// Basic add/remove
		{
			Old: map[string]interface{}{
				"foo": "bar",
			},
			New: map[string]interface{}{
				"bar": "baz",
			},
			Create: map[string]string{
				"bar": "baz",
			},
			Remove: map[string]string{
				"foo": "bar",
			},
		},

		// Modify
		{
			Old: map[string]interface{}{
				"foo": "bar",
			},
			New: map[string]interface{}{
				"foo": "baz",
			},
			Create: map[string]string{
				"foo": "baz",
			},
			Remove: map[string]string{
				"foo": "bar",
			},
		},
	}

	for i, tc := range cases {
		c, r := diffTagsDS(tagsFromMapDS(tc.Old), tagsFromMapDS(tc.New))
		cm := tagsToMapDS(c)
		rm := tagsToMapDS(r)
		if !reflect.DeepEqual(cm, tc.Create) {
			t.Fatalf("%d: bad create: %#v", i, cm)
		}
		if !reflect.DeepEqual(rm, tc.Remove) {
			t.Fatalf("%d: bad remove: %#v", i, rm)
		}
	}
}

func TestAccAWSDirectoryServiceDirectory_importBasic(t *testing.T) {
	resourceName := "aws_directory_service_directory.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig,
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

func TestAccAWSDirectoryServiceDirectory_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists("aws_directory_service_directory.bar"),
					resource.TestCheckResourceAttrSet("aws_directory_service_directory.bar", "security_group_id"),
				),
			},
		},
	})
}

func TestAccAWSDirectoryServiceDirectory_tags(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryTagsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists("aws_directory_service_directory.bar"),
					resource.TestCheckResourceAttr("aws_directory_service_directory.bar", "tags.%", "2"),
				),
			},
		},
	})
}

func TestAccAWSDirectoryServiceDirectory_microsoft(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig_microsoft,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists("aws_directory_service_directory.bar"),
					resource.TestCheckResourceAttr("aws_directory_service_directory.bar", "edition", directoryservice.DirectoryEditionEnterprise),
				),
			},
		},
	})
}

func TestAccAWSDirectoryServiceDirectory_microsoftStandard(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig_microsoftStandard,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists("aws_directory_service_directory.bar"),
					resource.TestCheckResourceAttr("aws_directory_service_directory.bar", "edition", directoryservice.DirectoryEditionStandard),
				),
			},
		},
	})
}

func TestAccAWSDirectoryServiceDirectory_connector(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig_connector,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists("aws_directory_service_directory.connector"),
				),
			},
		},
	})
}

func TestAccAWSDirectoryServiceDirectory_withAliasAndSso(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDirectoryServiceDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryServiceDirectoryConfig_withAlias,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists("aws_directory_service_directory.bar_a"),
					testAccCheckServiceDirectoryAlias("aws_directory_service_directory.bar_a",
						fmt.Sprintf("tf-d-%d", randomInteger)),
					testAccCheckServiceDirectorySso("aws_directory_service_directory.bar_a", false),
				),
			},
			{
				Config: testAccDirectoryServiceDirectoryConfig_withSso,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists("aws_directory_service_directory.bar_a"),
					testAccCheckServiceDirectoryAlias("aws_directory_service_directory.bar_a",
						fmt.Sprintf("tf-d-%d", randomInteger)),
					testAccCheckServiceDirectorySso("aws_directory_service_directory.bar_a", true),
				),
			},
			{
				Config: testAccDirectoryServiceDirectoryConfig_withSso_modified,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceDirectoryExists("aws_directory_service_directory.bar_a"),
					testAccCheckServiceDirectoryAlias("aws_directory_service_directory.bar_a",
						fmt.Sprintf("tf-d-%d", randomInteger)),
					testAccCheckServiceDirectorySso("aws_directory_service_directory.bar_a", false),
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
		if err != nil {
			// EntityDoesNotExistException means it's gone, this is good
			if dserr, ok := err.(awserr.Error); ok && dserr.Code() == "EntityDoesNotExistException" {
				return nil
			}
			return err
		}

		if out != nil && len(out.DirectoryDescriptions) > 0 {
			return fmt.Errorf("Expected AWS Directory Service Directory to be gone, but was still found")
		}

		return nil
	}

	return fmt.Errorf("Default error in Service Directory Test")
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

const testAccDirectoryServiceDirectoryConfig = `
resource "aws_directory_service_directory" "bar" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size = "Small"

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
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
  availability_zone = "us-west-2a"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-foo"
  }
}
resource "aws_subnet" "bar" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "us-west-2b"
  cidr_block = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-bar"
  }
}
`

const testAccDirectoryServiceDirectoryTagsConfig = `
resource "aws_directory_service_directory" "bar" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size = "Small"

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
  }

	tags = {
		foo = "bar"
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
  availability_zone = "us-west-2a"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-tags-foo"
  }
}
resource "aws_subnet" "bar" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "us-west-2b"
  cidr_block = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-tags-bar"
  }
}
`

const testAccDirectoryServiceDirectoryConfig_connector = `
resource "aws_directory_service_directory" "bar" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size = "Small"

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
  }
}

resource "aws_directory_service_directory" "connector" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size = "Small"
  type = "ADConnector"

  connect_settings {
    customer_dns_ips = ["${aws_directory_service_directory.bar.dns_ip_addresses}"]
    customer_username = "Administrator"
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
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
  availability_zone = "us-west-2a"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-connector-foo"
  }
}
resource "aws_subnet" "bar" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "us-west-2b"
  cidr_block = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-connector-bar"
  }
}
`

const testAccDirectoryServiceDirectoryConfig_microsoft = `
resource "aws_directory_service_directory" "bar" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type = "MicrosoftAD"

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
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
  availability_zone = "us-west-2a"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-microsoft-foo"
  }
}
resource "aws_subnet" "bar" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "us-west-2b"
  cidr_block = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-microsoft-bar"
  }
}
`

const testAccDirectoryServiceDirectoryConfig_microsoftStandard = `
resource "aws_directory_service_directory" "bar" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type = "MicrosoftAD"
  edition = "Standard"

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
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
  availability_zone = "us-west-2a"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-microsoft-foo"
  }
}
resource "aws_subnet" "bar" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "us-west-2b"
  cidr_block = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-microsoft-bar"
  }
}
`

var randomInteger = acctest.RandInt()
var testAccDirectoryServiceDirectoryConfig_withAlias = fmt.Sprintf(`
resource "aws_directory_service_directory" "bar_a" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size = "Small"
  alias = "tf-d-%d"

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
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
  availability_zone = "us-west-2a"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-with-alias-foo"
  }
}
resource "aws_subnet" "bar" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "us-west-2b"
  cidr_block = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-with-alias-bar"
  }
}
`, randomInteger)

var testAccDirectoryServiceDirectoryConfig_withSso = fmt.Sprintf(`
resource "aws_directory_service_directory" "bar_a" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size = "Small"
  alias = "tf-d-%d"
  enable_sso = true

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
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
  availability_zone = "us-west-2a"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-with-sso-foo"
  }
}
resource "aws_subnet" "bar" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "us-west-2b"
  cidr_block = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-with-sso-bar"
  }
}
`, randomInteger)

var testAccDirectoryServiceDirectoryConfig_withSso_modified = fmt.Sprintf(`
resource "aws_directory_service_directory" "bar_a" {
  name = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  size = "Small"
  alias = "tf-d-%d"
  enable_sso = false

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
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
  availability_zone = "us-west-2a"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-with-sso-foo"
  }
}
resource "aws_subnet" "bar" {
  vpc_id = "${aws_vpc.main.id}"
  availability_zone = "us-west-2b"
  cidr_block = "10.0.2.0/24"
  tags = {
    Name = "tf-acc-directory-service-directory-with-sso-bar"
  }
}
`, randomInteger)
