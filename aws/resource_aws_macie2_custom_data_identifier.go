package aws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
)

func resourceAwsMacie2CustomDataIdentifier() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMacie2CustomDataIdentifierCreate,
		ReadWithoutTimeout:   resourceMacie2CustomDataIdentifierRead,
		DeleteWithoutTimeout: resourceMacie2CustomDataIdentifierDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"keywords": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 50,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(3, 90),
				},
			},
			"ignore_words": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 10,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(4, 90),
				},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validation.StringLenBetween(0, 128),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validation.StringLenBetween(0, 128-resource.UniqueIDSuffixLength),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"maximum_match_distance": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 300),
			},
			"tags":     tagsSchemaForceNew(),
			"tags_all": tagsSchemaComputed(),
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMacie2CustomDataIdentifierCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &macie2.CreateCustomDataIdentifierInput{
		ClientToken: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("regex"); ok {
		input.Regex = aws.String(v.(string))
	}
	if v, ok := d.GetOk("keywords"); ok {
		input.Keywords = expandStringSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("ignore_words"); ok {
		input.IgnoreWords = expandStringSet(v.(*schema.Set))
	}
	input.Name = aws.String(naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string)))
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("maximum_match_distance"); ok {
		input.MaximumMatchDistance = aws.Int64(int64(v.(int)))
	}
	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().Macie2Tags()
	}

	var err error
	var output *macie2.CreateCustomDataIdentifierOutput
	err = resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
		output, err = conn.CreateCustomDataIdentifierWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeClientError) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		output, err = conn.CreateCustomDataIdentifierWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Macie CustomDataIdentifier: %w", err))
	}

	d.SetId(aws.StringValue(output.CustomDataIdentifierId))

	return resourceMacie2CustomDataIdentifierRead(ctx, d, meta)
}

func resourceMacie2CustomDataIdentifierRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig
	input := &macie2.GetCustomDataIdentifierInput{
		Id: aws.String(d.Id()),
	}

	resp, err := conn.GetCustomDataIdentifierWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
		tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
		log.Printf("[WARN] Macie CustomDataIdentifier (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Macie CustomDataIdentifier (%s): %w", d.Id(), err))
	}

	d.Set("regex", resp.Regex)
	if err = d.Set("keywords", flattenStringList(resp.Keywords)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie CustomDataIdentifier (%s): %w", "keywords", d.Id(), err))
	}
	if err = d.Set("ignore_words", flattenStringList(resp.IgnoreWords)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie CustomDataIdentifier (%s): %w", "ignore_words", d.Id(), err))
	}
	d.Set("name", resp.Name)
	d.Set("name_prefix", naming.NamePrefixFromName(aws.StringValue(resp.Name)))
	d.Set("description", resp.Description)
	d.Set("maximum_match_distance", resp.MaximumMatchDistance)
	tags := keyvaluetags.Macie2KeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	if err = d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie CustomDataIdentifier (%s): %w", "tags", d.Id(), err))
	}

	if err = d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie CustomDataIdentifier (%s): %w", "tags_all", d.Id(), err))
	}

	if aws.BoolValue(resp.Deleted) {
		log.Printf("[WARN] Macie CustomDataIdentifier (%s) is soft deleted, removing from state", d.Id())
		d.SetId("")
	}

	d.Set("created_at", aws.TimeValue(resp.CreatedAt).Format(time.RFC3339))
	d.Set("arn", resp.Arn)

	return nil
}

func resourceMacie2CustomDataIdentifierDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.DeleteCustomDataIdentifierInput{
		Id: aws.String(d.Id()),
	}

	_, err := conn.DeleteCustomDataIdentifierWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting Macie CustomDataIdentifier (%s): %w", d.Id(), err))
	}
	return nil
}
