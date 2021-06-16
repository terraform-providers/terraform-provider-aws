// +build !generate

package keyvaluetags

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicecatalog"
)

// Custom Service Catalog tag service update functions using the same format as generated code.

func ServiceCatalogPortfolioUpdateTags(conn *servicecatalog.ServiceCatalog, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := New(oldTagsMap)
	newTags := New(newTagsMap)

	input := &servicecatalog.UpdatePortfolioInput{
		Id: aws.String(identifier),
	}

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input.RemoveTags = aws.StringSlice(removedTags.IgnoreAws().Keys())
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input.AddTags = updatedTags.IgnoreAws().ServicecatalogTags()
	}

	_, err := conn.UpdatePortfolio(input)

	if err != nil {
		return fmt.Errorf("error updating tags for Service Catalog Product (%s): %w", identifier, err)
	}

	return nil
}

func ServiceCatalogProductUpdateTags(conn *servicecatalog.ServiceCatalog, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := New(oldTagsMap)
	newTags := New(newTagsMap)

	input := &servicecatalog.UpdateProductInput{
		Id: aws.String(identifier),
	}

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input.RemoveTags = aws.StringSlice(removedTags.IgnoreAws().Keys())
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input.AddTags = updatedTags.IgnoreAws().ServicecatalogTags()
	}

	_, err := conn.UpdateProduct(input)

	if err != nil {
		return fmt.Errorf("error updating tags for Service Catalog Product (%s): %w", identifier, err)
	}

	return nil
}
