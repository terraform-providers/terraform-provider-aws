package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	resource.AddTestSweepers("aws_elasticsearch_domain", &resource.Sweeper{
		Name: "aws_elasticsearch_domain",
		F:    testSweepElasticSearchDomains,
	})
}

func testSweepElasticSearchDomains(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).esconn

	out, err := conn.ListDomainNames(&elasticsearch.ListDomainNamesInput{})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Elasticsearch Domain sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Elasticsearch Domains: %s", err)
	}
	for _, domain := range out.DomainNames {
		log.Printf("[INFO] Deleting Elasticsearch Domain: %s", *domain.DomainName)

		_, err := conn.DeleteElasticsearchDomain(&elasticsearch.DeleteElasticsearchDomainInput{
			DomainName: domain.DomainName,
		})
		if err != nil {
			log.Printf("[ERROR] Failed to delete Elasticsearch Domain %s: %s", *domain.DomainName, err)
			continue
		}
		err = resourceAwsElasticSearchDomainDeleteWaiter(*domain.DomainName, conn)
		if err != nil {
			log.Printf("[ERROR] Failed to wait for deletion of Elasticsearch Domain %s: %s", *domain.DomainName, err)
		}
	}

	return nil
}

func TestAccAWSElasticSearchDomain_basic(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceName := "aws_elasticsearch_domain.test"
	resourceId := fmt.Sprintf("tf-test-%d", ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(
						resourceName, "elasticsearch_version", "1.5"),
					resource.TestMatchResourceAttr(resourceName, "kibana_endpoint", regexp.MustCompile(`.*es\..*/_plugin/kibana/`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     resourceId,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_RequireHTTPS(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceId := fmt.Sprintf("tf-test-%d", ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig_DomainEndpointOptions(ri, true, "Policy-Min-TLS-1-0-2019-07"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists("aws_elasticsearch_domain.example", &domain),
					testAccCheckESDomainEndpointOptions(true, "Policy-Min-TLS-1-0-2019-07", &domain),
				),
			},
			{
				ResourceName:      "aws_elasticsearch_domain.example",
				ImportState:       true,
				ImportStateId:     resourceId,
				ImportStateVerify: true,
			},
			{
				Config: testAccESDomainConfig_DomainEndpointOptions(ri, true, "Policy-Min-TLS-1-2-2019-07"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists("aws_elasticsearch_domain.example", &domain),
					testAccCheckESDomainEndpointOptions(true, "Policy-Min-TLS-1-2-2019-07", &domain),
				),
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_ClusterConfig_ZoneAwarenessConfig(t *testing.T) {
	var domain1, domain2, domain3, domain4 elasticsearch.ElasticsearchDomainStatus
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandStringFromCharSet(16, acctest.CharSetAlphaNum)) // len = 28
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig_ClusterConfig_ZoneAwarenessConfig_AvailabilityZoneCount(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain1),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.0.availability_zone_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_enabled", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     rName,
				ImportStateVerify: true,
			},
			{
				Config: testAccESDomainConfig_ClusterConfig_ZoneAwarenessConfig_AvailabilityZoneCount(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain2),
					testAccCheckAWSESDomainNotRecreated(&domain1, &domain2),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.0.availability_zone_count", "2"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_enabled", "true"),
				),
			},
			{
				Config: testAccESDomainConfig_ClusterConfig_ZoneAwarenessEnabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain3),
					testAccCheckAWSESDomainNotRecreated(&domain2, &domain3),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.#", "0"),
				),
			},
			{
				Config: testAccESDomainConfig_ClusterConfig_ZoneAwarenessConfig_AvailabilityZoneCount(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain4),
					testAccCheckAWSESDomainNotRecreated(&domain3, &domain4),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_config.0.availability_zone_count", "3"),
					resource.TestCheckResourceAttr(resourceName, "cluster_config.0.zone_awareness_enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_withDedicatedMaster(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceName := "aws_elasticsearch_domain.test"
	resourceId := fmt.Sprintf("tf-test-%d", ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig_WithDedicatedClusterMaster(ri, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     resourceId,
				ImportStateVerify: true,
			},
			{
				Config: testAccESDomainConfig_WithDedicatedClusterMaster(ri, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
				),
			},
			{
				Config: testAccESDomainConfig_WithDedicatedClusterMaster(ri, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
				),
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_duplicate(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceId := fmt.Sprintf("tf-test-%d", ri)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		Providers: testAccProviders,
		CheckDestroy: func(s *terraform.State) error {
			conn := testAccProvider.Meta().(*AWSClient).esconn
			_, err := conn.DeleteElasticsearchDomain(&elasticsearch.DeleteElasticsearchDomainInput{
				DomainName: aws.String(resourceId),
			})
			return err
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Create duplicate
					conn := testAccProvider.Meta().(*AWSClient).esconn
					_, err := conn.CreateElasticsearchDomain(&elasticsearch.CreateElasticsearchDomainInput{
						DomainName: aws.String(resourceId),
						EBSOptions: &elasticsearch.EBSOptions{
							EBSEnabled: aws.Bool(true),
							VolumeSize: aws.Int64(10),
						},
					})
					if err != nil {
						t.Fatal(err)
					}

					err = waitForElasticSearchDomainCreation(conn, resourceId, resourceId)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testAccESDomainConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(
						resourceName, "elasticsearch_version", "1.5"),
				),
				ExpectError: regexp.MustCompile(`domain .+ already exists`),
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_v23(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceId := fmt.Sprintf("tf-test-%d", ri)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfigV23(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(
						resourceName, "elasticsearch_version", "2.3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     resourceId,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_complex(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceId := fmt.Sprintf("tf-test-%d", ri)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig_complex(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     resourceId,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_vpc(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceId := fmt.Sprintf("tf-test-%d", ri)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig_vpc(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     resourceId,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_vpc_update(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceId := fmt.Sprintf("tf-test-%d", ri)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig_vpc_update(ri, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
					testAccCheckESNumberOfSecurityGroups(1, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     resourceId,
				ImportStateVerify: true,
			},
			{
				Config: testAccESDomainConfig_vpc_update(ri, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
					testAccCheckESNumberOfSecurityGroups(2, &domain),
				),
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_internetToVpcEndpoint(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceId := fmt.Sprintf("tf-test-%d", ri)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     resourceId,
				ImportStateVerify: true,
			},
			{
				Config: testAccESDomainConfig_internetToVpcEndpoint(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
				),
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_LogPublishingOptions(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceId := fmt.Sprintf("tf-test-%d", ri)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig_LogPublishingOptions(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     resourceId,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_CognitoOptionsCreateAndRemove(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceName := "aws_elasticsearch_domain.test"
	resourceId := fmt.Sprintf("tf-test-%d", ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSCognitoIdentityProvider(t)
			testAccPreCheckIamServiceLinkedRoleEs(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig_CognitoOptions(ri, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
					testAccCheckESCognitoOptions(true, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     resourceId,
				ImportStateVerify: true,
			},
			{
				Config: testAccESDomainConfig_CognitoOptions(ri, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
					testAccCheckESCognitoOptions(false, &domain),
				),
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_CognitoOptionsUpdate(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceId := fmt.Sprintf("tf-test-%d", ri)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSCognitoIdentityProvider(t)
			testAccPreCheckIamServiceLinkedRoleEs(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig_CognitoOptions(ri, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
					testAccCheckESCognitoOptions(false, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     resourceId,
				ImportStateVerify: true,
			},
			{
				Config: testAccESDomainConfig_CognitoOptions(ri, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
					testAccCheckESCognitoOptions(true, &domain),
				),
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_policy(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	ri := acctest.RandInt()
	resourceId := fmt.Sprintf("tf-test-%d", ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfigWithPolicy(ri, ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     resourceId,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_encrypt_at_rest_default_key(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	ri := acctest.RandInt()
	resourceId := fmt.Sprintf("tf-test-%d", ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfigWithEncryptAtRestDefaultKey(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
					testAccCheckESEncrypted(true, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     resourceId,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_encrypt_at_rest_specify_key(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	ri := acctest.RandInt()
	resourceId := fmt.Sprintf("tf-test-%d", ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfigWithEncryptAtRestWithKey(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
					testAccCheckESEncrypted(true, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     resourceId,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_NodeToNodeEncryption(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	resourceName := "aws_elasticsearch_domain.test"
	ri := acctest.RandInt()
	resourceId := fmt.Sprintf("tf-test-%d", ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfigwithNodeToNodeEncryption(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
					testAccCheckESNodetoNodeEncrypted(true, &domain),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     resourceId,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_tags(t *testing.T) {
	var domain elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceId := fmt.Sprintf("tf-test-%d", ri)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSELBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     resourceId,
				ImportStateVerify: true,
			},
			{
				Config: testAccESDomainConfig_TagUpdate(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.new", "type"),
				),
			},
		},
	})
}

func TestAccAWSElasticSearchDomain_update(t *testing.T) {
	var input elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceId := fmt.Sprintf("tf-test-%d", ri)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig_ClusterUpdate(ri, 2, 22),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &input),
					testAccCheckESNumberOfInstances(2, &input),
					testAccCheckESSnapshotHour(22, &input),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     resourceId,
				ImportStateVerify: true,
			},
			{
				Config: testAccESDomainConfig_ClusterUpdate(ri, 4, 23),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &input),
					testAccCheckESNumberOfInstances(4, &input),
					testAccCheckESSnapshotHour(23, &input),
				),
			},
		}})
}

func TestAccAWSElasticSearchDomain_update_volume_type(t *testing.T) {
	var input elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceId := fmt.Sprintf("tf-test-%d", ri)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig_ClusterUpdateEBSVolume(ri, 24),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &input),
					testAccCheckESEBSVolumeEnabled(true, &input),
					testAccCheckESEBSVolumeSize(24, &input),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     resourceId,
				ImportStateVerify: true,
			},
			{
				Config: testAccESDomainConfig_ClusterUpdateInstanceStore(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &input),
					testAccCheckESEBSVolumeEnabled(false, &input),
				),
			},
			{
				Config: testAccESDomainConfig_ClusterUpdateEBSVolume(ri, 12),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &input),
					testAccCheckESEBSVolumeEnabled(true, &input),
					testAccCheckESEBSVolumeSize(12, &input),
				),
			},
		}})
}

func TestAccAWSElasticSearchDomain_update_version(t *testing.T) {
	var domain1, domain2, domain3 elasticsearch.ElasticsearchDomainStatus
	ri := acctest.RandInt()
	resourceId := fmt.Sprintf("tf-test-%d", ri)
	resourceName := "aws_elasticsearch_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckESDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccESDomainConfig_ClusterUpdateVersion(ri, "5.5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain1),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_version", "5.5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateId:     resourceId,
				ImportStateVerify: true,
			},
			{
				Config: testAccESDomainConfig_ClusterUpdateVersion(ri, "5.6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain2),
					testAccCheckAWSESDomainNotRecreated(&domain1, &domain2),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_version", "5.6"),
				),
			},
			{
				Config: testAccESDomainConfig_ClusterUpdateVersion(ri, "6.3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckESDomainExists(resourceName, &domain3),
					testAccCheckAWSESDomainNotRecreated(&domain2, &domain3),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_version", "6.3"),
				),
			},
		}})
}

func testAccCheckESDomainEndpointOptions(enforceHTTPS bool, tls string, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		options := status.DomainEndpointOptions
		if *options.EnforceHTTPS != enforceHTTPS {
			return fmt.Errorf("EnforceHTTPS differ. Given: %t, Expected: %t", *options.EnforceHTTPS, enforceHTTPS)
		}
		if *options.TLSSecurityPolicy != tls {
			return fmt.Errorf("TLSSecurityPolicy differ. Given: %s, Expected: %s", *options.TLSSecurityPolicy, tls)
		}
		return nil
	}
}

func testAccCheckESNumberOfSecurityGroups(numberOfSecurityGroups int, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		count := len(status.VPCOptions.SecurityGroupIds)
		if count != numberOfSecurityGroups {
			return fmt.Errorf("Number of security groups differ. Given: %d, Expected: %d", count, numberOfSecurityGroups)
		}
		return nil
	}
}

func testAccCheckESEBSVolumeSize(ebsVolumeSize int, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.EBSOptions
		if *conf.VolumeSize != int64(ebsVolumeSize) {
			return fmt.Errorf("EBS volume size differ. Given: %d, Expected: %d", *conf.VolumeSize, ebsVolumeSize)
		}
		return nil
	}
}
func testAccCheckESEBSVolumeEnabled(ebsEnabled bool, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.EBSOptions
		if *conf.EBSEnabled != ebsEnabled {
			return fmt.Errorf("EBS volume enabled. Given: %t, Expected: %t", *conf.EBSEnabled, ebsEnabled)
		}
		return nil
	}
}

func testAccCheckESSnapshotHour(snapshotHour int, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.SnapshotOptions
		if *conf.AutomatedSnapshotStartHour != int64(snapshotHour) {
			return fmt.Errorf("Snapshots start hour differ. Given: %d, Expected: %d", *conf.AutomatedSnapshotStartHour, snapshotHour)
		}
		return nil
	}
}

func testAccCheckESNumberOfInstances(numberOfInstances int, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.ElasticsearchClusterConfig
		if *conf.InstanceCount != int64(numberOfInstances) {
			return fmt.Errorf("Number of instances differ. Given: %d, Expected: %d", *conf.InstanceCount, numberOfInstances)
		}
		return nil
	}
}

func testAccCheckESEncrypted(encrypted bool, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.EncryptionAtRestOptions
		if *conf.Enabled != encrypted {
			return fmt.Errorf("Encrypt at rest not set properly. Given: %t, Expected: %t", *conf.Enabled, encrypted)
		}
		return nil
	}
}

func testAccCheckESNodetoNodeEncrypted(encrypted bool, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		options := status.NodeToNodeEncryptionOptions
		if aws.BoolValue(options.Enabled) != encrypted {
			return fmt.Errorf("Node-to-Node Encryption not set properly. Given: %t, Expected: %t", aws.BoolValue(options.Enabled), encrypted)
		}
		return nil
	}
}

func testAccCheckESCognitoOptions(enabled bool, status *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conf := status.CognitoOptions
		if *conf.Enabled != enabled {
			return fmt.Errorf("CognitoOptions not set properly. Given: %t, Expected: %t", *conf.Enabled, enabled)
		}
		return nil
	}
}

func testAccCheckESDomainExists(n string, domain *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ES Domain ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).esconn
		opts := &elasticsearch.DescribeElasticsearchDomainInput{
			DomainName: aws.String(rs.Primary.Attributes["domain_name"]),
		}

		resp, err := conn.DescribeElasticsearchDomain(opts)
		if err != nil {
			return fmt.Errorf("Error describing domain: %s", err.Error())
		}

		*domain = *resp.DomainStatus

		return nil
	}
}

func testAccCheckAWSESDomainNotRecreated(i, j *elasticsearch.ElasticsearchDomainStatus) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).esconn

		iConfig, err := conn.DescribeElasticsearchDomainConfig(&elasticsearch.DescribeElasticsearchDomainConfigInput{
			DomainName: i.DomainName,
		})
		if err != nil {
			return err
		}
		jConfig, err := conn.DescribeElasticsearchDomainConfig(&elasticsearch.DescribeElasticsearchDomainConfigInput{
			DomainName: j.DomainName,
		})
		if err != nil {
			return err
		}

		if aws.TimeValue(iConfig.DomainConfig.ElasticsearchClusterConfig.Status.CreationDate) != aws.TimeValue(jConfig.DomainConfig.ElasticsearchClusterConfig.Status.CreationDate) {
			return fmt.Errorf("ES Domain was recreated")
		}

		return nil
	}
}

func testAccCheckESDomainDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticsearch_domain" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).esconn
		opts := &elasticsearch.DescribeElasticsearchDomainInput{
			DomainName: aws.String(rs.Primary.Attributes["domain_name"]),
		}

		_, err := conn.DescribeElasticsearchDomain(opts)
		// Verify the error is what we want
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ResourceNotFoundException" {
				continue
			}
			return err
		}
	}
	return nil
}

func testAccPreCheckIamServiceLinkedRoleEs(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).iamconn
	dnsSuffix := testAccProvider.Meta().(*AWSClient).dnsSuffix

	input := &iam.ListRolesInput{
		PathPrefix: aws.String("/aws-service-role/es."),
	}

	var role *iam.Role
	err := conn.ListRolesPages(input, func(page *iam.ListRolesOutput, lastPage bool) bool {
		for _, r := range page.Roles {
			if strings.HasPrefix(aws.StringValue(r.Path), "/aws-service-role/es.") {
				role = r
			}
		}

		return !lastPage
	})

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if role == nil {
		t.Fatalf("missing IAM Service Linked Role (es.%s), please create it in the AWS account and retry", dnsSuffix)
	}
}

func testAccESDomainConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = "tf-test-%d"

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, randInt)
}

func testAccESDomainConfig_DomainEndpointOptions(randInt int, enforceHttps bool, tlsSecurityPolicy string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "example" {
  domain_name = "tf-test-%[1]d"

  domain_endpoint_options {
    enforce_https       = %[2]t
    tls_security_policy = %[3]q
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, randInt, enforceHttps, tlsSecurityPolicy)
}

func testAccESDomainConfig_ClusterConfig_ZoneAwarenessConfig_AvailabilityZoneCount(rName string, availabilityZoneCount int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

  cluster_config {
    instance_type          = "t2.small.elasticsearch"
    instance_count         = 6
    zone_awareness_enabled = true

    zone_awareness_config {
      availability_zone_count = %[2]d
    }
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, rName, availabilityZoneCount)
}

func testAccESDomainConfig_ClusterConfig_ZoneAwarenessEnabled(rName string, zoneAwarenessEnabled bool) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = %[1]q

  cluster_config {
    instance_type          = "t2.small.elasticsearch"
    instance_count         = 6
    zone_awareness_enabled = %[2]t
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, rName, zoneAwarenessEnabled)
}

