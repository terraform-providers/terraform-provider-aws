package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/accessanalyzer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const (
	// Maximum amount of time to wait for Organizations eventual consistency on creation
	// This timeout value is much higher than usual since the cross-service validation
	// appears to be consistently caching for 5 minutes:
	// --- PASS: TestAccAWSAccessAnalyzer_serial/Analyzer/Type_Organization (315.86s)
	accessAnalyzerOrganizationCreationTimeout = 10 * time.Minute
)

func resourceAwsAccessAnalyzerAnalyzer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAccessAnalyzerAnalyzerCreate,
		Read:   resourceAwsAccessAnalyzerAnalyzerRead,
		Update: resourceAwsAccessAnalyzerAnalyzerUpdate,
		Delete: resourceAwsAccessAnalyzerAnalyzerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"analyzer_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]*$`), "must begin with a letter and contain only alphanumeric, underscore, period, or hyphen characters"),
				),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tagsSchema(),
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  accessanalyzer.TypeAccount,
				ValidateFunc: validation.StringInSlice([]string{
					accessanalyzer.TypeAccount,
					accessanalyzer.TypeOrganization,
				}, false),
			},
		},
	}
}

func resourceAwsAccessAnalyzerAnalyzerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).accessanalyzerconn
	analyzerName := d.Get("analyzer_name").(string)

	input := &accessanalyzer.CreateAnalyzerInput{
		AnalyzerName: aws.String(analyzerName),
		ClientToken:  aws.String(resource.UniqueId()),
		Tags:         keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().AccessanalyzerTags(),
		Type:         aws.String(d.Get("type").(string)),
	}

	// Handle Organizations eventual consistency
	err := resource.Retry(accessAnalyzerOrganizationCreationTimeout, func() *resource.RetryError {
		_, err := conn.CreateAnalyzer(input)

		if isAWSErr(err, accessanalyzer.ErrCodeValidationException, "You must create an organization") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.CreateAnalyzer(input)
	}

	if err != nil {
		return fmt.Errorf("error creating Access Analyzer Analyzer (%s): %s", analyzerName, err)
	}

	d.SetId(analyzerName)

	return resourceAwsAccessAnalyzerAnalyzerRead(d, meta)
}

func resourceAwsAccessAnalyzerAnalyzerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).accessanalyzerconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &accessanalyzer.GetAnalyzerInput{
		AnalyzerName: aws.String(d.Id()),
	}

	output, err := conn.GetAnalyzer(input)

	if isAWSErr(err, accessanalyzer.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Access Analyzer Analyzer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Access Analyzer Analyzer (%s): %s", d.Id(), err)
	}

	if output == nil || output.Analyzer == nil {
		return fmt.Errorf("error getting Access Analyzer Analyzer (%s): empty response", d.Id())
	}

	d.Set("analyzer_name", output.Analyzer.Name)
	d.Set("arn", output.Analyzer.Arn)

	if err := d.Set("tags", keyvaluetags.AccessanalyzerKeyValueTags(output.Analyzer.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("type", output.Analyzer.Type)

	return nil
}

func resourceAwsAccessAnalyzerAnalyzerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).accessanalyzerconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.AccessanalyzerUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Access Analyzer Analyzer (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsAccessAnalyzerAnalyzerRead(d, meta)
}

func resourceAwsAccessAnalyzerAnalyzerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).accessanalyzerconn

	input := &accessanalyzer.DeleteAnalyzerInput{
		AnalyzerName: aws.String(d.Id()),
		ClientToken:  aws.String(resource.UniqueId()),
	}

	_, err := conn.DeleteAnalyzer(input)

	if isAWSErr(err, accessanalyzer.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Access Analyzer Analyzer (%s): %s", d.Id(), err)
	}

	return nil
}
