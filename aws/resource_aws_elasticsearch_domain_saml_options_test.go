package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSElasticSearchDomainSAMLOptions_basic(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus

	rName := acctest.RandomWithPrefix("acc-test")
	rUserName := acctest.RandomWithPrefix("es-master-user")
	resourceName := "aws_elasticsearch_domain_saml_options.main"
	esDomainResourceName := "aws_elasticsearch_domain.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticsearch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainSAMLOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainSAMLOptionsConfig(rUserName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(esDomainResourceName, &domain),
					testAccCheckESDomainSAMLOptions(esDomainResourceName, resourceName),
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.idp.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.idp.0.entity_id", "https://terraform-dev-ed.my.salesforce.com"),
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

func TestAccAWSElasticSearchDomainSAMLOptions_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("acc-test")
	rUserName := acctest.RandomWithPrefix("es-master-user")
	resourceName := "aws_elasticsearch_domain_saml_options.main"
	esDomainResourceName := "aws_elasticsearch_domain.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticsearch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainSAMLOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainSAMLOptionsConfig(rUserName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainSAMLOptions(esDomainResourceName, resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsElasticSearchDomainSAMLOptions(), resourceName),
				),
			},
		},
	})
}

func TestAccAWSElasticSearchDomainSAMLOptions_disappears_Domain(t *testing.T) {
	rName := acctest.RandomWithPrefix("acc-test")
	rUserName := acctest.RandomWithPrefix("es-master-user")
	resourceName := "aws_elasticsearch_domain_saml_options.main"
	esDomainResourceName := "aws_elasticsearch_domain.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticsearch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainSAMLOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainSAMLOptionsConfig(rUserName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainSAMLOptions(esDomainResourceName, resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsElasticSearchDomain(), esDomainResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSElasticSearchDomainSAMLOptions_Update(t *testing.T) {
	rName := acctest.RandomWithPrefix("acc-test")
	rUserName := acctest.RandomWithPrefix("es-master-user")
	resourceName := "aws_elasticsearch_domain_saml_options.main"
	esDomainResourceName := "aws_elasticsearch_domain.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticsearch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainSAMLOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainSAMLOptionsConfig(rUserName, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.session_timeout_minutes", "60"),
					testAccCheckESDomainSAMLOptions(esDomainResourceName, resourceName),
				),
			},
			{
				Config: testAccESDomainSAMLOptionsConfigUpdate(rUserName, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.session_timeout_minutes", "180"),
					testAccCheckESDomainSAMLOptions(esDomainResourceName, resourceName),
				),
			},
		},
	})
}

func TestAccAWSElasticSearchDomainSAMLOptions_Disabled(t *testing.T) {
	rName := acctest.RandomWithPrefix("acc-test")
	rUserName := acctest.RandomWithPrefix("es-master-user")
	resourceName := "aws_elasticsearch_domain_saml_options.main"
	esDomainResourceName := "aws_elasticsearch_domain.example"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, elasticsearch.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainSAMLOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainSAMLOptionsConfig(rUserName, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.session_timeout_minutes", "60"),
					testAccCheckESDomainSAMLOptions(esDomainResourceName, resourceName),
				),
			},
			{
				Config: testAccESDomainSAMLOptionsConfigDisabled(rUserName, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "saml_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "saml_options.0.session_timeout_minutes", "0"),
					testAccCheckESDomainSAMLOptions(esDomainResourceName, resourceName),
				),
			},
		},
	})
}

func testAccCheckESDomainSAMLOptionsDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).esconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticsearch_domain_saml_options" {
			continue
		}

		resp, err := conn.DescribeElasticsearchDomain(&elasticsearch.DescribeElasticsearchDomainInput{
			DomainName: aws.String(rs.Primary.Attributes["domain_name"]),
		})

		if err == nil {
			return fmt.Errorf("Elasticsearch Domain still exists %s", resp)
		}

		awsErr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if awsErr.Code() != "ResourceNotFoundException" {
			return err
		}

	}

	return nil
}

