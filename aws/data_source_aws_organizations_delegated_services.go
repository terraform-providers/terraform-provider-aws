package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsOrganizationsDelegatedServices() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAwsOrganizationsDelegatedServicesRead,
		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateAwsAccountId,
			},
			"delegated_services": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delegation_enabled_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service_principal": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsOrganizationsDelegatedServicesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).organizationsconn

	input := &organizations.ListDelegatedServicesForAccountInput{
		AccountId: aws.String(d.Get("account_id").(string)),
	}

	var delegators []*organizations.DelegatedService
	err := conn.ListDelegatedServicesForAccountPagesWithContext(ctx, input, func(page *organizations.ListDelegatedServicesForAccountOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		delegators = append(delegators, page.DelegatedServices...)

		return !lastPage
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error describing organizations delegated services: %w", err))
	}

	if err = d.Set("delegated_services", flattenOrganizationsDelegatedServices(delegators)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting delegated_services: %w", err))
	}

	d.SetId(meta.(*AWSClient).accountid)

	return nil
}

func flattenOrganizationsDelegatedServices(delegatedServices []*organizations.DelegatedService) []map[string]interface{} {
	if len(delegatedServices) == 0 {
		return nil
	}

	var result []map[string]interface{}
	for _, delegated := range delegatedServices {
		result = append(result, map[string]interface{}{
			"delegation_enabled_date": aws.TimeValue(delegated.DelegationEnabledDate).Format(time.RFC3339),
			"service_principal":       aws.StringValue(delegated.ServicePrincipal),
		})
	}
	return result
}
