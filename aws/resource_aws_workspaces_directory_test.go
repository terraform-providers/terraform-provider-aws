package aws

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/workspaces"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_workspaces_directory", &resource.Sweeper{
		Name:         "aws_workspaces_directory",
		F:            testSweepWorkspacesDirectories,
		Dependencies: []string{"aws_workspaces_workspace"},
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

func TestAccAwsWorkspacesDirectory_basic(t *testing.T) {
	var v workspaces.WorkspaceDirectory
	rName := acctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"
	directoryResourceName := "aws_directory_service_directory.main"
	iamRoleDataSourceName := "data.aws_iam_role.workspaces-default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			testAccPreCheckAWSDirectoryServiceSimpleDirectory(t)
			testAccPreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDirectoryConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "alias", directoryResourceName, "alias"),
					resource.TestCheckResourceAttrPair(resourceName, "directory_id", directoryResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "directory_name", directoryResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "directory_type", workspaces.WorkspaceDirectoryTypeSimpleAd),
					resource.TestCheckResourceAttr(resourceName, "dns_ip_addresses.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "iam_role_id", iamRoleDataSourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "registration_code"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.change_compute_type", "false"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.increase_volume_size", "false"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.rebuild_workspace", "false"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.restart_workspace", "true"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.switch_running_mode", "false"),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.custom_security_group_id", ""),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.default_ou", ""),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.enable_internet_access", "false"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.enable_maintenance_mode", "true"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.user_enabled_as_local_administrator", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "workspace_security_group_id"),
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

func TestAccAwsWorkspacesDirectory_disappears(t *testing.T) {
	var v workspaces.WorkspaceDirectory
	rName := acctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			testAccPreCheckAWSDirectoryServiceSimpleDirectory(t)
			testAccPreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDirectoryConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					testAccCheckAwsWorkspacesDirectoryDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsWorkspacesDirectory_subnetIds(t *testing.T) {
	var v workspaces.WorkspaceDirectory
	rName := acctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			testAccPreCheckAWSDirectoryServiceSimpleDirectory(t)
			testAccPreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDirectoryConfig_subnetIds(rName),
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

func TestAccAwsWorkspacesDirectory_tags(t *testing.T) {
	var v workspaces.WorkspaceDirectory
	rName := acctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			testAccPreCheckAWSDirectoryServiceSimpleDirectory(t)
			testAccPreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDirectoryConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
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
				Config: testAccWorkspacesDirectoryConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccWorkspacesDirectoryConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAwsWorkspacesDirectory_selfServicePermissions(t *testing.T) {
	var v workspaces.WorkspaceDirectory
	rName := acctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			testAccPreCheckAWSDirectoryServiceSimpleDirectory(t)
			testAccPreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDirectory_selfServicePermissions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.change_compute_type", "false"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.increase_volume_size", "true"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.rebuild_workspace", "true"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.restart_workspace", "false"),
					resource.TestCheckResourceAttr(resourceName, "self_service_permissions.0.switch_running_mode", "true"),
				),
			},
		},
	})
}

func TestAccAwsWorkspacesDirectory_workspaceCreationProperties(t *testing.T) {
	var v workspaces.WorkspaceDirectory
	rName := acctest.RandString(8)

	resourceName := "aws_workspaces_directory.main"
	resourceSecurityGroup := "aws_security_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			testAccPreCheckAWSDirectoryServiceSimpleDirectory(t)
			testAccPreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesDirectoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesDirectoryConfig_workspaceCreationProperties(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesDirectoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "workspace_creation_properties.0.custom_security_group_id", resourceSecurityGroup, "id"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.default_ou", "OU=AWS,DC=Workgroup,DC=Example,DC=com"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.enable_internet_access", "true"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.enable_maintenance_mode", "false"),
					resource.TestCheckResourceAttr(resourceName, "workspace_creation_properties.0.user_enabled_as_local_administrator", "false"),
				),
			},
		},
	})
}

func testAccPreCheckHasIAMRole(t *testing.T, roleName string) {
	conn := testAccProvider.Meta().(*AWSClient).iamconn

	input := &iam.GetRoleInput{
		RoleName: aws.String(roleName),
	}
	_, err := conn.GetRole(input)

	if isAWSErr(err, iam.ErrCodeNoSuchEntityException, "") {
		t.Skipf("skipping acceptance test: required IAM role \"%s\" is not present", roleName)
	}
	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance test: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
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

func testAccPreCheckWorkspacesDirectory(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).workspacesconn

	input := &workspaces.DescribeWorkspaceDirectoriesInput{}

	_, err := conn.DescribeWorkspaceDirectories(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName string) string {
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
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}

resource "aws_subnet" "primary" {
  vpc_id               = aws_vpc.main.id
  availability_zone_id = local.workspaces_az_ids[0]
  cidr_block           = "10.0.1.0/24"

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s-primary"
  }
}

resource "aws_subnet" "secondary" {
  vpc_id               = aws_vpc.main.id
  availability_zone_id = local.workspaces_az_ids[1]
  cidr_block           = "10.0.2.0/24"

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s-secondary"
  }
}

resource "aws_directory_service_directory" "main" {
  size     = "Small"
  name     = "tf-acctest.neverland.com"
  password = "#S1ncerely"

  vpc_settings {
    vpc_id     = aws_vpc.main.id
    subnet_ids = [aws_subnet.primary.id, aws_subnet.secondary.id]
  }

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}
`, rName)
}

func testAccWorkspacesDirectoryConfig(rName string) string {
	return composeConfig(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName),
		`
resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id
}

data "aws_iam_role" "workspaces-default" {
  name = "workspaces_DefaultRole"
}
`)
}

func testAccWorkspacesDirectory_selfServicePermissions(rName string) string {
	return composeConfig(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName),
		`
resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id

  self_service_permissions {
    change_compute_type  = false
    increase_volume_size = true
    rebuild_workspace    = true
    restart_workspace    = false
    switch_running_mode  = true
  }
}
`)
}

func testAccWorkspacesDirectoryConfig_subnetIds(rName string) string {
	return composeConfig(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName),
		`
resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id
  subnet_ids   = [aws_subnet.primary.id, aws_subnet.secondary.id]
}
`)
}

func testAccWorkspacesDirectoryConfigTags1(rName, tagKey1, tagValue1 string) string {
	return composeConfig(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName),
		fmt.Sprintf(`
resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccWorkspacesDirectoryConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName),
		fmt.Sprintf(`
resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccWorkspacesDirectoryConfig_workspaceCreationProperties(rName string) string {
	return composeConfig(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName),
		fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = "tf-acctest-%[1]s"
  vpc_id = aws_vpc.main.id
}

resource "aws_workspaces_directory" "main" {
  directory_id = aws_directory_service_directory.main.id

  workspace_creation_properties {
    custom_security_group_id            = aws_security_group.test.id
    default_ou                          = "OU=AWS,DC=Workgroup,DC=Example,DC=com"
    enable_internet_access              = true
    enable_maintenance_mode             = false
    user_enabled_as_local_administrator = false
  }
}
`, rName))
}