func testAccCheckESDomainSAMLOptions(esResource string, samlOptionsResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[esResource]
		if !ok {
			return fmt.Errorf("Not found: %s", esResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		options, ok := s.RootModule().Resources[samlOptionsResource]
		if !ok {
			return fmt.Errorf("Not found: %s", samlOptionsResource)
		}

		conn := testAccProvider.Meta().(*AWSClient).esconn
		_, err := conn.DescribeElasticsearchDomain(&elasticsearch.DescribeElasticsearchDomainInput{
			DomainName: aws.String(options.Primary.Attributes["domain_name"]),
		})

		return err
	}
}

func testAccESDomainSAMLOptionsConfig(userName string, domainName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "es_master_user" {
  name = "%s"
}

resource "aws_elasticsearch_domain" "example" {
  domain_name           = "%s"
  elasticsearch_version = "7.10"

  cluster_config {
    instance_type = "r5.large.elasticsearch"
  }

  # Advanced security option must be enabled to configure SAML.
  advanced_security_options {
    enabled                        = true
    internal_user_database_enabled = false
    master_user_options {
      master_user_arn = aws_iam_user.es_master_user.arn
    }
  }

  # You must enable node-to-node encryption to use advanced security options.
  encrypt_at_rest {
    enabled = true
  }

  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
  }

  node_to_node_encryption {
    enabled = true
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}

resource "aws_elasticsearch_domain_saml_options" "main" {
  domain_name = aws_elasticsearch_domain.example.domain_name

  saml_options {
    enabled = true
    idp {
      entity_id        = "https://terraform-dev-ed.my.salesforce.com"
      metadata_content = file("./test-fixtures/saml-metadata.xml")
    }
  }
}
`, userName, domainName)
}

func testAccESDomainSAMLOptionsConfigUpdate(userName string, domainName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "es_master_user" {
  name = "%s"
}

resource "aws_elasticsearch_domain" "example" {
  domain_name           = "%s"
  elasticsearch_version = "7.10"

  cluster_config {
    instance_type = "r5.large.elasticsearch"
  }

  # Advanced security option must be enabled to configure SAML.
  advanced_security_options {
    enabled                        = true
    internal_user_database_enabled = false
    master_user_options {
      master_user_arn = aws_iam_user.es_master_user.arn
    }
  }

  # You must enable node-to-node encryption to use advanced security options.
  encrypt_at_rest {
    enabled = true
  }

  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
  }

  node_to_node_encryption {
    enabled = true
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}

resource "aws_elasticsearch_domain_saml_options" "main" {
  domain_name = aws_elasticsearch_domain.example.domain_name

  saml_options {
    enabled = true
    idp {
      entity_id        = "https://terraform-dev-ed.my.salesforce.com"
      metadata_content = file("./test-fixtures/saml-metadata.xml")
    }
    session_timeout_minutes = 180
  }
}
`, userName, domainName)
}

func testAccESDomainSAMLOptionsConfigDisabled(userName string, domainName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "es_master_user" {
  name = "%s"
}

resource "aws_elasticsearch_domain" "example" {
  domain_name           = "%s"
  elasticsearch_version = "7.10"

  cluster_config {
    instance_type = "r5.large.elasticsearch"
  }

  # Advanced security option must be enabled to configure SAML.
  advanced_security_options {
    enabled                        = true
    internal_user_database_enabled = false
    master_user_options {
      master_user_arn = aws_iam_user.es_master_user.arn
    }
  }

  # You must enable node-to-node encryption to use advanced security options.
  encrypt_at_rest {
    enabled = true
  }

  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
  }

  node_to_node_encryption {
    enabled = true
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}

resource "aws_elasticsearch_domain_saml_options" "main" {
  domain_name = aws_elasticsearch_domain.example.domain_name

  saml_options {
    enabled = false
  }
}
`, userName, domainName)
}
