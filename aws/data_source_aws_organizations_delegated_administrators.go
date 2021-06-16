package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsOrganizationsDelegatedAdministrators() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAwsOrganizationsDelegatedAdministratorsRead,
		Schema: map[string]*schema.Schema{
			"service_principal": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"delegated_administrators": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"delegation_enabled_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"email": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"joined_method": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"joined_timestamp": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAwsOrganizationsDelegatedAdministratorsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).organizationsconn

	input := &organizations.ListDelegatedAdministratorsInput{}

	if v, ok := d.GetOk("service_principal"); ok {
		input.ServicePrincipal = aws.String(v.(string))
	}

	var delegators []*organizations.DelegatedAdministrator

	err := conn.ListDelegatedAdministratorsPagesWithContext(ctx, input, func(page *organizations.ListDelegatedAdministratorsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		delegators = append(delegators, page.DelegatedAdministrators...)

		return !lastPage
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("error describing organizations delegated Administrators: %w", err))
	}

	if err = d.Set("delegated_administrators", flattenOrganizationsDelegatedAdministrators(delegators)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting delegated_administrators: %w", err))
	}

	d.SetId(meta.(*AWSClient).accountid)

	return nil
}

func flattenOrganizationsDelegatedAdministrators(delegatedAdministrators []*organizations.DelegatedAdministrator) []map[string]interface{} {
	if len(delegatedAdministrators) == 0 {
		return nil
	}

	var result []map[string]interface{}
	for _, delegated := range delegatedAdministrators {
		result = append(result, map[string]interface{}{
			"arn":                     aws.StringValue(delegated.Arn),
			"delegation_enabled_date": aws.TimeValue(delegated.DelegationEnabledDate).Format(time.RFC3339),
			"email":                   aws.StringValue(delegated.Email),
			"id":                      aws.StringValue(delegated.Id),
			"joined_method":           aws.StringValue(delegated.JoinedMethod),
			"joined_timestamp":        aws.TimeValue(delegated.JoinedTimestamp).Format(time.RFC3339),
			"name":                    aws.StringValue(delegated.Name),
			"status":                  aws.StringValue(delegated.Status),
		})
	}
	return result
}
