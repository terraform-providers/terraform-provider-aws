package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
)

// DomainByName returns the Domain corresponding to the specified name.
func DomainByName(conn *elasticsearch.ElasticsearchService, name string) (*elasticsearch.ElasticsearchDomainStatus, error) {
	input := &elasticsearch.DescribeElasticsearchDomainInput{
		DomainName: aws.String(name),
	}

	output, err := conn.DescribeElasticsearchDomain(input)
	if err != nil {
		return nil, err
	}

	return output.DomainStatus, nil
}
