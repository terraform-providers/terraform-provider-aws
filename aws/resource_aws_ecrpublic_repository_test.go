package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecrpublic"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_ecrpublic_repository", &resource.Sweeper{
		Name: "aws_ecrpublic_repository",
		F:    testSweepEcrPublicRepositories,
	})
}

func testSweepEcrPublicRepositories(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).ecrpublicconn
	sweepResources := make([]*testSweepResource, 0)
	var errs *multierror.Error

	err = conn.DescribeRepositoriesPages(&ecrpublic.DescribeRepositoriesInput{}, func(page *ecrpublic.DescribeRepositoriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, repository := range page.Repositories {
			r := resourceAwsEcrPublicRepository()
			d := r.Data(nil)
			d.SetId(aws.StringValue(repository.RepositoryName))
			d.Set("registry_id", repository.RegistryId)
			d.Set("force_destroy", true)

			sweepResources = append(sweepResources, NewTestSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing ECR Public Repositories for %s: %w", region, err))
	}

	if err = testSweepResourceOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping ECR Public Repositories for %s: %w", region, err))
	}

	if testSweepSkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping ECR Public Repositories sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSEcrPublicRepository_basic(t *testing.T) {
	var v ecrpublic.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsEcrPublic(t) },
		ErrorCheck:   testAccErrorCheck(t, ecrpublic.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrPublicRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrPublicRepositoryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "repository_name", rName),
					testAccCheckResourceAttrAccountID(resourceName, "registry_id"),
					testAccCheckResourceAttrGlobalARN(resourceName, "arn", "ecr-public", "repository/"+rName),
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

func TestAccAWSEcrPublicRepository_catalogdata_abouttext(t *testing.T) {
	var v ecrpublic.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsEcrPublic(t) },
		ErrorCheck:   testAccErrorCheck(t, ecrpublic.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrPublicRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrPublicRepositoryCatalogDataConfigAboutText(rName, "about_text_1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.about_text", "about_text_1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEcrPublicRepositoryCatalogDataConfigAboutText(rName, "about_text_2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.about_text", "about_text_2"),
				),
			},
		},
	})
}

func TestAccAWSEcrPublicRepository_catalogdata_architectures(t *testing.T) {
	var v ecrpublic.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsEcrPublic(t) },
		ErrorCheck:   testAccErrorCheck(t, ecrpublic.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrPublicRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrPublicRepositoryCatalogDataConfigArchitectures(rName, "Linux"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.architectures.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.architectures.0", "Linux"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEcrPublicRepositoryCatalogDataConfigArchitectures(rName, "Windows"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.architectures.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.architectures.0", "Windows"),
				),
			},
		},
	})
}

func TestAccAWSEcrPublicRepository_catalogdata_description(t *testing.T) {
	var v ecrpublic.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsEcrPublic(t) },
		ErrorCheck:   testAccErrorCheck(t, ecrpublic.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrPublicRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrPublicRepositoryCatalogDataConfigDescription(rName, "description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.description", "description 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEcrPublicRepositoryCatalogDataConfigDescription(rName, "description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.description", "description 2"),
				),
			},
		},
	})
}

func TestAccAWSEcrPublicRepository_catalogdata_operatingsystems(t *testing.T) {
	var v ecrpublic.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsEcrPublic(t) },
		ErrorCheck:   testAccErrorCheck(t, ecrpublic.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrPublicRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrPublicRepositoryCatalogDataConfigOperatingSystems(rName, "ARM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.operating_systems.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.operating_systems.0", "ARM"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEcrPublicRepositoryCatalogDataConfigOperatingSystems(rName, "x86"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.operating_systems.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.operating_systems.0", "x86"),
				),
			},
		},
	})
}

func TestAccAWSEcrPublicRepository_catalogdata_usagetext(t *testing.T) {
	var v ecrpublic.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsEcrPublic(t) },
		ErrorCheck:   testAccErrorCheck(t, ecrpublic.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrPublicRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrPublicRepositoryCatalogDataConfigUsageText(rName, "usage text 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.usage_text", "usage text 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEcrPublicRepositoryCatalogDataConfigUsageText(rName, "usage text 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.usage_text", "usage text 2"),
				),
			},
		},
	})
}

