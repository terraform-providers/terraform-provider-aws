package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/envvar"
)

func TestAccAWSCodePipeline_basic(t *testing.T) {
	var p1, p2 codepipeline.PipelineDeclaration
	name := acctest.RandString(10)
	resourceName := "aws_codepipeline.test"
	codestarConnectionResourceName := "aws_codestarconnections_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSCodePipelineSupported(t)
			testAccPartitionHasServicePreCheck(codestarconnections.EndpointsID, t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p1),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", "aws_iam_role.codepipeline_role", "arn"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "codepipeline", regexp.MustCompile(fmt.Sprintf("test-pipeline-%s", name))),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "stage.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "stage.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.category", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.owner", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.provider", "CodeStarSourceConnection"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.input_artifacts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.FullRepositoryId", "lifesum-terraform/test"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.BranchName", "main"),
					resource.TestCheckResourceAttrPair(resourceName, "stage.0.action.0.configuration.ConnectionArn", codestarConnectionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.run_order", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.region", ""),

					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.category", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.owner", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.provider", "CodeBuild"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.0", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.output_artifacts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.ProjectName", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.run_order", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodePipelineConfig_basicUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p2),

					resource.TestCheckResourceAttr(resourceName, "stage.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "stage.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.input_artifacts.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.output_artifacts.0", "artifacts"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.FullRepositoryId", "test-terraform/test-repo"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.BranchName", "stable"),
					resource.TestCheckResourceAttrPair(resourceName, "stage.0.action.0.configuration.ConnectionArn", codestarConnectionResourceName, "arn"),

					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.0", "artifacts"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.configuration.ProjectName", "test"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"stage.0.action.0.configuration.%",
					"stage.0.action.0.configuration.OAuthToken",
				},
			},
		},
	})
}

