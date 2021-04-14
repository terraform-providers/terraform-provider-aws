package aws

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsGuardDutyFilter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGuardDutyFilterCreate,
		Read:   resourceAwsGuardDutyFilterRead,
		Update: resourceAwsGuardDutyFilterUpdate,
		Delete: resourceAwsGuardDutyFilterDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"detector_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 64),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
			"tags": tagsSchema(),
			"finding_criteria": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"criterion": {
							Type:     schema.TypeSet,
							MinItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"field": {
										Type:     schema.TypeString,
										Required: true,
									},
									"equals": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"not_equals": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 1,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"greater_than": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateTypeStringIsDateOrPositiveInt,
									},
									"greater_than_or_equal": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateTypeStringIsDateOrPositiveInt,
									},
									"less_than": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateTypeStringIsDateOrPositiveInt,
									},
									"less_than_or_equal": {
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
			"action": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					guardduty.FilterActionNoop,
					guardduty.FilterActionArchive,
				}, false),
			},
			"rank": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceAwsGuardDutyFilterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	input := guardduty.CreateFilterInput{
		Action:      aws.String(d.Get("action").(string)),
		Description: aws.String(d.Get("description").(string)),
		DetectorId:  aws.String(d.Get("detector_id").(string)),
		Name:        aws.String(d.Get("name").(string)),
		Rank:        aws.Int64(int64(d.Get("rank").(int))),
	}

	var err error
	input.FindingCriteria, err = expandFindingCriteria(d.Get("finding_criteria").([]interface{}))
	if err != nil {
		return err
	}

	if v, ok := d.GetOk("tags"); ok {
		tags := v.(map[string]interface{})
		if len(tags) > 0 {
			input.Tags = keyvaluetags.New(tags).GuarddutyTags()
		}
	}

	log.Printf("[DEBUG] Creating GuardDuty Filter: %s", input)
	output, err := conn.CreateFilter(&input)
	if err != nil {
		return fmt.Errorf("error creating GuardDuty Filter: %w", err)
	}

	d.SetId(guardDutyFilterCreateID(d.Get("detector_id").(string), aws.StringValue(output.Name)))

	return resourceAwsGuardDutyFilterRead(d, meta)
}

func resourceAwsGuardDutyFilterRead(d *schema.ResourceData, meta interface{}) error {
	var detectorID, name string
	var err error

	if _, ok := d.GetOk("detector_id"); !ok {
		// If there is no "detector_id" passed, then it's an import.
		detectorID, name, err = guardDutyFilterParseID(d.Id())
		if err != nil {
			return err
		}
	} else {
		detectorID = d.Get("detector_id").(string)
		name = d.Get("name").(string)
	}

	conn := meta.(*AWSClient).guarddutyconn

	input := guardduty.GetFilterInput{
		DetectorId: aws.String(detectorID),
		FilterName: aws.String(name),
	}

	log.Printf("[DEBUG] Reading GuardDuty Filter: %s", input)
	filter, err := conn.GetFilter(&input)

	if err != nil {
		if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected since no such resource found.") {
			log.Printf("[WARN] GuardDuty detector %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading GuardDuty Filter '%s': %w", name, err)
	}

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "guardduty",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("detector/%s/filter/%s", detectorID, name),
	}.String()
	d.Set("arn", arn)

	err = d.Set("finding_criteria", flattenFindingCriteria(filter.FindingCriteria))
	if err != nil {
		return fmt.Errorf("Setting GuardDuty Filter FindingCriteria failed: %w", err)
	}

	d.Set("action", filter.Action)
	d.Set("description", filter.Description)
	d.Set("name", filter.Name)
	d.Set("detector_id", detectorID)
	d.Set("rank", filter.Rank)
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig
	d.Set("tags", keyvaluetags.GuarddutyKeyValueTags(filter.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map())
	d.SetId(guardDutyFilterCreateID(detectorID, name))

	return nil
}

func resourceAwsGuardDutyFilterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	if d.HasChanges("action", "description", "finding_criteria", "rank") {
		input := guardduty.UpdateFilterInput{
			Action:      aws.String(d.Get("action").(string)),
			Description: aws.String(d.Get("description").(string)),
			DetectorId:  aws.String(d.Get("detector_id").(string)),
			FilterName:  aws.String(d.Get("name").(string)),
			Rank:        aws.Int64(int64(d.Get("rank").(int))),
		}

		var err error
		input.FindingCriteria, err = expandFindingCriteria(d.Get("finding_criteria").([]interface{}))
		if err != nil {
			return err
		}

		log.Printf("[DEBUG] Updating GuardDuty Filter: %s", input)

		_, err = conn.UpdateFilter(&input)
		if err != nil {
			return fmt.Errorf("error updating GuardDuty Filter %s: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.GuarddutyUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating GuardDuty Filter (%s) tags: %s", d.Get("arn").(string), err)
		}
	}

	return resourceAwsGuardDutyFilterRead(d, meta)
}

func resourceAwsGuardDutyFilterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).guarddutyconn

	detectorId := d.Get("detector_id").(string)
	name := d.Get("name").(string)

	input := guardduty.DeleteFilterInput{
		FilterName: aws.String(name),
		DetectorId: aws.String(detectorId),
	}

	log.Printf("[DEBUG] Delete GuardDuty Filter: %s", input)

	_, err := conn.DeleteFilter(&input)
	if tfawserr.ErrMessageContains(err, guardduty.ErrCodeBadRequestException, "The request is rejected since no such resource found.") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting GuardDuty Filter %s: %w", d.Id(), err)
	}
	return nil
}