func testAccESDomainConfig_WithDedicatedClusterMaster(randInt int, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = "tf-test-%d"

  cluster_config {
    instance_type            = "t2.small.elasticsearch"
    instance_count           = "1"
    dedicated_master_enabled = %t
    dedicated_master_count   = "3"
    dedicated_master_type    = "t2.small.elasticsearch"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}
`, randInt, enabled)
}

func testAccESDomainConfig_ClusterUpdate(randInt, instanceInt, snapshotInt int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = "tf-test-%d"

  advanced_options = {
    "indices.fielddata.cache.size" = 80
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = %d
    zone_awareness_enabled = true
    instance_type          = "t2.small.elasticsearch"
  }

  snapshot_options {
    automated_snapshot_start_hour = %d
  }
}
`, randInt, instanceInt, snapshotInt)
}

func testAccESDomainConfig_ClusterUpdateEBSVolume(randInt, volumeSize int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = "tf-test-%d"

  elasticsearch_version = "6.0"

  advanced_options = {
    "indices.fielddata.cache.size" = 80
  }

  ebs_options {
    ebs_enabled = true
    volume_size = %d
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.small.elasticsearch"
  }
}
`, randInt, volumeSize)
}

func testAccESDomainConfig_ClusterUpdateVersion(randInt int, version string) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = "tf-test-%d"

  elasticsearch_version = "%v"

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 1
    zone_awareness_enabled = false
    instance_type          = "t2.small.elasticsearch"
  }
}
`, randInt, version)
}