func TestAccAWSCodePipeline_disappears(t *testing.T) {
	var p codepipeline.PipelineDeclaration
	name := acctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSCodePipelineSupported(t)
			testAccPartitionHasServicePreCheck(codestarconnections.EndpointsID, t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCodePipeline(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCodePipeline_emptyStageArtifacts(t *testing.T) {
	var p codepipeline.PipelineDeclaration
	name := acctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSCodePipelineSupported(t)
			testAccPartitionHasServicePreCheck(codestarconnections.EndpointsID, t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineConfig_emptyStageArtifacts(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "codepipeline", regexp.MustCompile(fmt.Sprintf("test-pipeline-%s$", name))),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.category", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.owner", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.provider", "CodeBuild"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.input_artifacts.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.output_artifacts.#", "0"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodePipeline_deployWithServiceRole(t *testing.T) {
	var p codepipeline.PipelineDeclaration
	name := acctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSCodePipelineSupported(t)
			testAccPartitionHasServicePreCheck(codestarconnections.EndpointsID, t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineConfig_deployWithServiceRole(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "stage.2.name", "Deploy"),
					resource.TestCheckResourceAttr(resourceName, "stage.2.action.0.category", "Deploy"),
					resource.TestCheckResourceAttrPair(resourceName, "stage.2.action.0.role_arn", "aws_iam_role.codepipeline_action_role", "arn"),
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

func TestAccAWSCodePipeline_tags(t *testing.T) {
	var p1, p2, p3 codepipeline.PipelineDeclaration
	name := acctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSCodePipelineSupported(t)
			testAccPartitionHasServicePreCheck(codestarconnections.EndpointsID, t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineConfigWithTags(name, "tag1value", "tag2value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", fmt.Sprintf("test-pipeline-%s", name)),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "tag2value"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodePipelineConfigWithTags(name, "tag1valueUpdate", "tag2valueUpdate"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", fmt.Sprintf("test-pipeline-%s", name)),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1valueUpdate"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "tag2valueUpdate"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodePipelineConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSCodePipeline_multiregion_basic(t *testing.T) {
	var p codepipeline.PipelineDeclaration
	resourceName := "aws_codepipeline.test"
	var providers []*schema.Provider

	name := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
			testAccPreCheckAWSCodePipelineSupported(t, testAccGetAlternateRegion())
			testAccPartitionHasServicePreCheck(codestarconnections.EndpointsID, t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineConfig_multiregion(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", testAccGetRegion()),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.name", fmt.Sprintf("%s-Build", testAccGetAlternateRegion())),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.region", testAccGetAlternateRegion()),
				),
			},
			{
				Config:            testAccAWSCodePipelineConfig_multiregion(name),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodePipeline_multiregion_Update(t *testing.T) {
	var p1, p2 codepipeline.PipelineDeclaration
	resourceName := "aws_codepipeline.test"
	var providers []*schema.Provider

	name := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
			testAccPreCheckAWSCodePipelineSupported(t, testAccGetAlternateRegion())
			testAccPartitionHasServicePreCheck(codestarconnections.EndpointsID, t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineConfig_multiregion(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p1),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", testAccGetRegion()),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.name", fmt.Sprintf("%s-Build", testAccGetAlternateRegion())),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.region", testAccGetAlternateRegion()),
				),
			},
			{
				Config: testAccAWSCodePipelineConfig_multiregionUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p2),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "BuildUpdated"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", testAccGetRegion()),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.name", fmt.Sprintf("%s-BuildUpdated", testAccGetAlternateRegion())),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.region", testAccGetAlternateRegion()),
				),
			},
			{
				Config:            testAccAWSCodePipelineConfig_multiregionUpdated(name),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodePipeline_multiregion_ConvertSingleRegion(t *testing.T) {
	var p1, p2 codepipeline.PipelineDeclaration
	resourceName := "aws_codepipeline.test"
	var providers []*schema.Provider

	name := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccMultipleRegionPreCheck(t, 2)
			testAccPreCheckAWSCodePipelineSupported(t, testAccGetAlternateRegion())
			testAccPartitionHasServicePreCheck(codestarconnections.EndpointsID, t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineConfig_basic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p1),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", ""),
				),
			},
			{
				Config: testAccAWSCodePipelineConfig_multiregion(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p2),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", testAccGetRegion()),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.name", fmt.Sprintf("%s-Build", testAccGetAlternateRegion())),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.1.region", testAccGetAlternateRegion()),
				),
			},
			{
				Config: testAccAWSCodePipelineConfig_backToBasic(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p1),
					resource.TestCheckResourceAttr(resourceName, "artifact_store.#", "1"),

					resource.TestCheckResourceAttr(resourceName, "stage.1.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.name", "Build"),
					resource.TestCheckResourceAttr(resourceName, "stage.1.action.0.region", testAccGetRegion()),
				),
			},
			{
				Config:            testAccAWSCodePipelineConfig_backToBasic(name),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCodePipeline_WithNamespace(t *testing.T) {
	var p1 codepipeline.PipelineDeclaration
	name := acctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSCodePipelineSupported(t)
			testAccPartitionHasServicePreCheck(codestarconnections.EndpointsID, t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineConfigWithNamespace(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &p1),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "codepipeline", regexp.MustCompile(fmt.Sprintf("test-pipeline-%s", name))),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.namespace", "SourceVariables"),
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

func TestAccAWSCodePipeline_WithGitHubv1SourceAction(t *testing.T) {
	githubToken := envvar.TestSkipIfEmpty(t, envvar.GithubToken, "token with GitHub permissions to repository for CodePipeline source configuration")

	var v codepipeline.PipelineDeclaration
	name := acctest.RandString(10)
	resourceName := "aws_codepipeline.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSCodePipelineSupported(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodePipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodePipelineConfig_WithGitHubv1SourceAction(name, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &v),

					resource.TestCheckResourceAttr(resourceName, "stage.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.category", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.owner", "ThirdParty"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.provider", "GitHub"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.Owner", "lifesum-terraform"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.Repo", "test"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.Branch", "main"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.OAuthToken", githubToken),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"stage.0.action.0.configuration.%",
					"stage.0.action.0.configuration.OAuthToken",
				},
			},
			{
				Config: testAccAWSCodePipelineConfig_WithGitHubv1SourceAction_Updated(name, githubToken),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodePipelineExists(resourceName, &v),

					resource.TestCheckResourceAttr(resourceName, "stage.#", "2"),

					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.name", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.category", "Source"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.owner", "ThirdParty"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.provider", "GitHub"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.version", "1"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.Owner", "test-terraform"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.Repo", "test-repo"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.Branch", "stable"),
					resource.TestCheckResourceAttr(resourceName, "stage.0.action.0.configuration.OAuthToken", githubToken),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"stage.0.action.0.configuration.%",
					"stage.0.action.0.configuration.OAuthToken",
				},
			},
		},
	})
}

func testAccCheckAWSCodePipelineExists(n string, pipeline *codepipeline.PipelineDeclaration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CodePipeline ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).codepipelineconn

		out, err := conn.GetPipeline(&codepipeline.GetPipelineInput{
			Name: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*pipeline = *out.Pipeline

		return nil
	}
}

func testAccCheckAWSCodePipelineDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codepipelineconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codepipeline" {
			continue
		}

		_, err := conn.GetPipeline(&codepipeline.GetPipelineInput{
			Name: aws.String(rs.Primary.ID),
		})

		if err == nil {
			return fmt.Errorf("Expected AWS CodePipeline to be gone, but was still found")
		}
		if isAWSErr(err, "PipelineNotFoundException", "") {
			continue
		}
		return err
	}

	return nil
}

func testAccPreCheckAWSCodePipelineSupported(t *testing.T, regions ...string) {
	regions = append(regions, testAccGetRegion())
	for _, region := range regions {
		conf := &Config{
			Region: region,
		}
		client, err := conf.Client()
		if err != nil {
			t.Fatalf("error getting AWS client for region %s", region)
		}
		conn := client.(*AWSClient).codepipelineconn

		input := &codepipeline.ListPipelinesInput{}
		_, err = conn.ListPipelines(input)

		if testAccPreCheckSkipError(err) {
			t.Skipf("skipping acceptance testing: %s", err)
		}

		if err != nil {
			t.Fatalf("unexpected PreCheck error: %s", err)
		}
	}
}

func testAccAWSCodePipelineServiceIAMRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "codepipeline_role" {
  name = "codepipeline-role-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "codepipeline.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "codepipeline_policy" {
  name = "codepipeline_policy"
  role = aws_iam_role.codepipeline_role.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:GetObjectVersion",
        "s3:GetBucketVersioning"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "codebuild:BatchGetBuilds",
        "codebuild:StartBuild"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}
`, rName)
}

func testAccAWSCodePipelineServiceIAMRoleWithAssumeRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "codepipeline_role" {
  name = "codepipeline-role-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "codepipeline.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "codepipeline_policy" {
  name = "codepipeline_policy"
  role = aws_iam_role.codepipeline_role.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect":"Allow",
      "Action": [
        "s3:GetObject",
        "s3:GetObjectVersion",
        "s3:GetBucketVersioning"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "codebuild:BatchGetBuilds",
        "codebuild:StartBuild"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "sts:AssumeRole"
      ],
      "Resource": "${aws_iam_role.codepipeline_action_role.arn}"
    }
  ]
}
EOF
}
`, rName)
}

func testAccAWSCodePipelineConfig_basic(rName string) string {
	return composeConfig(
		testAccAWSCodePipelineS3DefaultBucket(rName),
		testAccAWSCodePipelineServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.test.bucket
    type     = "S3"

    encryption_key {
      id   = "1234"
      type = "KMS"
    }
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["test"]

      configuration = {
        ConnectionArn    = aws_codestarconnections_connection.test.arn
        FullRepositoryId = "lifesum-terraform/test"
        BranchName       = "main"
      }
    }
  }

  stage {
    name = "Build"

    action {
      name            = "Build"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      input_artifacts = ["test"]
      version         = "1"

      configuration = {
        ProjectName = "test"
      }
    }
  }
}

resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "GitHub"
}
`, rName))
}

func testAccAWSCodePipelineConfig_basicUpdated(rName string) string {
	return composeConfig(
		testAccAWSCodePipelineS3DefaultBucket(rName),
		testAccAWSCodePipelineS3Bucket("updated", rName),
		testAccAWSCodePipelineServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%s"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.updated.bucket
    type     = "S3"

    encryption_key {
      id   = "4567"
      type = "KMS"
    }
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["artifacts"]

      configuration = {
        ConnectionArn    = aws_codestarconnections_connection.test.arn
        FullRepositoryId = "test-terraform/test-repo"
        BranchName       = "stable"
      }
    }
  }

  stage {
    name = "Build"

    action {
      name            = "Build"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      input_artifacts = ["artifacts"]
      version         = "1"

      configuration = {
        ProjectName = "test"
      }
    }
  }
}

resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "GitHub"
}
`, rName))
}

func testAccAWSCodePipelineConfig_emptyStageArtifacts(rName string) string {
	return composeConfig(
		testAccAWSCodePipelineS3DefaultBucket(rName),
		testAccAWSCodePipelineServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.test.bucket
    type     = "S3"
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["test"]

      configuration = {
        ConnectionArn    = aws_codestarconnections_connection.test.arn
        FullRepositoryId = "lifesum-terraform/test"
        BranchName       = "main"
      }
    }
  }

  stage {
    name = "Build"

    action {
      name             = "Build"
      category         = "Build"
      owner            = "AWS"
      provider         = "CodeBuild"
      input_artifacts  = ["test", ""]
      output_artifacts = [""]
      version          = "1"

      configuration = {
        ProjectName = "test"
      }
    }
  }
}

resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "GitHub"
}
`, rName))
}

func testAccAWSCodePipelineDeployActionIAMRole(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "codepipeline_action_role" {
  name = "codepipeline-action-role-%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "codepipeline_action_policy" {
  name = "codepipeline_action_policy"
  role = aws_iam_role.codepipeline_action_role.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:GetObjectVersion",
        "s3:GetBucketVersioning"
      ],
      "Resource": [
        "${aws_s3_bucket.test.arn}",
        "${aws_s3_bucket.test.arn}/*"
      ]
    }
  ]
}
EOF
}
`, rName)
}

func testAccAWSCodePipelineConfig_deployWithServiceRole(rName string) string {
	return composeConfig(
		testAccAWSCodePipelineS3DefaultBucket(rName),
		testAccAWSCodePipelineServiceIAMRoleWithAssumeRole(rName),
		testAccAWSCodePipelineDeployActionIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%s"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.test.bucket
    type     = "S3"

    encryption_key {
      id   = "4567"
      type = "KMS"
    }
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["artifacts"]

      configuration = {
        ConnectionArn    = aws_codestarconnections_connection.test.arn
        FullRepositoryId = "lifesum-terraform/test"
        BranchName       = "main"
      }
    }
  }

  stage {
    name = "Build"

    action {
      name             = "Build"
      category         = "Build"
      owner            = "AWS"
      provider         = "CodeBuild"
      input_artifacts  = ["artifacts"]
      output_artifacts = ["artifacts2"]
      version          = "1"

      configuration = {
        ProjectName = "test"
      }
    }
  }

  stage {
    name = "Deploy"

    action {
      name            = "CreateChangeSet"
      category        = "Deploy"
      owner           = "AWS"
      provider        = "CloudFormation"
      input_artifacts = ["artifacts2"]
      role_arn        = aws_iam_role.codepipeline_action_role.arn
      version         = "1"

      configuration = {
        ActionMode    = "CHANGE_SET_REPLACE"
        ChangeSetName = "changeset"
        StackName     = "stack"
        TemplatePath  = "artifacts2::template.yaml"
      }
    }
  }
}

resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "GitHub"
}
`, rName))
}

func testAccAWSCodePipelineConfigWithTags(rName, tag1, tag2 string) string {
	return composeConfig(
		testAccAWSCodePipelineS3DefaultBucket(rName),
		testAccAWSCodePipelineServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.test.bucket
    type     = "S3"

    encryption_key {
      id   = "1234"
      type = "KMS"
    }
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["test"]

      configuration = {
        ConnectionArn    = aws_codestarconnections_connection.test.arn
        FullRepositoryId = "lifesum-terraform/test"
        BranchName       = "main"
      }
    }
  }

  stage {
    name = "Build"

    action {
      name            = "Build"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      input_artifacts = ["test"]
      version         = "1"

      configuration = {
        ProjectName = "test"
      }
    }
  }

  tags = {
    Name = "test-pipeline-%[1]s"
    tag1 = %[2]q
    tag2 = %[3]q
  }
}

resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "GitHub"
}
`, rName, tag1, tag2))
}

func testAccAWSCodePipelineConfig_multiregion(rName string) string {
	return composeConfig(
		testAccAlternateRegionProviderConfig(),
		testAccAWSCodePipelineS3DefaultBucket(rName),
		testAccAWSCodePipelineServiceIAMRole(rName),
		testAccAWSCodePipelineS3BucketWithProvider("alternate", rName, "awsalternate"),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.test.bucket
    type     = "S3"

    encryption_key {
      id   = "1234"
      type = "KMS"
    }

    region = "%[2]s"
  }

  artifact_store {
    location = aws_s3_bucket.alternate.bucket
    type     = "S3"

    encryption_key {
      id   = "5678"
      type = "KMS"
    }

    region = "%[3]s"
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["test"]

      configuration = {
        ConnectionArn    = aws_codestarconnections_connection.test.arn
        FullRepositoryId = "lifesum-terraform/test"
        BranchName       = "main"
      }
    }
  }

  stage {
    name = "Build"

    action {
      region          = "%[2]s"
      name            = "Build"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      input_artifacts = ["test"]
      version         = "1"

      configuration = {
        ProjectName = "Test"
      }
    }

    action {
      region          = "%[3]s"
      name            = "%[3]s-Build"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      input_artifacts = ["test"]
      version         = "1"

      configuration = {
        ProjectName = "%[3]s-Test"
      }
    }
  }
}

resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "GitHub"
}
`, rName, testAccGetRegion(), testAccGetAlternateRegion()))
}

func testAccAWSCodePipelineConfig_multiregionUpdated(rName string) string {
	return composeConfig(
		testAccAlternateRegionProviderConfig(),
		testAccAWSCodePipelineS3DefaultBucket(rName),
		testAccAWSCodePipelineServiceIAMRole(rName),
		testAccAWSCodePipelineS3BucketWithProvider("alternate", rName, "awsalternate"),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.test.bucket
    type     = "S3"

    encryption_key {
      id   = "4321"
      type = "KMS"
    }

    region = "%[2]s"
  }

  artifact_store {
    location = aws_s3_bucket.alternate.bucket
    type     = "S3"

    encryption_key {
      id   = "8765"
      type = "KMS"
    }

    region = "%[3]s"
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["test"]

      configuration = {
        ConnectionArn    = aws_codestarconnections_connection.test.arn
        FullRepositoryId = "lifesum-terraform/test"
        BranchName       = "main"
      }
    }
  }

  stage {
    name = "Build"

    action {
      region          = "%[2]s"
      name            = "BuildUpdated"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      input_artifacts = ["test"]
      version         = "1"

      configuration = {
        ProjectName = "Test"
      }
    }

    action {
      region          = "%[3]s"
      name            = "%[3]s-BuildUpdated"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      input_artifacts = ["test"]
      version         = "1"

      configuration = {
        ProjectName = "%[3]s-Test"
      }
    }
  }
}

resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "GitHub"
}
`, rName, testAccGetRegion(), testAccGetAlternateRegion()))
}

func testAccAWSCodePipelineConfig_backToBasic(rName string) string {
	return composeConfig(
		testAccAlternateRegionProviderConfig(),
		testAccAWSCodePipelineConfig_basic(rName),
	)
}

func testAccAWSCodePipelineS3DefaultBucket(rName string) string {
	return testAccAWSCodePipelineS3Bucket("test", rName)
}

func testAccAWSCodePipelineS3Bucket(bucket, rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "%[1]s" {
  bucket = "tf-test-pipeline-%[1]s-%[2]s"
  acl    = "private"
}
`, bucket, rName)
}

func testAccAWSCodePipelineS3BucketWithProvider(bucket, rName, provider string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "%[1]s" {
  bucket   = "tf-test-pipeline-%[1]s-%[2]s"
  acl      = "private"
  provider = %[3]s
}
`, bucket, rName, provider)
}

func testAccAWSCodePipelineConfigWithNamespace(rName string) string {
	return composeConfig(
		testAccAWSCodePipelineS3DefaultBucket(rName),
		testAccAWSCodePipelineServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.foo.bucket
    type     = "S3"

    encryption_key {
      id   = "1234"
      type = "KMS"
    }
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["test"]
      namespace        = "SourceVariables"

      configuration = {
        ConnectionArn    = aws_codestarconnections_connection.test.arn
        FullRepositoryId = "lifesum-terraform/test"
        BranchName       = "main"
      }
    }
  }

  stage {
    name = "Build"

    action {
      name            = "Build"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      input_artifacts = ["test"]
      version         = "1"

      configuration = {
        ProjectName = "test"
      }
    }
  }
}

resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "GitHub"
}

resource "aws_s3_bucket" "foo" {
  bucket = "tf-test-pipeline-%[1]s"
  acl    = "private"
}
`, rName))
}

func testAccAWSCodePipelineConfig_WithGitHubv1SourceAction(rName, githubToken string) string {
	return composeConfig(
		testAccAWSCodePipelineS3DefaultBucket(rName),
		testAccAWSCodePipelineServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.test.bucket
    type     = "S3"

    encryption_key {
      id   = "1234"
      type = "KMS"
    }
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "ThirdParty"
      provider         = "GitHub"
      version          = "1"
      output_artifacts = ["test"]

      configuration = {
        Owner      = "lifesum-terraform"
        Repo       = "test"
        Branch     = "main"
        OAuthToken = %[2]q
      }
    }
  }

  stage {
    name = "Build"

    action {
      name            = "Build"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      input_artifacts = ["test"]
      version         = "1"

      configuration = {
        ProjectName = "test"
      }
    }
  }
}
`, rName, githubToken))
}

