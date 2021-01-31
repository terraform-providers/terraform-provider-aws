package aws

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/workspaces/waiter"
)

func init() {
	resource.AddTestSweepers("aws_workspaces_workspace", &resource.Sweeper{
		Name: "aws_workspaces_workspace",
		F:    testSweepWorkspacesWorkspace,
	})
}

func testSweepWorkspacesWorkspace(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).workspacesconn

	var errors error
	input := &workspaces.DescribeWorkspacesInput{}
	err = conn.DescribeWorkspacesPages(input, func(resp *workspaces.DescribeWorkspacesOutput, _ bool) bool {
		for _, workspace := range resp.Workspaces {
			err := workspaceDelete(conn, aws.StringValue(workspace.WorkspaceId), waiter.WorkspaceTerminatedTimeout)
			if err != nil {
				errors = multierror.Append(errors, err)
			}

		}
		return true
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping workspaces sweep for %s: %s", region, err)
		return errors // In case we have completed some pages, but had errors
	}
	if err != nil {
		errors = multierror.Append(errors, fmt.Errorf("error listing workspaces: %s", err))
	}

	return errors
}

func TestAccAwsWorkspacesWorkspace_basic(t *testing.T) {
	var v workspaces.Workspace
	rName := acctest.RandString(8)

	resourceName := "aws_workspaces_workspace.test"
	directoryResourceName := "aws_workspaces_directory.test"
	bundleDataSourceName := "data.aws_workspaces_bundle.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			testAccPreCheckAWSDirectoryServiceSimpleDirectory(t)
			testAccPreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Destroy: false,
				Config:  testAccWorkspacesWorkspaceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesWorkspaceExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "directory_id", directoryResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "bundle_id", bundleDataSourceName, "id"),
					resource.TestMatchResourceAttr(resourceName, "ip_address", regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)),
					resource.TestCheckResourceAttr(resourceName, "state", workspaces.WorkspaceStateAvailable),
					resource.TestCheckResourceAttr(resourceName, "root_volume_encryption_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "user_name", "Administrator"),
					resource.TestCheckResourceAttr(resourceName, "volume_encryption_key", ""),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.compute_type_name", workspaces.ComputeValue),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.root_volume_size_gib", "80"),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.running_mode", workspaces.RunningModeAlwaysOn),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.running_mode_auto_stop_timeout_in_minutes", "0"),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.user_volume_size_gib", "10"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", fmt.Sprintf("tf-testacc-workspaces-workspace-%[1]s", rName)),
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

func TestAccAwsWorkspacesWorkspace_tags(t *testing.T) {
	var v1, v2, v3 workspaces.Workspace
	rName := acctest.RandString(8)

	resourceName := "aws_workspaces_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			testAccPreCheckAWSDirectoryServiceSimpleDirectory(t)
			testAccPreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesWorkspaceConfig_TagsA(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesWorkspaceExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", fmt.Sprintf("tf-testacc-workspaces-workspace-%[1]s", rName)),
					resource.TestCheckResourceAttr(resourceName, "tags.Alpha", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkspacesWorkspaceConfig_TagsB(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesWorkspaceExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", fmt.Sprintf("tf-testacc-workspaces-workspace-%[1]s", rName)),
					resource.TestCheckResourceAttr(resourceName, "tags.Beta", "2"),
				),
			},
			{
				Config: testAccWorkspacesWorkspaceConfig_TagsC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesWorkspaceExists(resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", fmt.Sprintf("tf-testacc-workspaces-workspace-%[1]s", rName)),
				),
			},
		},
	})
}