func testAccESDomainConfig_ClusterUpdateInstanceStore(randInt int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = "tf-test-%d"

  elasticsearch_version = "6.0"

  advanced_options = {
    "indices.fielddata.cache.size" = 80
  }

  ebs_options {
    ebs_enabled = false
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "i3.large.elasticsearch"
  }
}
`, randInt)
}

func testAccESDomainConfig_TagUpdate(randInt int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = "tf-test-%d"

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  tags = {
    foo = "bar"
    new = "type"
  }
}
`, randInt)
}

func testAccESDomainConfigWithPolicy(randESId int, randRoleId int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_elasticsearch_domain" "test" {
  domain_name = "tf-test-%d"

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  access_policies = <<CONFIG
  {
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
	"AWS": "${aws_iam_role.example_role.arn}"
      },
      "Action": "es:*",
      "Resource": "arn:${data.aws_partition.current.partition}:es:*"
    }
  ]
  }
CONFIG
}

resource "aws_iam_role" "example_role" {
  name               = "es-domain-role-%d"
  assume_role_policy = "${data.aws_iam_policy_document.instance-assume-role-policy.json}"
}

data "aws_iam_policy_document" "instance-assume-role-policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.${data.aws_partition.current.dns_suffix}"]
    }
  }
}
`, randESId, randRoleId)
}

func testAccESDomainConfigWithEncryptAtRestDefaultKey(randESId int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = "tf-test-%d"

  elasticsearch_version = "6.0"

  # Encrypt at rest requires m4/c4/r4/i2 instances. See http://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/aes-supported-instance-types.html
  cluster_config {
    instance_type = "m4.large.elasticsearch"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  encrypt_at_rest {
    enabled = true
  }
}
`, randESId)
}

