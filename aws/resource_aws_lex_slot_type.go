package aws

import (
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/lex/waiter"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	LexSlotTypeCreateTimeout = 1 * time.Minute
	LexSlotTypeUpdateTimeout = 1 * time.Minute
	LexSlotTypeDeleteTimeout = 5 * time.Minute
	LexSlotTypeVersionLatest = "$LATEST"
)

func resourceAwsLexSlotType() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsLexSlotTypeCreate,
		Read:   resourceAwsLexSlotTypeRead,
		Update: resourceAwsLexSlotTypeUpdate,
		Delete: resourceAwsLexSlotTypeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(LexSlotTypeCreateTimeout),
			Update: schema.DefaultTimeout(LexSlotTypeUpdateTimeout),
			Delete: schema.DefaultTimeout(LexSlotTypeDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"checksum": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"create_version": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "",
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"enumeration_value": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 10000,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"synonyms": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 140),
							},
						},
						"value": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 140),
						},
					},
				},
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 100),
					validation.StringMatch(regexp.MustCompile(`^((AMAZON\.)_?|[A-Za-z]_?)+`), ""),
				),
			},
			"value_selection_strategy": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      lexmodelbuildingservice.SlotValueSelectionStrategyOriginalValue,
				ValidateFunc: validation.StringInSlice(lexmodelbuildingservice.SlotValueSelectionStrategy_Values(), false),
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsLexSlotTypeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn
	name := d.Get("name").(string)

	input := &lexmodelbuildingservice.PutSlotTypeInput{
		CreateVersion:          aws.Bool(d.Get("create_version").(bool)),
		Description:            aws.String(d.Get("description").(string)),
		Name:                   aws.String(name),
		ValueSelectionStrategy: aws.String(d.Get("value_selection_strategy").(string)),
	}

	if v, ok := d.GetOk("enumeration_value"); ok {
		input.EnumerationValues = expandLexEnumerationValues(v.(*schema.Set).List())
	}

	err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		output, err := conn.PutSlotType(input)

		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeConflictException) {
			input.Checksum = output.Checksum
			return resource.RetryableError(fmt.Errorf("%q slot type still creating, another operation is pending: %s", d.Id(), err))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutSlotType(input)
	}

	if err != nil {
		return fmt.Errorf("error creating slot type %s: %w", name, err)
	}

	d.SetId(name)

	return resourceAwsLexSlotTypeRead(d, meta)
}

func resourceAwsLexSlotTypeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	resp, err := conn.GetSlotType(&lexmodelbuildingservice.GetSlotTypeInput{
		Name:    aws.String(d.Id()),
		Version: aws.String(LexSlotTypeVersionLatest),
	})
	if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error getting slot type %s: %w", d.Id(), err)
	}

	d.Set("checksum", resp.Checksum)
	d.Set("created_date", resp.CreatedDate.Format(time.RFC3339))
	d.Set("description", resp.Description)
	d.Set("last_updated_date", resp.LastUpdatedDate.Format(time.RFC3339))
	d.Set("name", resp.Name)
	d.Set("value_selection_strategy", resp.ValueSelectionStrategy)

	if resp.EnumerationValues != nil {
		d.Set("enumeration_value", flattenLexEnumerationValues(resp.EnumerationValues))
	}

	version, err := getLatestLexSlotTypeVersion(conn, &lexmodelbuildingservice.GetSlotTypeVersionsInput{
		Name: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}
	d.Set("version", version)

	return nil
}

func getLatestLexSlotTypeVersion(conn *lexmodelbuildingservice.LexModelBuildingService, input *lexmodelbuildingservice.GetSlotTypeVersionsInput) (string, error) {
	version := LexSlotTypeVersionLatest

	for {
		page, err := conn.GetSlotTypeVersions(input)
		if err != nil {
			return "", err
		}

		// At least 1 version will always be returned.
		if len(page.SlotTypes) == 1 {
			break
		}

		for _, slotType := range page.SlotTypes {
			if *slotType.Version == LexSlotTypeVersionLatest {
				continue
			}
			if *slotType.Version > version {
				version = *slotType.Version
			}
		}

		if page.NextToken == nil {
			break
		}
		input.NextToken = page.NextToken
	}

	return version, nil
}

func resourceAwsLexSlotTypeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	input := &lexmodelbuildingservice.PutSlotTypeInput{
		Checksum:               aws.String(d.Get("checksum").(string)),
		CreateVersion:          aws.Bool(d.Get("create_version").(bool)),
		Description:            aws.String(d.Get("description").(string)),
		Name:                   aws.String(d.Id()),
		ValueSelectionStrategy: aws.String(d.Get("value_selection_strategy").(string)),
	}

	if v, ok := d.GetOk("enumeration_value"); ok {
		input.EnumerationValues = expandLexEnumerationValues(v.(*schema.Set).List())
	}

	err := resource.Retry(d.Timeout(schema.TimeoutUpdate), func() *resource.RetryError {
		_, err := conn.PutSlotType(input)

		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeConflictException) {
			return resource.RetryableError(fmt.Errorf("%q: slot type still updating", d.Id()))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.PutSlotType(input)
	}

	if err != nil {
		return fmt.Errorf("error updating slot type %s: %w", d.Id(), err)
	}

	return resourceAwsLexSlotTypeRead(d, meta)
}

func resourceAwsLexSlotTypeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).lexmodelconn

	input := &lexmodelbuildingservice.DeleteSlotTypeInput{
		Name: aws.String(d.Id()),
	}

	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		_, err := conn.DeleteSlotType(input)

		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeConflictException) {
			return resource.RetryableError(fmt.Errorf("%q: there is a pending operation, slot type still deleting", d.Id()))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteSlotType(input)
	}

	if err != nil {
		return fmt.Errorf("error deleting slot type %s: %w", d.Id(), err)
	}

	_, err = waiter.LexSlotTypeDeleted(conn, d.Id())

	return err
}

func flattenLexEnumerationValues(values []*lexmodelbuildingservice.EnumerationValue) (flattened []map[string]interface{}) {
	for _, value := range values {
		flattened = append(flattened, map[string]interface{}{
			"synonyms": flattenStringList(value.Synonyms),
			"value":    aws.StringValue(value.Value),
		})
	}

	return
}

func expandLexEnumerationValues(rawValues []interface{}) []*lexmodelbuildingservice.EnumerationValue {
	enums := make([]*lexmodelbuildingservice.EnumerationValue, 0, len(rawValues))
	for _, rawValue := range rawValues {
		value, ok := rawValue.(map[string]interface{})
		if !ok {
			continue
		}

		enums = append(enums, &lexmodelbuildingservice.EnumerationValue{
			Synonyms: expandStringSet(value["synonyms"].(*schema.Set)),
			Value:    aws.String(value["value"].(string)),
		})
	}
	return enums
}
