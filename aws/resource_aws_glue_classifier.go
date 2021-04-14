package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsGlueClassifier() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGlueClassifierCreate,
		Read:   resourceAwsGlueClassifierRead,
		Update: resourceAwsGlueClassifierUpdate,
		Delete: resourceAwsGlueClassifierDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		CustomizeDiff: customdiff.Sequence(
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				// ForceNew when changing classifier type
				// InvalidInputException: UpdateClassifierRequest can't change the type of the classifier
				if diff.HasChange("csv_classifier") && diff.HasChange("grok_classifier") {
					diff.ForceNew("csv_classifier")
					diff.ForceNew("grok_classifier")
				}
				if diff.HasChange("csv_classifier") && diff.HasChange("json_classifier") {
					diff.ForceNew("csv_classifier")
					diff.ForceNew("json_classifier")
				}
				if diff.HasChange("csv_classifier") && diff.HasChange("xml_classifier") {
					diff.ForceNew("csv_classifier")
					diff.ForceNew("xml_classifier")
				}
				if diff.HasChange("grok_classifier") && diff.HasChange("json_classifier") {
					diff.ForceNew("grok_classifier")
					diff.ForceNew("json_classifier")
				}
				if diff.HasChange("grok_classifier") && diff.HasChange("xml_classifier") {
					diff.ForceNew("grok_classifier")
					diff.ForceNew("xml_classifier")
				}
				if diff.HasChange("json_classifier") && diff.HasChange("xml_classifier") {
					diff.ForceNew("json_classifier")
					diff.ForceNew("xml_classifier")
				}
				return nil
			},
		),

		Schema: map[string]*schema.Schema{
			"csv_classifier": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"grok_classifier", "json_classifier", "xml_classifier"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_single_column": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"contains_header": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(glue.CsvHeaderOption_Values(), false),
						},
						"delimiter": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"disable_value_trimming": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"header": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"quote_symbol": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"grok_classifier": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"csv_classifier", "json_classifier", "xml_classifier"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"classification": {
							Type:     schema.TypeString,
							Required: true,
						},
						"custom_patterns": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 16000),
						},
						"grok_pattern": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
					},
				},
			},
			"json_classifier": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"csv_classifier", "grok_classifier", "xml_classifier"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"json_path": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"xml_classifier": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"csv_classifier", "grok_classifier", "json_classifier"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"classification": {
							Type:     schema.TypeString,
							Required: true,
						},
						"row_tag": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsGlueClassifierCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn
	name := d.Get("name").(string)

	input := &glue.CreateClassifierInput{}

	if v, ok := d.GetOk("csv_classifier"); ok {
		m := v.([]interface{})[0].(map[string]interface{})
		input.CsvClassifier = expandGlueCsvClassifierCreate(name, m)
	}

	if v, ok := d.GetOk("grok_classifier"); ok {
		m := v.([]interface{})[0].(map[string]interface{})
		input.GrokClassifier = expandGlueGrokClassifierCreate(name, m)
	}

	if v, ok := d.GetOk("json_classifier"); ok {
		m := v.([]interface{})[0].(map[string]interface{})
		input.JsonClassifier = expandGlueJsonClassifierCreate(name, m)
	}

	if v, ok := d.GetOk("xml_classifier"); ok {
		m := v.([]interface{})[0].(map[string]interface{})
		input.XMLClassifier = expandGlueXmlClassifierCreate(name, m)
	}

	log.Printf("[DEBUG] Creating Glue Classifier: %s", input)
	_, err := conn.CreateClassifier(input)
	if err != nil {
		return fmt.Errorf("error creating Glue Classifier (%s): %s", name, err)
	}

	d.SetId(name)

	return resourceAwsGlueClassifierRead(d, meta)
}

func resourceAwsGlueClassifierRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	input := &glue.GetClassifierInput{
		Name: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Glue Classifier: %s", input)
	output, err := conn.GetClassifier(input)
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			log.Printf("[WARN] Glue Classifier (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Glue Classifier (%s): %s", d.Id(), err)
	}

	classifier := output.Classifier
	if classifier == nil {
		log.Printf("[WARN] Glue Classifier (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := d.Set("csv_classifier", flattenGlueCsvClassifier(classifier.CsvClassifier)); err != nil {
		return fmt.Errorf("error setting match_criteria: %s", err)
	}

	if err := d.Set("grok_classifier", flattenGlueGrokClassifier(classifier.GrokClassifier)); err != nil {
		return fmt.Errorf("error setting match_criteria: %s", err)
	}

	if err := d.Set("json_classifier", flattenGlueJsonClassifier(classifier.JsonClassifier)); err != nil {
		return fmt.Errorf("error setting json_classifier: %s", err)
	}

	d.Set("name", d.Id())

	if err := d.Set("xml_classifier", flattenGlueXmlClassifier(classifier.XMLClassifier)); err != nil {
		return fmt.Errorf("error setting xml_classifier: %s", err)
	}

	return nil
}

func resourceAwsGlueClassifierUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	input := &glue.UpdateClassifierInput{}

	if v, ok := d.GetOk("csv_classifier"); ok {
		m := v.([]interface{})[0].(map[string]interface{})
		input.CsvClassifier = expandGlueCsvClassifierUpdate(d.Id(), m)
	}

	if v, ok := d.GetOk("grok_classifier"); ok {
		m := v.([]interface{})[0].(map[string]interface{})
		input.GrokClassifier = expandGlueGrokClassifierUpdate(d.Id(), m)
	}

	if v, ok := d.GetOk("json_classifier"); ok {
		m := v.([]interface{})[0].(map[string]interface{})
		input.JsonClassifier = expandGlueJsonClassifierUpdate(d.Id(), m)
	}

	if v, ok := d.GetOk("xml_classifier"); ok {
		m := v.([]interface{})[0].(map[string]interface{})
		input.XMLClassifier = expandGlueXmlClassifierUpdate(d.Id(), m)
	}

	log.Printf("[DEBUG] Updating Glue Classifier: %s", input)
	_, err := conn.UpdateClassifier(input)
	if err != nil {
		return fmt.Errorf("error updating Glue Classifier (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceAwsGlueClassifierDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).glueconn

	log.Printf("[DEBUG] Deleting Glue Classifier: %s", d.Id())
	err := deleteGlueClassifier(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error deleting Glue Classifier (%s): %s", d.Id(), err)
	}

	return nil
}

func deleteGlueClassifier(conn *glue.Glue, name string) error {
	input := &glue.DeleteClassifierInput{
		Name: aws.String(name),
	}

	_, err := conn.DeleteClassifier(input)
	if err != nil {
		if isAWSErr(err, glue.ErrCodeEntityNotFoundException, "") {
			return nil
		}
		return err
	}

	return nil
}

func expandGlueCsvClassifierCreate(name string, m map[string]interface{}) *glue.CreateCsvClassifierRequest {
	csvClassifier := &glue.CreateCsvClassifierRequest{
		AllowSingleColumn:    aws.Bool(m["allow_single_column"].(bool)),
		ContainsHeader:       aws.String(m["contains_header"].(string)),
		Delimiter:            aws.String(m["delimiter"].(string)),
		DisableValueTrimming: aws.Bool(m["disable_value_trimming"].(bool)),
		Name:                 aws.String(name),
	}

	if v, ok := m["quote_symbol"].(string); ok && v != "" {
		csvClassifier.QuoteSymbol = aws.String(v)
	}

	if v, ok := m["header"].([]interface{}); ok {
		csvClassifier.Header = expandStringList(v)
	}

	return csvClassifier
}

func expandGlueCsvClassifierUpdate(name string, m map[string]interface{}) *glue.UpdateCsvClassifierRequest {
	csvClassifier := &glue.UpdateCsvClassifierRequest{
		AllowSingleColumn:    aws.Bool(m["allow_single_column"].(bool)),
		ContainsHeader:       aws.String(m["contains_header"].(string)),
		Delimiter:            aws.String(m["delimiter"].(string)),
		DisableValueTrimming: aws.Bool(m["disable_value_trimming"].(bool)),
		Name:                 aws.String(name),
	}

	if v, ok := m["quote_symbol"].(string); ok && v != "" {
		csvClassifier.QuoteSymbol = aws.String(v)
	}

	if v, ok := m["header"].([]interface{}); ok {
		csvClassifier.Header = expandStringList(v)
	}

	return csvClassifier
}

func expandGlueGrokClassifierCreate(name string, m map[string]interface{}) *glue.CreateGrokClassifierRequest {
	grokClassifier := &glue.CreateGrokClassifierRequest{
		Classification: aws.String(m["classification"].(string)),
		GrokPattern:    aws.String(m["grok_pattern"].(string)),
		Name:           aws.String(name),
	}

	if v, ok := m["custom_patterns"]; ok && v.(string) != "" {
		grokClassifier.CustomPatterns = aws.String(v.(string))
	}

	return grokClassifier
}

func expandGlueGrokClassifierUpdate(name string, m map[string]interface{}) *glue.UpdateGrokClassifierRequest {
	grokClassifier := &glue.UpdateGrokClassifierRequest{
		Classification: aws.String(m["classification"].(string)),
		GrokPattern:    aws.String(m["grok_pattern"].(string)),
		Name:           aws.String(name),
	}

	if v, ok := m["custom_patterns"]; ok && v.(string) != "" {
		grokClassifier.CustomPatterns = aws.String(v.(string))
	}

	return grokClassifier
}

func expandGlueJsonClassifierCreate(name string, m map[string]interface{}) *glue.CreateJsonClassifierRequest {
	jsonClassifier := &glue.CreateJsonClassifierRequest{
		JsonPath: aws.String(m["json_path"].(string)),
		Name:     aws.String(name),
	}

	return jsonClassifier
}

func expandGlueJsonClassifierUpdate(name string, m map[string]interface{}) *glue.UpdateJsonClassifierRequest {
	jsonClassifier := &glue.UpdateJsonClassifierRequest{
		JsonPath: aws.String(m["json_path"].(string)),
		Name:     aws.String(name),
	}

	return jsonClassifier
}

func expandGlueXmlClassifierCreate(name string, m map[string]interface{}) *glue.CreateXMLClassifierRequest {
	xmlClassifier := &glue.CreateXMLClassifierRequest{
		Classification: aws.String(m["classification"].(string)),
		Name:           aws.String(name),
		RowTag:         aws.String(m["row_tag"].(string)),
	}

	return xmlClassifier
}

func expandGlueXmlClassifierUpdate(name string, m map[string]interface{}) *glue.UpdateXMLClassifierRequest {
	xmlClassifier := &glue.UpdateXMLClassifierRequest{
		Classification: aws.String(m["classification"].(string)),
		Name:           aws.String(name),
		RowTag:         aws.String(m["row_tag"].(string)),
	}

	if v, ok := m["row_tag"]; ok && v.(string) != "" {
		xmlClassifier.RowTag = aws.String(v.(string))
	}

	return xmlClassifier
}

func flattenGlueCsvClassifier(csvClassifier *glue.CsvClassifier) []map[string]interface{} {
	if csvClassifier == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"allow_single_column":    aws.BoolValue(csvClassifier.AllowSingleColumn),
		"contains_header":        aws.StringValue(csvClassifier.ContainsHeader),
		"delimiter":              aws.StringValue(csvClassifier.Delimiter),
		"disable_value_trimming": aws.BoolValue(csvClassifier.DisableValueTrimming),
		"header":                 aws.StringValueSlice(csvClassifier.Header),
		"quote_symbol":           aws.StringValue(csvClassifier.QuoteSymbol),
	}

	return []map[string]interface{}{m}
}

func flattenGlueGrokClassifier(grokClassifier *glue.GrokClassifier) []map[string]interface{} {
	if grokClassifier == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"classification":  aws.StringValue(grokClassifier.Classification),
		"custom_patterns": aws.StringValue(grokClassifier.CustomPatterns),
		"grok_pattern":    aws.StringValue(grokClassifier.GrokPattern),
	}

	return []map[string]interface{}{m}
}

func flattenGlueJsonClassifier(jsonClassifier *glue.JsonClassifier) []map[string]interface{} {
	if jsonClassifier == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"json_path": aws.StringValue(jsonClassifier.JsonPath),
	}

	return []map[string]interface{}{m}
}

func flattenGlueXmlClassifier(xmlClassifier *glue.XMLClassifier) []map[string]interface{} {
	if xmlClassifier == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"classification": aws.StringValue(xmlClassifier.Classification),
		"row_tag":        aws.StringValue(xmlClassifier.RowTag),
	}

	return []map[string]interface{}{m}
}
