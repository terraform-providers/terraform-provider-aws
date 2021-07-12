package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSEksClustersDataSource_badCases(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSEksClustersDataSourceConfig_NoFilter(),
				ExpectError: regexp.MustCompile("filter parameter is required"),
			},
			{
				Config:      testAccAWSEksClustersDataSourceConfig_IncorrectFilter(),
				ExpectError: regexp.MustCompile("only filtering by tag is supported"),
			},
			{
				Config:      testAccAWSEksClustersDataSourceConfig_NoResults(),
				ExpectError: regexp.MustCompile("your query returned no results, please change your search criteria"),
			},
		},
	})
}

func TestAccAWSEksClustersDataSource_Basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	dataSourceResourceName := "data.aws_eks_clusters.test"
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksClustersDataSourceConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceResourceName, "clusters.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceResourceName, "clusters.0.arn"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_authority.0.data", dataSourceResourceName, "clusters.0.certificate_authority.0.data"),
					resource.TestCheckResourceAttrPair(resourceName, "created_at", dataSourceResourceName, "clusters.0.created_at"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "clusters.0.enabled_cluster_log_types.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint", dataSourceResourceName, "clusters.0.endpoint"),
					resource.TestCheckResourceAttrPair(resourceName, "identity.#", dataSourceResourceName, "clusters.0.identity.#"),
					resource.TestCheckResourceAttrPair(resourceName, "identity.0.oidc.#", dataSourceResourceName, "clusters.0.identity.0.oidc.#"),
					resource.TestCheckResourceAttrPair(resourceName, "identity.0.oidc.0.issuer", dataSourceResourceName, "clusters.0.identity.0.oidc.0.issuer"),
					resource.TestMatchResourceAttr(dataSourceResourceName, "platform_version", regexp.MustCompile(`^eks\.\d+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", dataSourceResourceName, "clusters.0.role_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceResourceName, "clusters.0.status"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceResourceName, "clusters.0.tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "version", dataSourceResourceName, "clusters.0.version"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "clusters.0.vpc_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.cluster_security_group_id", dataSourceResourceName, "clusters.0.vpc_config.0.cluster_security_group_id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.endpoint_private_access", dataSourceResourceName, "clusters.0.vpc_config.0.endpoint_private_access"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.endpoint_public_access", dataSourceResourceName, "clusters.0.vpc_config.0.endpoint_public_access"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.security_group_ids.#", dataSourceResourceName, "clusters.0.vpc_config.0.security_group_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.subnet_ids.#", dataSourceResourceName, "clusters.0.vpc_config.0.subnet_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.public_access_cidrs.#", dataSourceResourceName, "clusters.0.vpc_config.0.public_access_cidrs.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", dataSourceResourceName, "clusters.0.vpc_config.0.vpc_id"),
				),
			},
		},
	})
}

func TestAccAWSEksClustersDataSource_TwoTags(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	dataSourceResourceName := "data.aws_eks_clusters.test"
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksClustersDataSourceConfig_TwoTags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceResourceName, "clusters.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceResourceName, "clusters.0.arn"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_authority.0.data", dataSourceResourceName, "clusters.0.certificate_authority.0.data"),
					resource.TestCheckResourceAttrPair(resourceName, "created_at", dataSourceResourceName, "clusters.0.created_at"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "clusters.0.enabled_cluster_log_types.#", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint", dataSourceResourceName, "clusters.0.endpoint"),
					resource.TestCheckResourceAttrPair(resourceName, "identity.#", dataSourceResourceName, "clusters.0.identity.#"),
					resource.TestCheckResourceAttrPair(resourceName, "identity.0.oidc.#", dataSourceResourceName, "clusters.0.identity.0.oidc.#"),
					resource.TestCheckResourceAttrPair(resourceName, "identity.0.oidc.0.issuer", dataSourceResourceName, "clusters.0.identity.0.oidc.0.issuer"),
					resource.TestMatchResourceAttr(dataSourceResourceName, "platform_version", regexp.MustCompile(`^eks\.\d+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", dataSourceResourceName, "clusters.0.role_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceResourceName, "clusters.0.status"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceResourceName, "clusters.0.tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "version", dataSourceResourceName, "clusters.0.version"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "clusters.0.vpc_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.cluster_security_group_id", dataSourceResourceName, "clusters.0.vpc_config.0.cluster_security_group_id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.endpoint_private_access", dataSourceResourceName, "clusters.0.vpc_config.0.endpoint_private_access"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.endpoint_public_access", dataSourceResourceName, "clusters.0.vpc_config.0.endpoint_public_access"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.security_group_ids.#", dataSourceResourceName, "clusters.0.vpc_config.0.security_group_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.subnet_ids.#", dataSourceResourceName, "clusters.0.vpc_config.0.subnet_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.public_access_cidrs.#", dataSourceResourceName, "clusters.0.vpc_config.0.public_access_cidrs.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", dataSourceResourceName, "clusters.0.vpc_config.0.vpc_id"),
				),
			},
		},
	})
}

