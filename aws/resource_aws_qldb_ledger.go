package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qldb"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsQLDBLedger() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsQLDBLedgerCreate,
		Read:   resourceAwsQLDBLedgerRead,
		Update: resourceAwsQLDBLedgerUpdate,
		Delete: resourceAwsQLDBLedgerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`^[A-Za-z0-9_-]+`), "must contain only alphanumeric characters, underscores, and hyphens"),
				),
			},

			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsQLDBLedgerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).qldbconn

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else {
		name = resource.PrefixedUniqueId("tf")
	}

	if err := d.Set("name", name); err != nil {
		return fmt.Errorf("error setting name: %s", err)
	}

	// Create the QLDB Ledger
	// The qldb.PermissionsModeAllowAll is currently hardcoded because AWS doesn't support changing the mode.
	createOpts := &qldb.CreateLedgerInput{
		Name:               aws.String(d.Get("name").(string)),
		PermissionsMode:    aws.String(qldb.PermissionsModeAllowAll),
		DeletionProtection: aws.Bool(d.Get("deletion_protection").(bool)),
		Tags:               keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().QldbTags(),
	}

	log.Printf("[DEBUG] QLDB Ledger create config: %#v", *createOpts)
	qldbResp, err := conn.CreateLedger(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating QLDB Ledger: %s", err)
	}

	// Set QLDB ledger name
	d.SetId(*qldbResp.Name)

	log.Printf("[INFO] QLDB Ledger name: %s", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending:    []string{qldb.LedgerStateCreating},
		Target:     []string{qldb.LedgerStateActive},
		Refresh:    qldbLedgerRefreshStatusFunc(conn, d.Id()),
		Timeout:    8 * time.Minute,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for QLDB Ledger status to be \"%s\": %s", qldb.LedgerStateActive, err)
	}

	// Update our attributes and return
	return resourceAwsQLDBLedgerRead(d, meta)
}

func resourceAwsQLDBLedgerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).qldbconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	// Refresh the QLDB state
	input := &qldb.DescribeLedgerInput{
		Name: aws.String(d.Id()),
	}

	qldbLedger, err := conn.DescribeLedger(input)

	if isAWSErr(err, qldb.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] QLDB Ledger (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing QLDB Ledger (%s): %s", d.Id(), err)
	}

	// QLDB stuff
	if err := d.Set("name", qldbLedger.Name); err != nil {
		return fmt.Errorf("error setting name: %s", err)
	}

	if err := d.Set("deletion_protection", qldbLedger.DeletionProtection); err != nil {
		return fmt.Errorf("error setting deletion protection: %s", err)
	}

	// ARN
	if err := d.Set("arn", qldbLedger.Arn); err != nil {
		return fmt.Errorf("error setting ARN: %s", err)
	}

	// Tags
	log.Printf("[INFO] Fetching tags for %s", d.Id())
	tags, err := keyvaluetags.QldbListTags(conn, d.Get("arn").(string))
	if err != nil {
		return fmt.Errorf("Error listing tags for QLDB Ledger: %s", err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsQLDBLedgerUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).qldbconn

	if d.HasChange("deletion_protection") {
		val := d.Get("deletion_protection").(bool)
		modifyOpts := &qldb.UpdateLedgerInput{
			Name:               aws.String(d.Id()),
			DeletionProtection: aws.Bool(val),
		}
		log.Printf(
			"[INFO] Modifying deletion_protection QLDB attribute for %s: %#v",
			d.Id(), modifyOpts)
		if _, err := conn.UpdateLedger(modifyOpts); err != nil {

			return err
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.QldbUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsQLDBLedgerRead(d, meta)
}

func resourceAwsQLDBLedgerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).qldbconn
	deleteLedgerOpts := &qldb.DeleteLedgerInput{
		Name: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting QLDB Ledger: %s", d.Id())

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteLedger(deleteLedgerOpts)

		if isAWSErr(err, qldb.ErrCodeResourceInUseException, "") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteLedger(deleteLedgerOpts)
	}

	if isAWSErr(err, qldb.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting QLDB Ledger (%s): %s", d.Id(), err)
	}

	if err := waitForQLDBLedgerDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for QLDB Ledger (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func qldbLedgerRefreshStatusFunc(conn *qldb.QLDB, ledger string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &qldb.DescribeLedgerInput{
			Name: aws.String(ledger),
		}
		resp, err := conn.DescribeLedger(input)
		if err != nil {
			return nil, "failed", err
		}
		return resp, aws.StringValue(resp.State), nil
	}
}

func waitForQLDBLedgerDeletion(conn *qldb.QLDB, ledgerName string) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{qldb.LedgerStateCreating,
			qldb.LedgerStateActive,
			qldb.LedgerStateDeleting},
		Target:     []string{""},
		Timeout:    5 * time.Minute,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeLedger(&qldb.DescribeLedgerInput{
				Name: aws.String(ledgerName),
			})

			if isAWSErr(err, qldb.ErrCodeResourceNotFoundException, "") {
				return 1, "", nil
			}

			if err != nil {
				return nil, qldb.ErrCodeResourceInUseException, err
			}

			return resp, aws.StringValue(resp.State), nil
		},
	}

	_, err := stateConf.WaitForState()

	return err
}
