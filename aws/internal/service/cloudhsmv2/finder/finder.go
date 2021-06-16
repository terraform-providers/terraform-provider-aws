package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
)

func Cluster(conn *cloudhsmv2.CloudHSMV2, id string) (*cloudhsmv2.Cluster, error) {
	input := &cloudhsmv2.DescribeClustersInput{
		Filters: map[string][]*string{
			"clusterIds": aws.StringSlice([]string{id}),
		},
	}

	var result *cloudhsmv2.Cluster

	err := conn.DescribeClustersPages(input, func(page *cloudhsmv2.DescribeClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cluster := range page.Clusters {
			if cluster == nil {
				continue
			}

			if aws.StringValue(cluster.ClusterId) == id {
				result = cluster
				break
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func Hsm(conn *cloudhsmv2.CloudHSMV2, hsmID string, eniID string) (*cloudhsmv2.Hsm, error) {
	input := &cloudhsmv2.DescribeClustersInput{}

	var result *cloudhsmv2.Hsm

	err := conn.DescribeClustersPages(input, func(page *cloudhsmv2.DescribeClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cluster := range page.Clusters {
			if cluster == nil {
				continue
			}

			for _, hsm := range cluster.Hsms {
				if hsm == nil {
					continue
				}

				// CloudHSMv2 HSM instances can be recreated, but the ENI ID will
				// remain consistent. Without this ENI matching, HSM instances
				// instances can become orphaned.
				if aws.StringValue(hsm.HsmId) == hsmID || aws.StringValue(hsm.EniId) == eniID {
					result = hsm
					return false
				}
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
