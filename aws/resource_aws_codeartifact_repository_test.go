package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_codeartifact_repository", &resource.Sweeper{
		Name: "aws_codeartifact_repository",
		F:    testSweepCodeArtifactRepositories,
	})
}

func testSweepCodeArtifactRepositories(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).codeartifactconn
	input := &codeartifact.ListRepositoriesInput{}
	var sweeperErrs *multierror.Error

	err = conn.ListRepositoriesPages(input, func(page *codeartifact.ListRepositoriesOutput, lastPage bool) bool {
		for _, repositoryPtr := range page.Repositories {
			if repositoryPtr == nil {
				continue
			}

			repository := aws.StringValue(repositoryPtr.Name)
			input := &codeartifact.DeleteRepositoryInput{
				Repository:  repositoryPtr.Name,
				Domain:      repositoryPtr.DomainName,
				DomainOwner: repositoryPtr.DomainOwner,
			}

			log.Printf("[INFO] Deleting CodeArtifact Repository: %s", repository)

			_, err := conn.DeleteRepository(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting CodeArtifact Repository (%s): %w", repository, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping CodeArtifact Repository sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CodeArtifact Repositories: %w", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSCodeArtifactRepository_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codeartifact.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeArtifactRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeArtifactRepositoryBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactRepositoryExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "codeartifact", fmt.Sprintf("repository/%s/%s", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "repository", rName),
					resource.TestCheckResourceAttr(resourceName, "domain", rName),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", "owner"),
					resource.TestCheckResourceAttrPair(resourceName, "administrator_account", "aws_codeartifact_domain.test", "owner"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "upstream.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.#", "0"),
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

func TestAccAWSCodeArtifactRepository_tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck("codeartifact", t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeArtifactRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeArtifactRepositoryConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactRepositoryExists(resourceName),
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
				Config: testAccAWSCodeArtifactRepositoryConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSCodeArtifactRepositoryConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSCodeArtifactRepository_owner(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codeartifact.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeArtifactRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeArtifactRepositoryOwnerConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactRepositoryExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "codeartifact", fmt.Sprintf("repository/%s/%s", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "repository", rName),
					resource.TestCheckResourceAttr(resourceName, "domain", rName),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", "owner"),
					resource.TestCheckResourceAttrPair(resourceName, "administrator_account", "aws_codeartifact_domain.test", "owner"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "upstream.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.#", "0"),
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

func TestAccAWSCodeArtifactRepository_description(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codeartifact.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeArtifactRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeArtifactRepositoryDescConfig(rName, "desc"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "desc"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeArtifactRepositoryDescConfig(rName, "desc2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "desc2"),
				),
			},
		},
	})
}

func TestAccAWSCodeArtifactRepository_upstreams(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codeartifact.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeArtifactRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeArtifactRepositoryUpstreamsConfig1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "upstream.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "upstream.0.repository_name", fmt.Sprintf("%s-upstream1", rName)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeArtifactRepositoryUpstreamsConfig2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "upstream.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "upstream.0.repository_name", fmt.Sprintf("%s-upstream1", rName)),
					resource.TestCheckResourceAttr(resourceName, "upstream.1.repository_name", fmt.Sprintf("%s-upstream2", rName)),
				),
			},
			{
				Config: testAccAWSCodeArtifactRepositoryUpstreamsConfig1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "upstream.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "upstream.0.repository_name", fmt.Sprintf("%s-upstream1", rName)),
				),
			},
		},
	})
}