func testAccESDomainConfigWithEncryptAtRestWithKey(randESId int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "es" {
  description             = "kms-key-for-tf-test-%d"
  deletion_window_in_days = 7
}

resource "aws_elasticsearch_domain" "test" {
  domain_name = "tf-test-%d"

  elasticsearch_version = "6.0"

  # Encrypt at rest requires m4/c4/r4/i2 instances. See http://docs.aws.amazon.com/elasticsearch-service/latest/developerguide/aes-supported-instance-types.html
  cluster_config {
    instance_type = "m4.large.elasticsearch"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  encrypt_at_rest {
    enabled    = true
    kms_key_id = "${aws_kms_key.es.key_id}"
  }
}
`, randESId, randESId)
}

func testAccESDomainConfigwithNodeToNodeEncryption(randInt int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = "tf-test-%d"

  elasticsearch_version = "6.0"

  cluster_config {
    instance_type = "m4.large.elasticsearch"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  node_to_node_encryption {
    enabled = true
  }
}
`, randInt)
}

func testAccESDomainConfig_complex(randInt int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = "tf-test-%d"

  advanced_options = {
    "indices.fielddata.cache.size" = 80
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.small.elasticsearch"
  }

  snapshot_options {
    automated_snapshot_start_hour = 23
  }

  tags = {
    bar = "complex"
  }
}
`, randInt)
}

func testAccESDomainConfigV23(randInt int) string {
	return fmt.Sprintf(`
