// Code generated by kafka/generators/sweepers/main.go; DO NOT EDIT.

package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/go-multierror"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kafka/lister"
)

func testSweepKafkaClusters(region string) error {
	conn, err := sharedKafkaClientForRegion(region)
	if err != nil {
		return err
	}

	var sweeperErrs *multierror.Error

	err = lister.ListAllClusterPages(conn, func(page *kafka.ListClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, r := range page.ClusterInfoList {
			name := aws.StringValue(r.ClusterName)

			log.Printf("[INFO] Deleting Kafka Cluster: %s", name)
			err := deleteKafkaCluster(conn, deleteKafkaClusterInputFromAPIResource(r))
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error deleting Kafka Cluster (%s): %w", name, err))
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Kafka Cluster sweeper for %q: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Kafka Clusters: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