func TestAccAwsWorkspacesWorkspace_workspaceProperties(t *testing.T) {
	var v1, v2, v3 workspaces.Workspace
	rName := acctest.RandString(8)

	resourceName := "aws_workspaces_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			testAccPreCheckAWSDirectoryServiceSimpleDirectory(t)
			testAccPreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Destroy: false,
				Config:  testAccWorkspacesWorkspaceConfig_WorkspacePropertiesA(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesWorkspaceExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.compute_type_name", workspaces.ComputeValue),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.root_volume_size_gib", "80"),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.running_mode", workspaces.RunningModeAutoStop),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.running_mode_auto_stop_timeout_in_minutes", "120"),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.user_volume_size_gib", "10"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccWorkspacesWorkspaceConfig_WorkspacePropertiesB(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesWorkspaceExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.compute_type_name", workspaces.ComputeValue),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.root_volume_size_gib", "80"),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.running_mode", workspaces.RunningModeAlwaysOn),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.running_mode_auto_stop_timeout_in_minutes", "0"),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.user_volume_size_gib", "10"),
				),
			},
			{
				Config: testAccWorkspacesWorkspaceConfig_WorkspacePropertiesC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesWorkspaceExists(resourceName, &v3),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.compute_type_name", workspaces.ComputeValue),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.root_volume_size_gib", "80"),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.running_mode", workspaces.RunningModeAlwaysOn),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.running_mode_auto_stop_timeout_in_minutes", "0"),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.user_volume_size_gib", "10"),
				),
			},
		},
	})
}

// TestAccAwsWorkspacesWorkspace_workspaceProperties_runningModeAlwaysOn
// validates workspace resource creation/import when workspace_properties.running_mode is set to ALWAYS_ON
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13558
func TestAccAwsWorkspacesWorkspace_workspaceProperties_runningModeAlwaysOn(t *testing.T) {
	var v1 workspaces.Workspace
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_workspaces_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			testAccPreCheckAWSDirectoryServiceSimpleDirectory(t)
			testAccPreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesWorkspaceConfig_WorkspacePropertiesB(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesWorkspaceExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.compute_type_name", workspaces.ComputeValue),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.root_volume_size_gib", "80"),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.running_mode", workspaces.RunningModeAlwaysOn),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.running_mode_auto_stop_timeout_in_minutes", "0"),
					resource.TestCheckResourceAttr(resourceName, "workspace_properties.0.user_volume_size_gib", "10"),
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

func TestAccAwsWorkspacesWorkspace_validateRootVolumeSize(t *testing.T) {
	rName := acctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccWorkspacesWorkspaceConfig_validateRootVolumeSize(rName),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("expected workspace_properties.0.root_volume_size_gib to be one of [80], got 90")),
			},
		},
	})
}

func TestAccAwsWorkspacesWorkspace_validateUserVolumeSize(t *testing.T) {
	rName := acctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccWorkspacesWorkspaceConfig_validateUserVolumeSize(rName),
				ExpectError: regexp.MustCompile(regexp.QuoteMeta("workspace_properties.0.user_volume_size_gib to be one of [10 50], got 60")),
			},
		},
	})
}

func TestAccAwsWorkspacesWorkspace_recreate(t *testing.T) {
	var v workspaces.Workspace
	rName := acctest.RandString(8)

	resourceName := "aws_workspaces_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			testAccPreCheckAWSDirectoryServiceSimpleDirectory(t)
			testAccPreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkspacesWorkspaceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesWorkspaceExists(resourceName, &v),
				),
			},
			{
				Taint:  []string{resourceName}, // Force workspace re-creation
				Config: testAccWorkspacesWorkspaceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesWorkspaceExists(resourceName, &v),
				),
			},
		},
	})
}

func TestAccAwsWorkspacesWorkspace_timeout(t *testing.T) {
	var v workspaces.Workspace
	rName := acctest.RandString(8)

	resourceName := "aws_workspaces_workspace.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckWorkspacesDirectory(t)
			testAccPreCheckAWSDirectoryServiceSimpleDirectory(t)
			testAccPreCheckHasIAMRole(t, "workspaces_DefaultRole")
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWorkspacesWorkspaceDestroy,
		Steps: []resource.TestStep{
			{
				Destroy: false,
				Config:  testAccWorkspacesWorkspaceConfig_timeout(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsWorkspacesWorkspaceExists(resourceName, &v),
				),
			},
		},
	})
}

func testAccCheckAwsWorkspacesWorkspaceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).workspacesconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_workspaces_workspace" {
			continue
		}

		resp, err := conn.DescribeWorkspaces(&workspaces.DescribeWorkspacesInput{
			WorkspaceIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}

		if len(resp.Workspaces) == 0 {
			return nil
		}
		ws := resp.Workspaces[0]

		if *ws.State != workspaces.WorkspaceStateTerminating && *ws.State != workspaces.WorkspaceStateTerminated {
			return fmt.Errorf("workspace %q was not terminated", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsWorkspacesWorkspaceExists(n string, v *workspaces.Workspace) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).workspacesconn

		output, err := conn.DescribeWorkspaces(&workspaces.DescribeWorkspacesInput{
			WorkspaceIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}

		if *output.Workspaces[0].WorkspaceId == rs.Primary.ID {
			*v = *output.Workspaces[0]
			return nil
		}

		return fmt.Errorf("workspace %q not found", rs.Primary.ID)
	}
}

func testAccAwsWorkspacesWorkspaceConfig_Prerequisites(rName string) string {
	return composeConfig(
		testAccAwsWorkspacesDirectoryConfig_Prerequisites(rName),
		fmt.Sprintf(`
data "aws_workspaces_bundle" "test" {
  bundle_id = "wsb-bh8rsxt14" # Value with Windows 10 (English)
}

resource "aws_workspaces_directory" "test" {
  directory_id = aws_directory_service_directory.main.id

  tags = {
    Name = "tf-testacc-workspaces-directory-%[1]s"
  }
}
`, rName))
}

func testAccWorkspacesWorkspaceConfig(rName string) string {
	return composeConfig(
		testAccAwsWorkspacesWorkspaceConfig_Prerequisites(rName),
		fmt.Sprintf(`
resource "aws_workspaces_workspace" "test" {
  bundle_id    = data.aws_workspaces_bundle.test.id
  directory_id = aws_workspaces_directory.test.id

  # NOTE: WorkSpaces API doesn't allow creating users in the directory.
  # However, "Administrator"" user is always present in a bare directory.
  user_name = "Administrator"

  tags = {
    Name = "tf-testacc-workspaces-workspace-%[1]s"
  }
}
`, rName))
}

func testAccWorkspacesWorkspaceConfig_TagsA(rName string) string {
	return composeConfig(
		testAccAwsWorkspacesWorkspaceConfig_Prerequisites(rName),
		fmt.Sprintf(`
resource "aws_workspaces_workspace" "test" {
  bundle_id    = data.aws_workspaces_bundle.test.id
  directory_id = aws_workspaces_directory.test.id

  # NOTE: WorkSpaces API doesn't allow creating users in the directory.
  # However, "Administrator"" user is always present in a bare directory.
  user_name = "Administrator"

  tags = {
    Name  = "tf-testacc-workspaces-workspace-%[1]s"
    Alpha = 1
  }
}
`, rName))
}

func testAccWorkspacesWorkspaceConfig_TagsB(rName string) string {
	return composeConfig(
		testAccAwsWorkspacesWorkspaceConfig_Prerequisites(rName),
		fmt.Sprintf(`
resource "aws_workspaces_workspace" "test" {
  bundle_id    = data.aws_workspaces_bundle.test.id
  directory_id = aws_workspaces_directory.test.id

  # NOTE: WorkSpaces API doesn't allow creating users in the directory.
  # However, "Administrator"" user is always present in a bare directory.
  user_name = "Administrator"

  tags = {
    Name = "tf-testacc-workspaces-workspace-%[1]s"
    Beta = 2
  }
}
`, rName))
}

func testAccWorkspacesWorkspaceConfig_TagsC(rName string) string {
	return composeConfig(
		testAccAwsWorkspacesWorkspaceConfig_Prerequisites(rName),
		fmt.Sprintf(`
resource "aws_workspaces_workspace" "test" {
  bundle_id    = data.aws_workspaces_bundle.test.id
  directory_id = aws_workspaces_directory.test.id

  # NOTE: WorkSpaces API doesn't allow creating users in the directory.
  # However, "Administrator"" user is always present in a bare directory.
  user_name = "Administrator"

  tags = {
    Name = "tf-testacc-workspaces-workspace-%[1]s"
  }
}
`, rName))
}

func testAccWorkspacesWorkspaceConfig_WorkspacePropertiesA(rName string) string {
	return composeConfig(
		testAccAwsWorkspacesWorkspaceConfig_Prerequisites(rName),
		fmt.Sprintf(`
resource "aws_workspaces_workspace" "test" {
  bundle_id    = data.aws_workspaces_bundle.test.id
  directory_id = aws_workspaces_directory.test.id

  # NOTE: WorkSpaces API doesn't allow creating users in the directory.
  # However, "Administrator"" user is always present in a bare directory.
  user_name = "Administrator"

  workspace_properties {
    # NOTE: Compute type and volume size update not allowed within 6 hours after creation.
    running_mode                              = "AUTO_STOP"
    running_mode_auto_stop_timeout_in_minutes = 120
  }

  tags = {
    Name = "tf-testacc-workspaces-workspace-%[1]s"
  }
}
`, rName))
}

func testAccWorkspacesWorkspaceConfig_WorkspacePropertiesB(rName string) string {
	return composeConfig(
		testAccAwsWorkspacesWorkspaceConfig_Prerequisites(rName),
		fmt.Sprintf(`
resource "aws_workspaces_workspace" "test" {
  bundle_id    = data.aws_workspaces_bundle.test.id
  directory_id = aws_workspaces_directory.test.id

  # NOTE: WorkSpaces API doesn't allow creating users in the directory.
  # However, "Administrator"" user is always present in a bare directory.
  user_name = "Administrator"

  workspace_properties {
    # NOTE: Compute type and volume size update not allowed within 6 hours after creation.
    running_mode = "ALWAYS_ON"
  }

  tags = {
    Name = "tf-testacc-workspaces-workspace-%[1]s"
  }
}
`, rName))
}

func testAccWorkspacesWorkspaceConfig_WorkspacePropertiesC(rName string) string {
	return composeConfig(
		testAccAwsWorkspacesWorkspaceConfig_Prerequisites(rName),
		fmt.Sprintf(`
resource "aws_workspaces_workspace" "test" {
  bundle_id    = data.aws_workspaces_bundle.test.id
  directory_id = aws_workspaces_directory.test.id

  # NOTE: WorkSpaces API doesn't allow creating users in the directory.
  # However, "Administrator"" user is always present in a bare directory.
  user_name = "Administrator"

  workspace_properties {
  }

  tags = {
    Name = "tf-testacc-workspaces-workspace-%[1]s"
  }
}
`, rName))
}

func testAccWorkspacesWorkspaceConfig_validateRootVolumeSize(rName string) string {
	return composeConfig(
		testAccAwsWorkspacesWorkspaceConfig_Prerequisites(rName),
		fmt.Sprintf(`
resource "aws_workspaces_workspace" "test" {
  bundle_id    = data.aws_workspaces_bundle.test.id
  directory_id = aws_workspaces_directory.test.id

  # NOTE: WorkSpaces API doesn't allow creating users in the directory.
  # However, "Administrator"" user is always present in a bare directory.
  user_name = "Administrator"

  workspace_properties {
    root_volume_size_gib = 90
    user_volume_size_gib = 50
  }

  tags = {
    Name = "tf-testacc-workspaces-workspace-%[1]s"
  }
}
`, rName))
}

func testAccWorkspacesWorkspaceConfig_validateUserVolumeSize(rName string) string {
	return composeConfig(
		testAccAwsWorkspacesWorkspaceConfig_Prerequisites(rName),
		fmt.Sprintf(`
resource "aws_workspaces_workspace" "test" {
  bundle_id    = data.aws_workspaces_bundle.test.id
  directory_id = aws_workspaces_directory.test.id

  # NOTE: WorkSpaces API doesn't allow creating users in the directory.
  # However, "Administrator"" user is always present in a bare directory.
  user_name = "Administrator"

  workspace_properties {
    root_volume_size_gib = 80
    user_volume_size_gib = 60
  }

  tags = {
    Name = "tf-testacc-workspaces-workspace-%[1]s"
  }
}
`, rName))
}

func testAccWorkspacesWorkspaceConfig_timeout(rName string) string {
	return composeConfig(
		testAccAwsWorkspacesWorkspaceConfig_Prerequisites(rName),
		fmt.Sprintf(`
resource "aws_workspaces_workspace" "test" {
  bundle_id    = data.aws_workspaces_bundle.test.id
  directory_id = aws_workspaces_directory.test.id

  # NOTE: WorkSpaces API doesn't allow creating users in the directory.
  # However, "Administrator"" user is always present in a bare directory.
  user_name = "Administrator"

  timeouts {
    create = "60m"
    update = "30m"
    delete = "30m"
  }

  tags = {
    Name = "tf-testacc-workspaces-workspace-%[1]s"
  }
}
`, rName))
}

func TestExpandWorkspaceProperties(t *testing.T) {
	cases := []struct {
		input    []interface{}
		expected *workspaces.WorkspaceProperties
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
					"compute_type_name":                         workspaces.ComputeValue,
					"root_volume_size_gib":                      80,
					"running_mode":                              workspaces.RunningModeAutoStop,
					"running_mode_auto_stop_timeout_in_minutes": 60,
					"user_volume_size_gib":                      10,
				},
			},
			expected: &workspaces.WorkspaceProperties{
				ComputeTypeName:                     aws.String(workspaces.ComputeValue),
				RootVolumeSizeGib:                   aws.Int64(80),
				RunningMode:                         aws.String(workspaces.RunningModeAutoStop),
				RunningModeAutoStopTimeoutInMinutes: aws.Int64(60),
				UserVolumeSizeGib:                   aws.Int64(10),
			},
		},
	}

	for _, c := range cases {
		actual := expandWorkspaceProperties(c.input)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Fatalf("expected\n\n%#+v\n\ngot\n\n%#+v", c.expected, actual)
		}
	}
}

func TestFlattenWorkspaceProperties(t *testing.T) {
	cases := []struct {
		input    *workspaces.WorkspaceProperties
		expected []map[string]interface{}
	}{
		// Empty
		{
			input:    nil,
			expected: []map[string]interface{}{},
		},
		// Full
		{
			input: &workspaces.WorkspaceProperties{
				ComputeTypeName:                     aws.String(workspaces.ComputeValue),
				RootVolumeSizeGib:                   aws.Int64(80),
				RunningMode:                         aws.String(workspaces.RunningModeAutoStop),
				RunningModeAutoStopTimeoutInMinutes: aws.Int64(60),
				UserVolumeSizeGib:                   aws.Int64(10),
			},
			expected: []map[string]interface{}{
				{
					"compute_type_name":                         workspaces.ComputeValue,
					"root_volume_size_gib":                      80,
					"running_mode":                              workspaces.RunningModeAutoStop,
					"running_mode_auto_stop_timeout_in_minutes": 60,
					"user_volume_size_gib":                      10,
				},
			},
		},
	}

	for _, c := range cases {
		actual := flattenWorkspaceProperties(c.input)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Fatalf("expected\n\n%#+v\n\ngot\n\n%#+v", c.expected, actual)
		}
	}
}