func testAccAWSCodePipelineConfig_WithGitHubv1SourceAction_Updated(rName, githubToken string) string {
	return composeConfig(
		testAccAWSCodePipelineS3DefaultBucket(rName),
		testAccAWSCodePipelineServiceIAMRole(rName),
		fmt.Sprintf(`
resource "aws_codepipeline" "test" {
  name     = "test-pipeline-%[1]s"
  role_arn = aws_iam_role.codepipeline_role.arn

  artifact_store {
    location = aws_s3_bucket.test.bucket
    type     = "S3"

    encryption_key {
      id   = "1234"
      type = "KMS"
    }
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "ThirdParty"
      provider         = "GitHub"
      version          = "1"
      output_artifacts = ["artifacts"]

      configuration = {
        Owner      = "test-terraform"
        Repo       = "test-repo"
        Branch     = "stable"
        OAuthToken = %[2]q
      }
    }
  }

  stage {
    name = "Build"

    action {
      name            = "Build"
      category        = "Build"
      owner           = "AWS"
      provider        = "CodeBuild"
      input_artifacts = ["artifacts"]
      version         = "1"

      configuration = {
        ProjectName = "test"
      }
    }
  }
}
`, rName, githubToken))
}

func TestResourceAWSCodePipelineExpandArtifactStoresValidation(t *testing.T) {
	cases := []struct {
		Name          string
		Input         []interface{}
		ExpectedError string
	}{
		{
			Name: "Single-region",
			Input: []interface{}{
				map[string]interface{}{
					"location":       "",
					"type":           "",
					"encryption_key": []interface{}{},
					"region":         "",
				},
			},
		},
		{
			Name: "Single-region, names region",
			Input: []interface{}{
				map[string]interface{}{
					"location":       "",
					"type":           "",
					"encryption_key": []interface{}{},
					"region":         "us-west-2", //lintignore:AWSAT003
				},
			},
			ExpectedError: "region cannot be set for a single-region CodePipeline",
		},
		{
			Name: "Cross-region",
			Input: []interface{}{
				map[string]interface{}{
					"location":       "",
					"type":           "",
					"encryption_key": []interface{}{},
					"region":         "us-west-2", //lintignore:AWSAT003
				},
				map[string]interface{}{
					"location":       "",
					"type":           "",
					"encryption_key": []interface{}{},
					"region":         "us-east-1", //lintignore:AWSAT003
				},
			},
		},
		{
			Name: "Cross-region, no regions",
			Input: []interface{}{
				map[string]interface{}{
					"location":       "",
					"type":           "",
					"encryption_key": []interface{}{},
					"region":         "",
				},
				map[string]interface{}{
					"location":       "",
					"type":           "",
					"encryption_key": []interface{}{},
					"region":         "",
				},
			},
			ExpectedError: "region must be set for a cross-region CodePipeline",
		},
		{
			Name: "Cross-region, not all regions",
			Input: []interface{}{
				map[string]interface{}{
					"location":       "",
					"type":           "",
					"encryption_key": []interface{}{},
					"region":         "us-west-2", //lintignore:AWSAT003
				},
				map[string]interface{}{
					"location":       "",
					"type":           "",
					"encryption_key": []interface{}{},
					"region":         "",
				},
			},
			ExpectedError: "region must be set for a cross-region CodePipeline",
		},
		{
			Name: "Duplicate regions",
			Input: []interface{}{
				map[string]interface{}{
					"location":       "",
					"type":           "",
					"encryption_key": []interface{}{},
					"region":         "us-west-2", //lintignore:AWSAT003
				},
				map[string]interface{}{
					"location":       "",
					"type":           "",
					"encryption_key": []interface{}{},
					"region":         "us-west-2", //lintignore:AWSAT003
				},
			},
			ExpectedError: "only one Artifact Store can be defined per region for a cross-region CodePipeline",
		},
	}

	for _, tc := range cases {
		tc := tc
		_, err := expandAwsCodePipelineArtifactStores(tc.Input)
		if tc.ExpectedError == "" {
			if err != nil {
				t.Errorf("%s: Did not expect an error, but got: %w", tc.Name, err)
			}
		} else {
			if err == nil {
				t.Errorf("%s: Expected an error, but did not get one", tc.Name)
			} else {
				if err.Error() != tc.ExpectedError {
					t.Errorf("%s: Expected error %q, got %w", tc.Name, tc.ExpectedError, err)
				}
			}
		}
	}
}
