package aws

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_workspaces_directory", &resource.Sweeper{
		Name: "aws_workspaces_directory",
		F:    testSweepWorkspacesDirectories,
	})
}

func testSweepWorkspacesDirectories(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).workspacesconn

	var errors error
	input := &workspaces.DescribeWorkspaceDirectoriesInput{}
	err = conn.DescribeWorkspaceDirectoriesPages(input, func(resp *workspaces.DescribeWorkspaceDirectoriesOutput, _ bool) bool {
		for _, directory := range resp.Directories {
			err := workspacesDirectoryDelete(aws.StringValue(directory.DirectoryId), conn)
			if err != nil {
				errors = multierror.Append(errors, err)
			}

		}
		return true
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Workspace Directory sweep for %s: %s", region, err)
		return errors // In case we have completed some pages, but had errors
	}
	if err != nil {
		errors = multierror.Append(errors, fmt.Errorf("error listing Workspace Directories: %s", err))
	}

	return errors
}

// These tests need to be serialized, because they all rely on the IAM Role `workspaces_DefaultRole`.
func TestAccAwsWorkspacesDirectory(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":      testAccAwsWorkspacesDirectory_basic,
		"disappears": testAccAwsWorkspacesDirectory_disappears,
		"subnetIds":  testAccAwsWorkspacesDirectory_subnetIds,
	}
	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccAwsWorkspacesDirectory_basic(t *testing.T) {
	var v workspaces.WorkspaceDirectory
	booster := acctest.RandString(8)
	resourceName := "aws_workspaces_directory.main"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDirectoryConfigA(booster),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.change_compute_type", "false"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.increase_volume_size", "false"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.rebuild_workspace", "false"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.restart_workspace", "true"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.switch_running_mode", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.Terraform", "true"),
					resource.TestCheckResourceAttr(resourceName, "tags.Directory", "tf-acctest.example.com"),
				),
			},
			{
				Config: testAccWorkspacesDirectoryConfigB(booster),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.change_compute_type", "false"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.increase_volume_size", "true"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.rebuild_workspace", "true"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.restart_workspace", "false"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.switch_running_mode", "true"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Directory", "tf-acctest.example.com"),
					resource.TestCheckResourceAttr(resourceName, "tags.Purpose", "test"),
				),
			},
			{
				Config: testAccWorkspacesDirectoryConfigC(booster),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.change_compute_type", "true"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.increase_volume_size", "false"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.rebuild_workspace", "false"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.restart_workspace", "true"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.switch_running_mode", "true"),
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

func testAccAwsWorkspacesDirectory_disappears(t *testing.T) {
	var v workspaces.WorkspaceDirectory
	booster := acctest.RandString(8)
	resourceName := "aws_workspaces_directory.main"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDirectoryConfigA(booster),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					testAccCheckAwsWorkspacesDirectoryDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAwsWorkspacesDirectory_subnetIds(t *testing.T) {
	var v workspaces.WorkspaceDirectory
	booster := acctest.RandString(8)
	resourceName := "aws_workspaces_directory.main"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDirectoryConfig_subnetIds(booster),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
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

func testAccCheckAwsWorkspacesDirectoryDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).workspacesconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_workspaces_directory" {
			continue
		}

		resp, err := conn.DescribeWorkspaceDirectories(&workspaces.DescribeWorkspaceDirectoriesInput{
			DirectoryIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}

		if len(resp.Directories) == 0 {
			return nil
		}

		dir := resp.Directories[0]
		if *dir.State != workspaces.WorkspaceDirectoryStateDeregistering && *dir.State != workspaces.WorkspaceDirectoryStateDeregistered {
			return fmt.Errorf("directory %q was not deregistered", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsWorkspacesDirectoryDisappears(v *workspaces.WorkspaceDirectory) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return workspacesDirectoryDelete(aws.StringValue(v.DirectoryId), testAccProvider.Meta().(*AWSClient).workspacesconn)
	}
}

func testAccCheckAwsWorkspacesDirectoryExists(n string, v *workspaces.WorkspaceDirectory) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("workspaces directory resource is not found: %q", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("workspaces directory resource ID is not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).workspacesconn
		resp, err := conn.DescribeWorkspaceDirectories(&workspaces.DescribeWorkspaceDirectoriesInput{
			DirectoryIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}

		if *resp.Directories[0].DirectoryId == rs.Primary.ID {
			*v = *resp.Directories[0]
			return nil
		}

		return fmt.Errorf("workspaces directory %q is not found", rs.Primary.ID)
	}
}

func TestExpandSelfServicePermissions(t *testing.T) {
	cases := []struct {
		input    []interface{}
		expected *workspaces.SelfservicePermissions
	}{
		// Empty
		{
			input:    []interface{}{},
			expected: nil,
		},
		// Full
		{
			input: []interface{}{
				map[string]interface{}{
					"change_compute_type":  false,
					"increase_volume_size": false,
					"rebuild_workspace":    true,
					"restart_workspace":    true,
					"switch_running_mode":  true,
				},
			},
			expected: &workspaces.SelfservicePermissions{
				ChangeComputeType:  aws.String(workspaces.ReconnectEnumDisabled),
				IncreaseVolumeSize: aws.String(workspaces.ReconnectEnumDisabled),
				RebuildWorkspace:   aws.String(workspaces.ReconnectEnumEnabled),
				RestartWorkspace:   aws.String(workspaces.ReconnectEnumEnabled),
				SwitchRunningMode:  aws.String(workspaces.ReconnectEnumEnabled),
			},
		},
	}

	for _, c := range cases {
		actual := expandSelfServicePermissions(c.input)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Fatalf("expected\n\n%#+v\n\ngot\n\n%#+v", c.expected, actual)
		}
	}
}

func TestFlattenSelfServicePermissions(t *testing.T) {
	cases := []struct {
		input    *workspaces.SelfservicePermissions
		expected []interface{}
	}{
		// Empty
		{
			input:    nil,
			expected: []interface{}{},
		},
		// Full
		{
			input: &workspaces.SelfservicePermissions{
				ChangeComputeType:  aws.String(workspaces.ReconnectEnumDisabled),
				IncreaseVolumeSize: aws.String(workspaces.ReconnectEnumDisabled),
				RebuildWorkspace:   aws.String(workspaces.ReconnectEnumEnabled),
				RestartWorkspace:   aws.String(workspaces.ReconnectEnumEnabled),
				SwitchRunningMode:  aws.String(workspaces.ReconnectEnumEnabled),
			},
			expected: []interface{}{
				map[string]interface{}{
					"change_compute_type":  false,
					"increase_volume_size": false,
					"rebuild_workspace":    true,
					"restart_workspace":    true,
					"switch_running_mode":  true,
				},
			},
		},
	}

	for _, c := range cases {
		actual := flattenSelfServicePermissions(c.input)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Fatalf("expected\n\n%#+v\n\ngot\n\n%#+v", c.expected, actual)
		}
	}
}

// Extract common infra
func testAccAwsWorkspacesDirectoryConfig_Prerequisites(booster string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  region_workspaces_az_ids = {
    "us-east-1" = formatlist("use1-az%%d", [2, 4, 6])
  }

  workspaces_az_ids = lookup(local.region_workspaces_az_ids, data.aws_region.current.name, data.aws_availability_zones.available.zone_ids)
}

 resource "aws_vpc" "main" {
   cidr_block = "10.0.0.0/16"

   tags = {
     Name = "tf-testacc-workspaces-directory-%s"
   }
 }

 resource "aws_subnet" "primary" {
   vpc_id = "${aws_vpc.main.id}"
   availability_zone_id = "${local.workspaces_az_ids[0]}"
   cidr_block = "10.0.1.0/24"

   tags = {
     Name = "tf-testacc-workspaces-directory-%s-primary"
   }
 }

 resource "aws_subnet" "secondary" {
   vpc_id = "${aws_vpc.main.id}"
   availability_zone_id = "${local.workspaces_az_ids[1]}"
   cidr_block = "10.0.2.0/24"

   tags = {
     Name = "tf-testacc-workspaces-directory-%s-secondary"
   }
 }

resource "aws_directory_service_directory" "main" {
  size = "Small"
  name = "tf-acctest.neverland.com"
  password = "#S1ncerely"

  vpc_settings {
    vpc_id = "${aws_vpc.main.id}"
    subnet_ids = ["${aws_subnet.primary.id}","${aws_subnet.secondary.id}"]
  }
}

data "aws_iam_policy_document" "workspaces" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["workspaces.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "workspaces-default" {
  name               = "workspaces_DefaultRole"
  assume_role_policy = data.aws_iam_policy_document.workspaces.json
}

resource "aws_iam_role_policy_attachment" "workspaces-default-service-access" {
  role       = aws_iam_role.workspaces-default.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonWorkSpacesServiceAccess"
}

resource "aws_iam_role_policy_attachment" "workspaces-default-self-service-access" {
  role       = aws_iam_role.workspaces-default.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonWorkSpacesSelfServiceAccess"
}
`, booster, booster, booster)
}

func testAccWorkspacesDirectoryConfigA(booster string) string {
	return testAccAwsWorkspacesDirectoryConfig_Prerequisites(booster) + fmt.Sprintf(`
resource "aws_workspaces_directory" "main" {
  directory_id = "${aws_directory_service_directory.main.id}"

  tags = {
    Name = "test"
    Terraform = true
    Directory = "tf-acctest.example.com"
  }
}
`)
}

func testAccWorkspacesDirectoryConfigB(booster string) string {
	return testAccAwsWorkspacesDirectoryConfig_Prerequisites(booster) + fmt.Sprintf(`
resource "aws_workspaces_directory" "main" {
  directory_id = "${aws_directory_service_directory.main.id}"

  self_service_permissions {
    change_compute_type = false
    increase_volume_size = true
    rebuild_workspace = true
    restart_workspace = false
    switch_running_mode = true
  }

  tags = {
    Purpose   = "test"
    Directory = "tf-acctest.example.com"
  }
}
`)
}

func testAccWorkspacesDirectoryConfigC(booster string) string {
	return testAccAwsWorkspacesDirectoryConfig_Prerequisites(booster) + fmt.Sprintf(`
resource "aws_workspaces_directory" "main" {
  directory_id = "${aws_directory_service_directory.main.id}"

  self_service_permissions {
    change_compute_type = true
    switch_running_mode = true
  }
}
`)
}

func testAccWorkspacesDirectoryConfig_subnetIds(booster string) string {
	return testAccAwsWorkspacesDirectoryConfig_Prerequisites(booster) + fmt.Sprintf(`
resource "aws_workspaces_directory" "main" {
  directory_id = "${aws_directory_service_directory.main.id}"
  subnet_ids = ["${aws_subnet.primary.id}","${aws_subnet.secondary.id}"]
}
`)
}