func TestAccAWSEcrPublicRepository_catalogdata_logoimageblob(t *testing.T) {
	var v ecrpublic.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsEcrPublic(t) },
		ErrorCheck:   testAccErrorCheck(t, ecrpublic.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrPublicRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrPublicRepositoryCatalogDataConfigLogoImageBlob(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "catalog_data.0.logo_image_blob"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"catalog_data.0.logo_image_blob"},
			},
		},
	})
}

func TestAccAWSEcrPublicRepository_basic_forcedestroy(t *testing.T) {
	var v ecrpublic.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsEcrPublic(t) },
		ErrorCheck:   testAccErrorCheck(t, ecrpublic.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrPublicRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrPublicRepositoryConfigForceDestroy(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "repository_name", rName),
					testAccCheckResourceAttrAccountID(resourceName, "registry_id"),
					testAccCheckResourceAttrGlobalARN(resourceName, "arn", "ecr-public", "repository/"+rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccAWSEcrPublicRepository_disappears(t *testing.T) {
	var v ecrpublic.Repository
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsEcrPublic(t) },
		ErrorCheck:   testAccErrorCheck(t, ecrpublic.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEcrPublicRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEcrPublicRepositoryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEcrPublicRepositoryExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEcrPublicRepository(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSEcrPublicRepositoryExists(name string, res *ecrpublic.Repository) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ECR Public repository ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ecrpublicconn

		output, err := conn.DescribeRepositories(&ecrpublic.DescribeRepositoriesInput{
			RepositoryNames: aws.StringSlice([]string{rs.Primary.ID}),
		})
		if err != nil {
			return err
		}
		if len(output.Repositories) == 0 {
			return fmt.Errorf("ECR Public repository %s not found", rs.Primary.ID)
		}

		*res = *output.Repositories[0]

		return nil
	}
}

func testAccCheckAWSEcrPublicRepositoryDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ecrpublicconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecrpublic_repository" {
			continue
		}

		input := ecrpublic.DescribeRepositoriesInput{
			RepositoryNames: []*string{aws.String(rs.Primary.Attributes["repository_name"])},
		}

		out, err := conn.DescribeRepositories(&input)

		if isAWSErr(err, ecrpublic.ErrCodeRepositoryNotFoundException, "") {
			return nil
		}

		if err != nil {
			return err
		}

		for _, repository := range out.Repositories {
			if aws.StringValue(repository.RepositoryName) == rs.Primary.Attributes["repository_name"] {
				return fmt.Errorf("ECR Public repository still exists: %s", rs.Primary.Attributes["repository_name"])
			}
		}
	}

	return nil
}

func testAccAWSEcrPublicRepositoryConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %q
}
`, rName)
}

func testAccAWSEcrPublicRepositoryConfigForceDestroy(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %q
  force_destroy   = true
}
`, rName)
}

func testAccAWSEcrPublicRepositoryCatalogDataConfigAboutText(rName string, aboutText string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
  catalog_data {
    about_text = %[2]q
  }
}
`, rName, aboutText)
}

func testAccAWSEcrPublicRepositoryCatalogDataConfigArchitectures(rName string, architecture string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
  catalog_data {
    architectures = [%[2]q]
  }
}
`, rName, architecture)
}

func testAccAWSEcrPublicRepositoryCatalogDataConfigDescription(rName string, description string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
  catalog_data {
    description = %[2]q
  }
}
`, rName, description)
}

func testAccAWSEcrPublicRepositoryCatalogDataConfigOperatingSystems(rName string, operatingSystem string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
  catalog_data {
    operating_systems = [%[2]q]
  }
}
`, rName, operatingSystem)
}

func testAccAWSEcrPublicRepositoryCatalogDataConfigUsageText(rName string, usageText string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
  catalog_data {
    usage_text = %[2]q
  }
}
`, rName, usageText)
}

func testAccAWSEcrPublicRepositoryCatalogDataConfigLogoImageBlob(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %q
  catalog_data {
    logo_image_blob = filebase64("test-fixtures/terraform_logo.png")
  }
}
`, rName)
}

func testAccPreCheckAwsEcrPublic(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).ecrpublicconn
	input := &ecrpublic.DescribeRepositoriesInput{}
	_, err := conn.DescribeRepositories(input)
	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