resource "aws_elasticsearch_domain" "test" {
  domain_name = "tf-test-%d"

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  elasticsearch_version = "2.3"
}
`, randInt)
}

func testAccESDomainConfig_vpc(randInt int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "elasticsearch_in_vpc" {
  cidr_block = "192.168.0.0/22"

  tags = {
    Name = "terraform-testacc-elasticsearch-domain-in-vpc"
  }
}

resource "aws_subnet" "first" {
  vpc_id            = "${aws_vpc.elasticsearch_in_vpc.id}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  cidr_block        = "192.168.0.0/24"

  tags = {
    Name = "tf-acc-elasticsearch-domain-in-vpc-first"
  }
}

resource "aws_subnet" "second" {
  vpc_id            = "${aws_vpc.elasticsearch_in_vpc.id}"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
  cidr_block        = "192.168.1.0/24"

  tags = {
    Name = "tf-acc-elasticsearch-domain-in-vpc-second"
  }
}

resource "aws_security_group" "first" {
  vpc_id = "${aws_vpc.elasticsearch_in_vpc.id}"
}

resource "aws_security_group" "second" {
  vpc_id = "${aws_vpc.elasticsearch_in_vpc.id}"
}

resource "aws_elasticsearch_domain" "test" {
  domain_name = "tf-test-%d"

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.small.elasticsearch"
  }

  vpc_options {
    security_group_ids = ["${aws_security_group.first.id}", "${aws_security_group.second.id}"]
    subnet_ids         = ["${aws_subnet.first.id}", "${aws_subnet.second.id}"]
  }
}
`, randInt)
}

