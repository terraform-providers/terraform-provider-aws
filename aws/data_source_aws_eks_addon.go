package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	tfeks "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/eks"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/eks/finder"
)

func dataSourceAwsEksAddon() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAwsEksAddonRead,
		Schema: map[string]*schema.Schema{
			"addon_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"addon_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateEKSClusterName,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modified_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_account_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsEksAddonRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).eksconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	addonName := d.Get("addon_name").(string)
	clusterName := d.Get("cluster_name").(string)
	id := tfeks.AddonCreateResourceID(clusterName, addonName)

	addon, err := finder.AddonByClusterNameAndAddonName(ctx, conn, clusterName, addonName)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading EKS Add-On (%s): %w", id, err))
	}

	d.SetId(id)
	d.Set("addon_version", addon.AddonVersion)
	d.Set("arn", addon.AddonArn)
	d.Set("created_at", aws.TimeValue(addon.CreatedAt).Format(time.RFC3339))
	d.Set("modified_at", aws.TimeValue(addon.ModifiedAt).Format(time.RFC3339))
	d.Set("service_account_role_arn", addon.ServiceAccountRoleArn)

	if err := d.Set("tags", keyvaluetags.EksKeyValueTags(addon.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	return nil
}
