package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/eks/waiter"
)

func resourceAwsEksAddon() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAwsEksAddonCreate,
		ReadWithoutTimeout:   resourceAwsEksAddonRead,
		UpdateWithoutTimeout: resourceAwsEksAddonUpdate,
		DeleteWithoutTimeout: resourceAwsEksAddonDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"addon_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateEKSClusterName,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"addon_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					// Regular expression taken from: https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
					validation.StringMatch(regexp.MustCompile(`^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`), "must follow semantic version format"),
				),
			},
			"service_account_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateArn,
			},
			"resolve_conflicts": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(eks.ResolveConflicts_Values(), false),
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modified_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
		},
	}
}

func resourceAwsEksAddonCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).eksconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	clusterName := d.Get("cluster_name").(string)
	addonName := d.Get("addon_name").(string)

	input := &eks.CreateAddonInput{
		ClusterName:        aws.String(clusterName),
		AddonName:          aws.String(addonName),
		ClientRequestToken: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("resolve_conflicts"); ok {
		input.ResolveConflicts = aws.String(v.(string))
	}

	if v, ok := d.GetOk("addon_version"); ok {
		input.AddonVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("service_account_role_arn"); ok {
		input.ServiceAccountRoleArn = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().EksTags()
	}

	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err := conn.CreateAddonWithContext(ctx, input)
		if err != nil {
			if isAWSErr(err, eks.ErrCodeInvalidParameterException, "CREATE_FAILED") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, eks.ErrCodeInvalidParameterException, "does not exist") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.CreateAddonWithContext(ctx, input)
	}
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating EKS add-on (%s): %w", addonName, err))
	}

	d.SetId(fmt.Sprintf("%s:%s", clusterName, addonName))

	_, err = waiter.EksAddonCreated(ctx, conn, clusterName, addonName)
	if err != nil {
		// Creating addon w/o setting resolve_conflicts to "OVERWRITE"
		// might result in a failed creation, if unmanaged version of addon is already deployed
		// and there are configuration conflicts:
		// ConfigurationConflict	Apply failed with 1 conflict: conflict with "kubectl"...
		//
		// Addon resource is tainted after failed creation, thus will be deleted and created again.
		// Re-creating like this will resolve the error, but it will also purge any
		// configurations that were applied by the user (that were conflicting). This might we an unwanted
		// side effect and should be left for the user to decide how to handle it.
		return diag.FromErr(fmt.Errorf("unexpected EKS add-on (%s) state returned during creation: %w\n[WARNING] Running terraform apply again will remove the kubernetes add-on and attempt to create it again effectively purging previous add-on configuration",
			d.Id(), err))
	}

	return resourceAwsEksAddonRead(ctx, d, meta)
}

func resourceAwsEksAddonRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).eksconn
	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	clusterName, addonName, err := resourceAwsEksAddonParseId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &eks.DescribeAddonInput{
		ClusterName: aws.String(clusterName),
		AddonName:   aws.String(addonName),
	}

	log.Printf("[DEBUG] Reading EKS add-on: %s", d.Id())
	output, err := conn.DescribeAddonWithContext(ctx, input)
	if err != nil {
		if isAWSErr(err, eks.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] EKS add-on (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("error reading EKS add-on (%s): %w", d.Id(), err))
	}

	addon := output.Addon
	if addon == nil {
		log.Printf("[WARN] EKS add-on (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("cluster_name", addon.ClusterName)
	d.Set("addon_name", addon.AddonName)
	d.Set("arn", addon.AddonArn)
	d.Set("addon_version", addon.AddonVersion)
	d.Set("service_account_role_arn", addon.ServiceAccountRoleArn)
	d.Set("created_at", aws.TimeValue(addon.CreatedAt).Format(time.RFC3339))
	d.Set("modified_at", aws.TimeValue(addon.ModifiedAt).Format(time.RFC3339))

	tags := keyvaluetags.EksKeyValueTags(addon.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceAwsEksAddonUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).eksconn

	clusterName, addonName, err := resourceAwsEksAddonParseId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &eks.UpdateAddonInput{
		ClusterName:        aws.String(clusterName),
		AddonName:          aws.String(addonName),
		ClientRequestToken: aws.String(resource.UniqueId()),
	}

	if d.HasChange("resolve_conflicts") {
		if v, ok := d.GetOk("resolve_conflicts"); ok {
			d.Set("resolve_conflicts", aws.String(v.(string)))
		}
	}

	if v, ok := d.GetOk("resolve_conflicts"); ok {
		input.ResolveConflicts = aws.String(v.(string))
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := keyvaluetags.EksUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating tags: %w", err))
		}
	}

	if d.HasChange("addon_version") {
		input.AddonVersion = aws.String(d.Get("addon_version").(string))
	}

	if d.HasChange("service_account_role_arn") {
		input.ServiceAccountRoleArn = aws.String(d.Get("service_account_role_arn").(string))
	}

	if d.HasChanges("addon_version", "service_account_role_arn") {
		output, err := conn.UpdateAddonWithContext(ctx, input)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating EKS add-on (%s) version: %w", d.Id(), err))
		}

		if output == nil || output.Update == nil || output.Update.Id == nil {
			return diag.FromErr(fmt.Errorf("error determining EKS add-on (%s) version update ID: empty response", d.Id()))
		}

		updateID := aws.StringValue(output.Update.Id)

		_, err = waiter.EksAddonUpdateSuccessful(ctx, conn, clusterName, addonName, updateID)
		if err != nil {
			if d.Get("resolve_conflicts") != eks.ResolveConflictsOverwrite {
				// Changing addon version w/o setting resolve_conflicts to "OVERWRITE"
				// might result in a failed update if there are conflicts:
				// ConfigurationConflict	Apply failed with 1 conflict: conflict with "kubectl"...
				return diag.FromErr(fmt.Errorf("error waiting for EKS add-on (%s) update (%s): %w, consider setting attribute %q to %q",
					d.Id(), updateID, err, "resolve_conflicts", eks.ResolveConflictsOverwrite))
			}
			return diag.FromErr(fmt.Errorf("error waiting for EKS add-on (%s) update (%s): %w", d.Id(), updateID, err))
		}
	}

	return resourceAwsEksAddonRead(ctx, d, meta)
}

func resourceAwsEksAddonDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).eksconn

	clusterName, addonName, err := resourceAwsEksAddonParseId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &eks.DeleteAddonInput{
		ClusterName: aws.String(clusterName),
		AddonName:   aws.String(addonName),
	}

	_, err = conn.DeleteAddonWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting EKS add-on (%s): %w", d.Id(), err))
	}

	_, err = waiter.EksAddonDeleted(ctx, conn, clusterName, addonName)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error waiting for EKS add-on (%s) deletion: %w", d.Id(), err))
	}

	return nil
}

func resourceAwsEksAddonParseId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected cluster-name:addon-name", id)
	}

	return parts[0], parts[1], nil
}