func testAccESDomainConfig_vpc_update(randInt int, update bool) string {
	var sg_ids, subnet_string string
	if update {
		sg_ids = "${aws_security_group.first.id}\", \"${aws_security_group.second.id}"
		subnet_string = "second"
	} else {
		sg_ids = "${aws_security_group.first.id}"
		subnet_string = "first"
	}

	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "elasticsearch_in_vpc" {
  cidr_block = "192.168.0.0/22"

  tags = {
    Name = "terraform-testacc-elasticsearch-domain-in-vpc-update"
  }
}

resource "aws_subnet" "az1_first" {
  vpc_id            = "${aws_vpc.elasticsearch_in_vpc.id}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  cidr_block        = "192.168.0.0/24"

  tags = {
    Name = "tf-acc-elasticsearch-domain-in-vpc-update-az1-first"
  }
}

resource "aws_subnet" "az2_first" {
  vpc_id            = "${aws_vpc.elasticsearch_in_vpc.id}"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
  cidr_block        = "192.168.1.0/24"

  tags = {
    Name = "tf-acc-elasticsearch-domain-in-vpc-update-az2-first"
  }
}

resource "aws_subnet" "az1_second" {
  vpc_id            = "${aws_vpc.elasticsearch_in_vpc.id}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  cidr_block        = "192.168.2.0/24"

  tags = {
    Name = "tf-acc-elasticsearch-domain-in-vpc-update-az1-second"
  }
}

resource "aws_subnet" "az2_second" {
  vpc_id            = "${aws_vpc.elasticsearch_in_vpc.id}"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
  cidr_block        = "192.168.3.0/24"

  tags = {
    Name = "tf-acc-elasticsearch-domain-in-vpc-update-az2-second"
  }
}

resource "aws_security_group" "first" {
  vpc_id = "${aws_vpc.elasticsearch_in_vpc.id}"
}

resource "aws_security_group" "second" {
  vpc_id = "${aws_vpc.elasticsearch_in_vpc.id}"
}

resource "aws_elasticsearch_domain" "test" {
  domain_name = "tf-test-%d"

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.small.elasticsearch"
  }

  vpc_options {
    security_group_ids = ["%s"]
    subnet_ids         = ["${aws_subnet.az1_%s.id}", "${aws_subnet.az2_%s.id}"]
  }
}
`, randInt, sg_ids, subnet_string, subnet_string)
}

func testAccESDomainConfig_internetToVpcEndpoint(randInt int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "elasticsearch_in_vpc" {
  cidr_block = "192.168.0.0/22"

  tags = {
    Name = "terraform-testacc-elasticsearch-domain-internet-to-vpc-endpoint"
  }
}

resource "aws_subnet" "first" {
  vpc_id            = "${aws_vpc.elasticsearch_in_vpc.id}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  cidr_block        = "192.168.0.0/24"

  tags = {
    Name = "tf-acc-elasticsearch-domain-internet-to-vpc-endpoint-first"
  }
}

resource "aws_subnet" "second" {
  vpc_id            = "${aws_vpc.elasticsearch_in_vpc.id}"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
  cidr_block        = "192.168.1.0/24"

  tags = {
    Name = "tf-acc-elasticsearch-domain-internet-to-vpc-endpoint-second"
  }
}

resource "aws_security_group" "first" {
  vpc_id = "${aws_vpc.elasticsearch_in_vpc.id}"
}

resource "aws_security_group" "second" {
  vpc_id = "${aws_vpc.elasticsearch_in_vpc.id}"
}

resource "aws_elasticsearch_domain" "test" {
  domain_name = "tf-test-%d"

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.small.elasticsearch"
  }

  vpc_options {
    security_group_ids = ["${aws_security_group.first.id}", "${aws_security_group.second.id}"]
    subnet_ids         = ["${aws_subnet.first.id}", "${aws_subnet.second.id}"]
  }
}
`, randInt)
}