const guardDutyFilterIDSeparator = ":"

func guardDutyFilterCreateID(detectorID, filterName string) string {
	return detectorID + guardDutyFilterIDSeparator + filterName
}

func guardDutyFilterParseID(importedId string) (string, string, error) {
	parts := strings.Split(importedId, guardDutyFilterIDSeparator)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("GuardDuty filter ID must be of the form <Detector ID>:<Filter name>. Got %q.", importedId)
	}
	return parts[0], parts[1], nil
}

func expandFindingCriteria(raw []interface{}) (*guardduty.FindingCriteria, error) {
	findingCriteria := raw[0].(map[string]interface{})
	inputFindingCriteria := findingCriteria["criterion"].(*schema.Set).List()

	criteria := map[string]*guardduty.Condition{}
	for _, criterion := range inputFindingCriteria {
		typedCriterion := criterion.(map[string]interface{})
		field := typedCriterion["field"].(string)

		condition := guardduty.Condition{}
		if x, ok := typedCriterion["equals"]; ok {
			if v, ok := x.([]interface{}); ok && len(v) > 0 {
				foo := make([]*string, len(v))
				for i := range v {
					s := v[i].(string)
					foo[i] = &s
				}
				condition.Equals = foo
			}
		}
		if x, ok := typedCriterion["not_equals"]; ok {
			if v, ok := x.([]interface{}); ok && len(v) > 0 {
				foo := make([]*string, len(v))
				for i := range v {
					s := v[i].(string)
					foo[i] = &s
				}
				condition.NotEquals = foo
			}
		}
		if x, ok := typedCriterion["greater_than"]; ok {
			if v, ok := x.(string); ok && v != "" {
				i, err := expandConditionIntField(field, v)
				if err != nil {
					return nil, fmt.Errorf("error parsing condition %q for field %q: %w", "greater_than", field, err)
				}
				condition.GreaterThan = aws.Int64(i)
			}
		}
		if x, ok := typedCriterion["greater_than_or_equal"]; ok {
			if v, ok := x.(string); ok && v != "" {
				i, err := expandConditionIntField(field, v)
				if err != nil {
					return nil, fmt.Errorf("error parsing condition %q for field %q: %w", "greater_than_or_equal", field, err)
				}
				condition.GreaterThanOrEqual = aws.Int64(i)
			}
		}
		if x, ok := typedCriterion["less_than"]; ok {
			if v, ok := x.(string); ok && v != "" {
				i, err := expandConditionIntField(field, v)
				if err != nil {
					return nil, fmt.Errorf("error parsing condition %q for field %q: %w", "less_than", field, err)
				}
				condition.LessThan = aws.Int64(i)
			}
		}
		if x, ok := typedCriterion["less_than_or_equal"]; ok {
			if v, ok := x.(string); ok && v != "" {
				i, err := expandConditionIntField(field, v)
				if err != nil {
					return nil, fmt.Errorf("error parsing condition %q for field %q: %w", "less_than_or_equal", field, err)
				}
				condition.LessThanOrEqual = aws.Int64(i)
			}
		}
		criteria[field] = &condition
	}

	return &guardduty.FindingCriteria{Criterion: criteria}, nil
}

func expandConditionIntField(field, v string) (int64, error) {
	if field == "updatedAt" {
		date, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return 0, err
		}
		return date.UnixNano() / 1000000, nil
	}

	return strconv.ParseInt(v, 10, 64)
}

func flattenFindingCriteria(findingCriteriaRemote *guardduty.FindingCriteria) []interface{} {
	var flatCriteria []interface{}

	for field, conditions := range findingCriteriaRemote.Criterion {
		criterion := map[string]interface{}{
			"field": field,
		}
		if len(conditions.Equals) > 0 {
			criterion["equals"] = aws.StringValueSlice(conditions.Equals)
		}
		if len(conditions.NotEquals) > 0 {
			criterion["not_equals"] = aws.StringValueSlice(conditions.NotEquals)
		}
		if v := aws.Int64Value(conditions.GreaterThan); v > 0 {
			criterion["greater_than"] = flattenConditionIntField(field, v)
		}
		if v := aws.Int64Value(conditions.GreaterThanOrEqual); v > 0 {
			criterion["greater_than_or_equal"] = flattenConditionIntField(field, v)
		}
		if v := aws.Int64Value(conditions.LessThan); v > 0 {
			criterion["less_than"] = flattenConditionIntField(field, v)
		}
		if v := aws.Int64Value(conditions.LessThanOrEqual); v > 0 {
			criterion["less_than_or_equal"] = flattenConditionIntField(field, v)
		}
		flatCriteria = append(flatCriteria, criterion)
	}

	return []interface{}{
		map[string][]interface{}{
			"criterion": flatCriteria,
		},
	}
}

func flattenConditionIntField(field string, v int64) string {
	if field == "updatedAt" {
		seconds := v / 1000
		nanoseconds := v % 1000
		date := time.Unix(seconds, nanoseconds).UTC()
		return date.Format(time.RFC3339)
	}
	return strconv.FormatInt(v, 10)
}
