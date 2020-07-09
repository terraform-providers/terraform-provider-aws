package lister

import (
	"github.com/aws/aws-sdk-go/service/kafka"
)

func ListAllClusterPages(conn *kafka.Kafka, fn func(*kafka.ListClustersOutput, bool) bool) error {
	return conn.ListClustersPages(&kafka.ListClustersInput{}, fn)
}