func testAccESDomainConfig_LogPublishingOptions(randInt int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cloudwatch_log_group" "test" {
  name = "tf-test-%d"
}

resource "aws_cloudwatch_log_resource_policy" "example" {
  policy_name = "tf-cwlp-%d"

  policy_document = <<CONFIG
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "es.${data.aws_partition.current.dns_suffix}"
      },
      "Action": [
        "logs:PutLogEvents",
        "logs:PutLogEventsBatch",
        "logs:CreateLogStream"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:logs:*"
    }
  ]
}
CONFIG
}

resource "aws_elasticsearch_domain" "test" {
  domain_name = "tf-test-%d"

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  log_publishing_options {
    log_type                 = "INDEX_SLOW_LOGS"
    cloudwatch_log_group_arn = "${aws_cloudwatch_log_group.test.arn}"
  }
}
`, randInt, randInt, randInt)
}

func testAccESDomainConfig_CognitoOptions(randInt int, includeCognitoOptions bool) string {

	var cognitoOptions string
	if includeCognitoOptions {
		cognitoOptions = `
		cognito_options {
			enabled = true
			user_pool_id = "${aws_cognito_user_pool.example.id}"
			identity_pool_id = "${aws_cognito_identity_pool.example.id}"
			role_arn = "${aws_iam_role.example.arn}"
		}`
	} else {
		cognitoOptions = ""
	}

	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_cognito_user_pool" "example" {
  name = "tf-test-%d"
}

resource "aws_cognito_user_pool_domain" "example" {
  domain = "tf-test-%d"
	user_pool_id = "${aws_cognito_user_pool.example.id}"
}

resource "aws_cognito_identity_pool" "example" {
  identity_pool_name = "tf_test_%d"
	allow_unauthenticated_identities = false

  lifecycle {
    ignore_changes = ["cognito_identity_providers"]
  }
}

resource "aws_iam_role" "example" {
	name = "tf-test-%d"
	path = "/service-role/"
	assume_role_policy = "${data.aws_iam_policy_document.assume-role-policy.json}"
}

data "aws_iam_policy_document" "assume-role-policy" {
  statement {
    sid     = ""
		actions = ["sts:AssumeRole"]
		effect  = "Allow"

    principals {
      type        = "Service"
      identifiers = ["es.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role_policy_attachment" "example" {
	role       = "${aws_iam_role.example.name}"
	policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonESCognitoAccess"
}

resource "aws_elasticsearch_domain" "test" {
	domain_name = "tf-test-%d"

	elasticsearch_version = "6.0"

	%s

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  depends_on = [
		"aws_iam_role.example",
		"aws_iam_role_policy_attachment.example"
	]
}
`, randInt, randInt, randInt, randInt, randInt, cognitoOptions)
}
