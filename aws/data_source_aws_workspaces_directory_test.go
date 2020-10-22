package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsWorkspacesDirectory_basic(t *testing.T) {
	rName := acctest.RandString(8)

	resourceName := "aws_workspaces_directory.test"
	dataSourceName := "data.aws_workspaces_directory.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckHasIAMRole(t, "workspaces_DefaultRole") },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsWorkspacesDirectoryConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "alias", resourceName, "alias"),
					resource.TestCheckResourceAttrPair(dataSourceName, "directory_id", resourceName, "directory_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "directory_name", resourceName, "directory_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "directory_type", resourceName, "directory_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_ip_addresses.#", resourceName, "dns_ip_addresses.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "iam_role_id", resourceName, "iam_role_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_group_ids", resourceName, "ip_group_ids"),
					resource.TestCheckResourceAttrPair(dataSourceName, "registration_code", resourceName, "registration_code"),
					resource.TestCheckResourceAttrPair(dataSourceName, "self_service_permissions.#", resourceName, "self_service_permissions.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "self_service_permissions.0.change_compute_type", resourceName, "self_service_permissions.0.change_compute_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "self_service_permissions.0.increase_volume_size", resourceName, "self_service_permissions.0.increase_volume_size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "self_service_permissions.0.rebuild_workspace", resourceName, "self_service_permissions.0.rebuild_workspace"),
					resource.TestCheckResourceAttrPair(dataSourceName, "self_service_permissions.0.restart_workspace", resourceName, "self_service_permissions.0.restart_workspace"),
					resource.TestCheckResourceAttrPair(dataSourceName, "self_service_permissions.0.switch_running_mode", resourceName, "self_service_permissions.0.switch_running_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_creation_properties.#", resourceName, "workspace_creation_properties.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_creation_properties.0.custom_security_group_id", resourceName, "workspace_creation_properties.0.custom_security_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_creation_properties.0.default_ou", resourceName, "workspace_creation_properties.0.default_ou"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_creation_properties.0.enable_internet_access", resourceName, "workspace_creation_properties.0.enable_internet_access"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_creation_properties.0.enable_maintenance_mode", resourceName, "workspace_creation_properties.0.enable_maintenance_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_creation_properties.0.user_enabled_as_local_administrator", resourceName, "workspace_creation_properties.0.user_enabled_as_local_administrator"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workspace_security_group_id", resourceName, "workspace_security_group_id"),
				),
			},
		},
	})
}

func testAccDataSourceAwsWorkspacesDirectoryConfig(rName string) string {
	return composeConfig(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName),
		`
resource "aws_workspaces_directory" "test" {
  directory_id = aws_directory_service_directory.main.id

  self_service_permissions {
    change_compute_type  = false
    increase_volume_size = true
    rebuild_workspace    = true
    restart_workspace    = false
    switch_running_mode  = true
  }
}

data "aws_workspaces_directory" "test" {
  directory_id = aws_workspaces_directory.test.id
}

data "aws_iam_role" "workspaces-default" {
  name = "workspaces_DefaultRole"
}
`)
}
