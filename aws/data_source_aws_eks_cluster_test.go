package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSEksClusterDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))
	dataSourceResourceName := "data.aws_eks_cluster.test"
	resourceName := "aws_eks_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEksClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEksClusterDataSourceConfig_Basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceResourceName, "arn"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "certificate_authority.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_authority.0.data", dataSourceResourceName, "certificate_authority.0.data"),
					resource.TestCheckResourceAttrPair(resourceName, "created_at", dataSourceResourceName, "created_at"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "enabled_cluster_log_types.#", "2"),
					resource.TestCheckTypeSetElemAttr(dataSourceResourceName, "enabled_cluster_log_types.*", "api"),
					resource.TestCheckTypeSetElemAttr(dataSourceResourceName, "enabled_cluster_log_types.*", "audit"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint", dataSourceResourceName, "endpoint"),
					resource.TestCheckResourceAttrPair(resourceName, "identity.#", dataSourceResourceName, "identity.#"),
					resource.TestCheckResourceAttrPair(resourceName, "identity.0.oidc.#", dataSourceResourceName, "identity.0.oidc.#"),
					resource.TestCheckResourceAttrPair(resourceName, "identity.0.oidc.0.issuer", dataSourceResourceName, "identity.0.oidc.0.issuer"),
					resource.TestCheckResourceAttrPair(resourceName, "kubernetes_network_config.#", dataSourceResourceName, "kubernetes_network_config.#"),
					resource.TestCheckResourceAttrPair(resourceName, "kubernetes_network_config.0.service_ipv4_cidr", dataSourceResourceName, "kubernetes_network_config.0.service_ipv4_cidr"),
					resource.TestMatchResourceAttr(dataSourceResourceName, "platform_version", regexp.MustCompile(`^eks\.\d+$`)),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", dataSourceResourceName, "role_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "status", dataSourceResourceName, "status"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceResourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "version", dataSourceResourceName, "version"),
					resource.TestCheckResourceAttr(dataSourceResourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.cluster_security_group_id", dataSourceResourceName, "vpc_config.0.cluster_security_group_id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.endpoint_private_access", dataSourceResourceName, "vpc_config.0.endpoint_private_access"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.endpoint_public_access", dataSourceResourceName, "vpc_config.0.endpoint_public_access"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.security_group_ids.#", dataSourceResourceName, "vpc_config.0.security_group_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.subnet_ids.#", dataSourceResourceName, "vpc_config.0.subnet_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.public_access_cidrs.#", dataSourceResourceName, "vpc_config.0.public_access_cidrs.#"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_config.0.vpc_id", dataSourceResourceName, "vpc_config.0.vpc_id"),
				),
			},
		},
	})
}

func testAccAWSEksClusterDataSourceConfig_Basic(rName string) string {
	return fmt.Sprintf(`
%[1]s

data "aws_eks_cluster" "test" {
  name = aws_eks_cluster.test.name
}
`, testAccAWSEksClusterConfig_Logging(rName, []string{"api", "audit"}))
}
