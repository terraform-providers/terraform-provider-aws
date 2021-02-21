package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsServiceQuotasService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsServiceQuotasServiceRead,

		Schema: map[string]*schema.Schema{
			"service_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAwsServiceQuotasServiceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).servicequotasconn

	serviceName := d.Get("service_name").(string)

	input := &servicequotas.ListServicesInput{}

	var service *servicequotas.ServiceInfo
	err := conn.ListServicesPages(input, func(page *servicequotas.ListServicesOutput, lastPage bool) bool {
		for _, s := range page.Services {
			if aws.StringValue(s.ServiceName) == serviceName {
				service = s
				break
			}
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing Services: %w", err)
	}

	if service == nil {
		return fmt.Errorf("error finding Service (%s): no results found", serviceName)
	}

	d.Set("service_code", service.ServiceCode)
	d.Set("service_name", service.ServiceName)
	d.SetId(aws.StringValue(service.ServiceCode))

	return nil
}
