package aws

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsWorkspacesWorkspace_byWorkspaceID(t *testing.T) {
	rName := acctest.RandString(8)
	dataSourceName := "data.aws_workspaces_workspace.test"
	resourceName := "aws_workspaces_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckHasIAMRole(t, "workspaces_DefaultRole") },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceWorkspacesWorkspaceConfig_byWorkspaceID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "directory_id", resourceName, "directory_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bundle_id", resourceName, "bundle_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_address", resourceName, "ip_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "state", resourceName, "state"),
					resource.TestCheckResourceAttrPair(dataSourceName, "root_volume_encryption_enabled", resourceName, "root_volume_encryption_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "user_name", resourceName, "user_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume_encryption_key", resourceName, "volume_encryption_key"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.#", resourceName, "workspace_properties.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.compute_type_name", resourceName, "workspace_properties.0.compute_type_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.root_volume_size_gib", resourceName, "workspace_properties.0.root_volume_size_gib"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.running_mode", resourceName, "workspace_properties.0.running_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.running_mode_auto_stop_timeout_in_minutes", resourceName, "workspace_properties.0.running_mode_auto_stop_timeout_in_minutes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.user_volume_size_gib", resourceName, "workspace_properties.0.user_volume_size_gib"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsWorkspacesWorkspace_byDirectoryID_userName(t *testing.T) {
	rName := acctest.RandString(8)
	dataSourceName := "data.aws_workspaces_workspace.test"
	resourceName := "aws_workspaces_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckHasIAMRole(t, "workspaces_DefaultRole") },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceWorkspacesWorkspaceConfig_byDirectoryID_userName(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "directory_id", resourceName, "directory_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bundle_id", resourceName, "bundle_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_address", resourceName, "ip_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "state", resourceName, "state"),
					resource.TestCheckResourceAttrPair(dataSourceName, "root_volume_encryption_enabled", resourceName, "root_volume_encryption_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "user_name", resourceName, "user_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume_encryption_key", resourceName, "volume_encryption_key"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.#", resourceName, "workspace_properties.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.compute_type_name", resourceName, "workspace_properties.0.compute_type_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.root_volume_size_gib", resourceName, "workspace_properties.0.root_volume_size_gib"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.running_mode", resourceName, "workspace_properties.0.running_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.running_mode_auto_stop_timeout_in_minutes", resourceName, "workspace_properties.0.running_mode_auto_stop_timeout_in_minutes"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_properties.0.user_volume_size_gib", resourceName, "workspace_properties.0.user_volume_size_gib"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
				),
				ExpectNonEmptyPlan: true, // Hack to overcome data source with depends_on refresh
			},
		},
	})
}

func TestAccDataSourceAwsWorkspacesWorkspace_workspaceIDAndDirectoryIDConflict(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckHasIAMRole(t, "workspaces_DefaultRole") },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWorkspacesWorkspaceConfig_workspaceIDAndDirectoryIDConflict(),
				ExpectError: regexp.MustCompile("\"workspace_id\": conflicts with directory_id"),
			},
		},
	})
}

func testAccDataSourceWorkspacesWorkspaceConfig_byWorkspaceID(rName string) string {
	return composeConfig(
		testAccAwsWorkspacesWorkspaceConfig_Prerequisites(rName),
		`
resource "aws_workspaces_workspace" "test" {
  bundle_id    = data.aws_workspaces_bundle.test.id
  directory_id = aws_workspaces_directory.test.id

  # NOTE: WorkSpaces API doesn't allow creating users in the directory.
  # However, "AWS_WorkSpaces" user is always present in a bare directory.
  user_name = "AWS_WorkSpaces"

  workspace_properties {
    root_volume_size_gib = 80
    user_volume_size_gib = 10
  }

  tags = {
    TerraformProviderAwsTest = true
  }
}

data "aws_workspaces_workspace" "test" {
  workspace_id = aws_workspaces_workspace.test.id
}
`)
}

func testAccDataSourceWorkspacesWorkspaceConfig_byDirectoryID_userName(rName string) string {
	return composeConfig(
		testAccAwsWorkspacesWorkspaceConfig_Prerequisites(rName),
		`
resource "aws_workspaces_workspace" "test" {
  bundle_id    = data.aws_workspaces_bundle.test.id
  directory_id = aws_workspaces_directory.test.id

  # NOTE: WorkSpaces API doesn't allow creating users in the directory.
  # However, "Administrator" user is always present in a bare directory.
  user_name = "Administrator"

  workspace_properties {
    root_volume_size_gib = 80
    user_volume_size_gib = 10
  }

  tags = {
    TerraformProviderAwsTest = true
  }
}

data "aws_workspaces_workspace" "test" {
  directory_id = aws_workspaces_directory.test.id
  user_name    = "Administrator"

  depends_on = [aws_workspaces_workspace.test]
}
`)
}

func testAccDataSourceAwsWorkspacesWorkspaceConfig_workspaceIDAndDirectoryIDConflict() string {
	return `
data "aws_workspaces_workspace" "test" {
  workspace_id = "ws-cj5xcxsz5"
  directory_id = "d-9967252f57"
  user_name    = "Administrator"
}
`
}
