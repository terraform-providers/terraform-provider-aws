package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsServiceQuotasServiceQuota() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsServiceQuotasServiceQuotaRead,

		Schema: map[string]*schema.Schema{
			"adjustable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_value": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"global_quota": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"quota_code": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"quota_name"},
			},
			"quota_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"quota_code"},
			},
			"service_code": {
				Type:     schema.TypeString,
				Required: true,
			},
			"service_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"value": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsServiceQuotasServiceQuotaRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).servicequotasconn

	quotaCode := d.Get("quota_code").(string)
	quotaName := d.Get("quota_name").(string)
	serviceCode := d.Get("service_code").(string)

	if quotaCode == "" && quotaName == "" {
		return fmt.Errorf("either quota_code or quota_name must be configured")
	}

	var serviceQuota *servicequotas.ServiceQuota

	if quotaCode == "" {
		input := &servicequotas.ListServiceQuotasInput{
			ServiceCode: aws.String(serviceCode),
		}

		err := conn.ListServiceQuotasPages(input, func(page *servicequotas.ListServiceQuotasOutput, lastPage bool) bool {
			for _, q := range page.Quotas {
				if aws.StringValue(q.QuotaName) == quotaName {
					serviceQuota = q
					break
				}
			}

			return !lastPage
		})

		if err != nil {
			return fmt.Errorf("error listing Service (%s) Quotas: %w", serviceCode, err)
		}

		if serviceQuota == nil {
			return fmt.Errorf("error finding Service (%s) Quota (%s): no results found", serviceCode, quotaName)
		}
	} else {
		input := &servicequotas.GetServiceQuotaInput{
			QuotaCode:   aws.String(quotaCode),
			ServiceCode: aws.String(serviceCode),
		}

		output, err := conn.GetServiceQuota(input)

		if err != nil {
			return fmt.Errorf("error getting Service (%s) Quota (%s): %w", serviceCode, quotaCode, err)
		}

		if output == nil {
			return fmt.Errorf("error getting Service (%s) Quota (%s): empty result", serviceCode, quotaCode)
		}

		serviceQuota = output.Quota
	}

	input := &servicequotas.GetAWSDefaultServiceQuotaInput{
		QuotaCode:   serviceQuota.QuotaCode,
		ServiceCode: serviceQuota.ServiceCode,
	}

	output, err := conn.GetAWSDefaultServiceQuota(input)

	if err != nil {
		return fmt.Errorf("error getting Service (%s) Default Quota (%s): %w", serviceCode, aws.StringValue(serviceQuota.QuotaCode), err)
	}

	if output == nil {
		return fmt.Errorf("error getting Service (%s) Default Quota (%s): empty result", serviceCode, aws.StringValue(serviceQuota.QuotaCode))
	}

	defaultQuota := output.Quota

	d.Set("adjustable", serviceQuota.Adjustable)
	d.Set("arn", serviceQuota.QuotaArn)
	d.Set("default_value", defaultQuota.Value)
	d.Set("global_quota", serviceQuota.GlobalQuota)
	d.Set("quota_code", serviceQuota.QuotaCode)
	d.Set("quota_name", serviceQuota.QuotaName)
	d.Set("service_code", serviceQuota.ServiceCode)
	d.Set("service_name", serviceQuota.ServiceName)
	d.Set("value", serviceQuota.Value)
	d.SetId(aws.StringValue(serviceQuota.QuotaArn))

	return nil
}
