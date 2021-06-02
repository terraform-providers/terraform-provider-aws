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

func resourceAwsMacie2FindingsFilter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMacie2FindingsFilterCreate,
		ReadWithoutTimeout:   resourceMacie2FindingsFilterRead,
		UpdateWithoutTimeout: resourceMacie2FindingsFilterUpdate,
		DeleteWithoutTimeout: resourceMacie2FindingsFilterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"finding_criteria": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"criterion": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"field": {
										Type:     schema.TypeString,
										Required: true,
									},
									"eq_exact_match": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"eq": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"neq": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"lt": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateTypeStringIsDateOrPositiveInt,
									},
									"lte": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateTypeStringIsDateOrPositiveInt,
									},
									"gt": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateTypeStringIsDateOrPositiveInt,
									},
									"gte": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateTypeStringIsDateOrPositiveInt,
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validation.StringLenBetween(3, 64),
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validation.StringLenBetween(3, 64-resource.UniqueIDSuffixLength),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"action": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(macie2.FindingsFilterAction_Values(), false),
			},
			"position": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"tags":     tagsSchema(),
			"tags_all": tagsSchemaComputed(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMacie2FindingsFilterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(keyvaluetags.New(d.Get("tags").(map[string]interface{})))

	input := &macie2.CreateFindingsFilterInput{
		ClientToken: aws.String(resource.UniqueId()),
		Name:        aws.String(naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string))),
		Action:      aws.String(d.Get("action").(string)),
	}

	var err error
	input.FindingCriteria, err = expandFindingCriteriaFilter(d.Get("finding_criteria").([]interface{}))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Macie FindingsFilter: %w", err))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("position"); ok {
		input.Position = aws.Int64(int64(v.(int)))
	}
	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().Macie2Tags()
	}

	var output *macie2.CreateFindingsFilterOutput
	err = resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
		output, err = conn.CreateFindingsFilterWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeClientError) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		output, err = conn.CreateFindingsFilterWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Macie FindingsFilter: %w", err))
	}

	d.SetId(aws.StringValue(output.Id))

	return resourceMacie2FindingsFilterRead(ctx, d, meta)
}

func resourceMacie2FindingsFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	defaultTagsConfig := meta.(*AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig
	input := &macie2.GetFindingsFilterInput{
		Id: aws.String(d.Id()),
	}

	resp, err := conn.GetFindingsFilterWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
			log.Printf("[WARN] Macie FindingsFilter (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("error reading Macie FindingsFilter (%s): %w", d.Id(), err))
	}

	if err = d.Set("finding_criteria", flattenFindingCriteriaFindingsFilter(resp.FindingCriteria)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie FindingsFilter (%s): %w", "finding_criteria", d.Id(), err))
	}
	d.Set("name", resp.Name)
	d.Set("name_prefix", naming.NamePrefixFromName(aws.StringValue(resp.Name)))
	d.Set("description", resp.Description)
	d.Set("action", resp.Action)
	d.Set("position", resp.Position)
	tags := keyvaluetags.Macie2KeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	if err = d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie FindingsFilter (%s): %w", "tags", d.Id(), err))
	}

	if err = d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie FindingsFilter (%s): %w", "tags_all", d.Id(), err))
	}

	d.Set("arn", resp.Arn)

	return nil
}

func resourceMacie2FindingsFilterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.UpdateFindingsFilterInput{
		Id: aws.String(d.Id()),
	}

	var err error
	if d.HasChange("finding_criteria") {
		input.FindingCriteria, err = expandFindingCriteriaFilter(d.Get("finding_criteria").([]interface{}))
		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating Macie FindingsFilter (%s): %w", d.Id(), err))
		}
	}
	if d.HasChange("name") {
		input.Name = aws.String(naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string)))
	}
	if d.HasChange("name_prefix") {
		input.Name = aws.String(naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string)))
	}
	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}
	if d.HasChange("action") {
		input.Action = aws.String(d.Get("action").(string))
	}
	if d.HasChange("position") {
		input.Position = aws.Int64(int64(d.Get("position").(int)))
	}

	_, err = conn.UpdateFindingsFilterWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating Macie FindingsFilter (%s): %w", d.Id(), err))
	}

	return resourceMacie2FindingsFilterRead(ctx, d, meta)
}