func TestAccAWSCodeArtifactRepository_externalConnection(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codeartifact.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeArtifactRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeArtifactRepositoryExternalConnectionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "external_connections.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.0.external_connection_name", "public:npmjs"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.0.package_format", "npm"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.0.status", "AVAILABLE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCodeArtifactRepositoryBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "external_connections.#", "0"),
				),
			},
			{
				Config: testAccAWSCodeArtifactRepositoryExternalConnectionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactRepositoryExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "external_connections.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.0.external_connection_name", "public:npmjs"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.0.package_format", "npm"),
					resource.TestCheckResourceAttr(resourceName, "external_connections.0.status", "AVAILABLE"),
				),
			},
		},
	})
}
func TestAccAWSCodeArtifactRepository_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_codeartifact_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codeartifact.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCodeArtifactRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCodeArtifactRepositoryBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCodeArtifactRepositoryExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCodeArtifactRepository(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSCodeArtifactRepositoryExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no CodeArtifact repository set")
		}

		conn := testAccProvider.Meta().(*AWSClient).codeartifactconn
		owner, domain, repo, err := decodeCodeArtifactRepositoryID(rs.Primary.ID)
		if err != nil {
			return err
		}
		_, err = conn.DescribeRepository(&codeartifact.DescribeRepositoryInput{
			Repository:  aws.String(repo),
			Domain:      aws.String(domain),
			DomainOwner: aws.String(owner),
		})

		return err
	}
}

func testAccCheckAWSCodeArtifactRepositoryDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codeartifact_repository" {
			continue
		}

		owner, domain, repo, err := decodeCodeArtifactRepositoryID(rs.Primary.ID)
		if err != nil {
			return err
		}
		conn := testAccProvider.Meta().(*AWSClient).codeartifactconn
		resp, err := conn.DescribeRepository(&codeartifact.DescribeRepositoryInput{
			Repository:  aws.String(repo),
			Domain:      aws.String(domain),
			DomainOwner: aws.String(owner),
		})

		if err == nil {
			if aws.StringValue(resp.Repository.Name) == repo &&
				aws.StringValue(resp.Repository.DomainName) == domain &&
				aws.StringValue(resp.Repository.DomainOwner) == owner {
				return fmt.Errorf("CodeArtifact Repository %s in Domain %s still exists", repo, domain)
			}
		}

		if isAWSErr(err, codeartifact.ErrCodeResourceNotFoundException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccAWSCodeArtifactRepositoryBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}
`, rName)
}

func testAccAWSCodeArtifactRepositoryBasicConfig(rName string) string {
	return testAccAWSCodeArtifactRepositoryBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain
}
`, rName)
}

func testAccAWSCodeArtifactRepositoryOwnerConfig(rName string) string {
	return testAccAWSCodeArtifactRepositoryBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "test" {
  repository   = %[1]q
  domain       = aws_codeartifact_domain.test.domain
  domain_owner = aws_codeartifact_domain.test.owner
}
`, rName)
}

func testAccAWSCodeArtifactRepositoryDescConfig(rName, desc string) string {
	return testAccAWSCodeArtifactRepositoryBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "test" {
  repository  = %[1]q
  domain      = aws_codeartifact_domain.test.domain
  description = %[2]q
}
`, rName, desc)
}

func testAccAWSCodeArtifactRepositoryUpstreamsConfig1(rName string) string {
	return testAccAWSCodeArtifactRepositoryBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "upstream1" {
  repository = "%[1]s-upstream1"
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain

  upstream {
    repository_name = aws_codeartifact_repository.upstream1.repository
  }
}
`, rName)
}

func testAccAWSCodeArtifactRepositoryUpstreamsConfig2(rName string) string {
	return testAccAWSCodeArtifactRepositoryBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "upstream1" {
  repository = "%[1]s-upstream1"
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_repository" "upstream2" {
  repository = "%[1]s-upstream2"
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain

  upstream {
    repository_name = aws_codeartifact_repository.upstream1.repository
  }

  upstream {
    repository_name = aws_codeartifact_repository.upstream2.repository
  }
}
`, rName)
}

func testAccAWSCodeArtifactRepositoryExternalConnectionConfig(rName string) string {
	return testAccAWSCodeArtifactRepositoryBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain

  external_connections {
    external_connection_name = "public:npmjs"
  }
}
`, rName)
}

func testAccAWSCodeArtifactRepositoryConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSCodeArtifactRepositoryBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSCodeArtifactRepositoryConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSCodeArtifactRepositoryBaseConfig(rName) + fmt.Sprintf(`
resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