func TestAccAWSEksClustersDataSource_TwoClustersOneFilter(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-1%s", acctest.RandString(5))
	dataSourceResourceName := "data.aws_eks_clusters.test"
	resourceName := "aws_eks_cluster.test"

	rName2 := fmt.Sprintf("tf-acc-test-2%s", acctest.RandString(5))
	dataSourceResourceName2 := "data.aws_eks_clusters.test2"
	resourceName2 := "aws_eks_cluster.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksClustersDataSourceConfig_TwoClustersOneFilter(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceResourceName, "clusters.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceResourceName, "clusters.0.arn"),
					resource.TestCheckResourceAttrPair(resourceName2, "arn", dataSourceResourceName2, "clusters.1.arn"),
					resource.TestCheckResourceAttr(dataSourceResourceName2, "clusters.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName2, "arn", dataSourceResourceName, "clusters.0.arn"),
				),
			},
		},
	})
}

func testAccAWSEksClustersDataSourceConfig_NoFilter() string {
	return fmt.Sprintf(`
data "aws_eks_clusters" "test" {
}
`)
}

func testAccAWSEksClustersDataSourceConfig_IncorrectFilter() string {
	return fmt.Sprintf(`
data "aws_eks_clusters" "test" {
	filter {
		name  = "tag:tagKey"
		values = ["tagValue"]
	}

	filter {
		name  = "tagKey"
		values = ["tagValue"]
	}
}
`)
}

func testAccAWSEksClustersDataSourceConfig_NoResults() string {
	return fmt.Sprintf(`
data "aws_eks_clusters" "test" {
	filter {
		name  = "tag:tagKey"
		values = ["tagValue"]
	}
}
`)
}

func testAccAWSEksClustersDataSourceConfig_Basic(rName string) string {
	return fmt.Sprintf(`
%[1]s

data "aws_eks_clusters" "test" {
	filter {
		name   = "tag:tagKey"
		values = [aws_eks_cluster.test.tags["tagKey"]]
	}
}
`, testAccAWSEksClusterConfigTags1(rName, "tagKey", "tagValue"), rName)
}

func testAccAWSEksClustersDataSourceConfig_TwoTags(rName string) string {
	return fmt.Sprintf(`
%[1]s

data "aws_eks_clusters" "test" {
	filter {
		name   = "tag:tagKey1"
		values = [aws_eks_cluster.test.tags["tagKey1"]]
	}

	filter {
		name   = "tag:tagKey2"
		values = [aws_eks_cluster.test.tags["tagKey2"]]
	}
}
`, testAccAWSEksClusterConfigTags2(rName, "tagKey1", "tagValue1", "tagKey2", "tagValue2"), rName)
}

func testAccAWSEksClustersDataSourceConfig_TwoClustersOneFilter(rName, rName2 string) string {
	return fmt.Sprintf(`
%[1]s

data "aws_eks_clusters" "test" {
	filter {
		name   = "tag:tagKey1"
		values = [aws_eks_cluster.test.tags["tagKey1"]]
	}

	filter {
		name   = "tag:tagKey2"
		values = [aws_eks_cluster.test.tags["tagKey2"]]
	}
}

resource "aws_eks_cluster" "test2" {
	name     = %[2]q
	role_arn = "${aws_iam_role.test.arn}"
  
	tags = {
	  %[3]q = %[4]q
	}
  
	vpc_config {
	  subnet_ids = ["${aws_subnet.test.*.id[0]}", "${aws_subnet.test.*.id[1]}"]
	}
  
	depends_on = ["aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy", "aws_iam_role_policy_attachment.test-AmazonEKSServicePolicy"]
}

data "aws_eks_clusters" "test2" {
	filter {
		name   = "tag:tagKey2"
		values = [aws_eks_cluster.test2.tags["tagKey2"]]
	}
}
`, testAccAWSEksClusterConfigTags1(rName, "tagKey1", "tagValue1"), rName2, "tagKey2", "tagValue2")
}
