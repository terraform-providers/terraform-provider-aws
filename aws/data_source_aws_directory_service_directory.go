package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceDirectoryService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDirectoryServiceRead,
		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"directory_id": {
				Type:     schema.TypeString,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"shortname": {
				Type:     schema.TypeString,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceDirectoryServiceRead(d *schema.ResourceData, meta interface{}) error {
	dsconn := meta.(*AWSClient).dsconn
	name, nameExists := d.GetOk("name")
	id, idExists := d.GetOk("directory_id")
	req := &directoryservice.DescribeDirectoriesInput{}

	if nameExists && idExists {
		return fmt.Errorf("directory_id and name arguments can't be used together")
	}

	if !nameExists && !idExists {
		return fmt.Errorf("Either name or directory_id must be set")
	}

	resp, err := dsconn.DescribeDirectories(req)

	if err != nil {
		return fmt.Errorf("error describing directories: %s", err)
	}
	var directoryDescriptionFound *directoryservice.DirectoryDescription

	for _, directoryDescription := range resp.DirectoryDescriptions {
		directoryDescriptionName := *directoryDescription.Name
		directoryDescriptionId := *directoryDescription.DirectoryId
		// Try to match by directory_id
		if idExists && directoryDescriptionId == id.(string) {
			directoryDescriptionFound = directoryDescription
			break
			//  Try to match by name
		} else if nameExists && directoryDescriptionName == name.(string) {
			directoryDescriptionFound = directoryDescription
			break
		}
	}

	if directoryDescriptionFound == nil {
		return fmt.Errorf("no matching directory service found")
	}
	idDirectoryService := (*directoryDescriptionFound.DirectoryId)
	d.SetId(idDirectoryService)
	d.Set("name", directoryDescriptionFound.Name)
	d.Set("shortname", directoryDescriptionFound.ShortName)

	return nil
}