func resourceMacie2FindingsFilterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.DeleteFindingsFilterInput{
		Id: aws.String(d.Id()),
	}

	_, err := conn.DeleteFindingsFilterWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrCodeResourceNotFoundException) ||
			tfawserr.ErrMessageContains(err, macie2.ErrCodeAccessDeniedException, "Macie is not enabled") {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting Macie FindingsFilter (%s): %w", d.Id(), err))
	}
	return nil
}

func expandFindingCriteriaFilter(findingCriterias []interface{}) (*macie2.FindingCriteria, error) {
	if len(findingCriterias) == 0 {
		return nil, nil
	}

	criteria := map[string]*macie2.CriterionAdditionalProperties{}
	findingCriteria := findingCriterias[0].(map[string]interface{})
	inputFindingCriteria := findingCriteria["criterion"].(*schema.Set).List()

	for _, criterion := range inputFindingCriteria {
		crit := criterion.(map[string]interface{})
		field := crit["field"].(string)
		conditional := macie2.CriterionAdditionalProperties{}

		if v, ok := crit["eq"].(*schema.Set); ok && v.Len() != 0 {
			foo := make([]*string, v.Len())
			for i, v1 := range v.List() {
				s := v1.(string)
				foo[i] = &s
			}
			conditional.Eq = foo
		}
		if v, ok := crit["neq"].(*schema.Set); ok && v.Len() != 0 {
			foo := make([]*string, v.Len())
			for i, v1 := range v.List() {
				s := v1.(string)
				foo[i] = &s
			}
			conditional.Neq = foo
		}
		if v, ok := crit["eq_exact_match"].(*schema.Set); ok && v.Len() != 0 {
			foo := make([]*string, v.Len())
			for i, v1 := range v.List() {
				s := v1.(string)
				foo[i] = &s
			}
			conditional.EqExactMatch = foo
		}
		if v, ok := crit["lt"].(string); ok && v != "" {
			i, err := expandConditionIntField(field, v)
			if err != nil {
				return nil, fmt.Errorf("error parsing condition %q for field %q: %w", "lt", field, err)
			}
			conditional.Lt = aws.Int64(i)
		}
		if v, ok := crit["lte"].(string); ok && v != "" {
			i, err := expandConditionIntField(field, v)
			if err != nil {
				return nil, fmt.Errorf("error parsing condition %q for field %q: %w", "lte", field, err)
			}
			conditional.Lte = aws.Int64(i)
		}
		if v, ok := crit["gt"].(string); ok && v != "" {
			i, err := expandConditionIntField(field, v)
			if err != nil {
				return nil, fmt.Errorf("error parsing condition %q for field %q: %w", "gt", field, err)
			}
			conditional.Gt = aws.Int64(i)
		}
		if v, ok := crit["gte"].(string); ok && v != "" {
			i, err := expandConditionIntField(field, v)
			if err != nil {
				return nil, fmt.Errorf("error parsing condition %q for field %q: %w", "gte", field, err)
			}
			conditional.Gte = aws.Int64(i)
		}
		criteria[field] = &conditional
	}

	return &macie2.FindingCriteria{Criterion: criteria}, nil
}

func flattenFindingCriteriaFindingsFilter(findingCriteria *macie2.FindingCriteria) []interface{} {
	if findingCriteria == nil {
		return nil
	}

	var flatCriteria []interface{}

	for field, conditions := range findingCriteria.Criterion {
		criterion := map[string]interface{}{
			"field": field,
		}
		if len(conditions.Eq) != 0 {
			criterion["eq"] = aws.StringValueSlice(conditions.Eq)
		}
		if len(conditions.Neq) != 0 {
			criterion["neq"] = aws.StringValueSlice(conditions.Neq)
		}
		if len(conditions.EqExactMatch) != 0 {
			criterion["eq_exact_match"] = aws.StringValueSlice(conditions.EqExactMatch)
		}
		if v := aws.Int64Value(conditions.Lt); v != 0 {
			criterion["lt"] = flattenConditionIntField(field, v)
		}
		if v := aws.Int64Value(conditions.Lte); v != 0 {
			criterion["lte"] = flattenConditionIntField(field, v)
		}
		if v := aws.Int64Value(conditions.Gt); v != 0 {
			criterion["gt"] = flattenConditionIntField(field, v)
		}
		if v := aws.Int64Value(conditions.Gte); v != 0 {
			criterion["gte"] = flattenConditionIntField(field, v)
		}
		flatCriteria = append(flatCriteria, criterion)
	}

	return []interface{}{
		map[string][]interface{}{
			"criterion": flatCriteria,
		},
	}
}
