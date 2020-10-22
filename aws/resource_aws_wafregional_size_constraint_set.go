package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsWafRegionalSizeConstraintSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafRegionalSizeConstraintSetCreate,
		Read:   resourceAwsWafRegionalSizeConstraintSetRead,
		Update: resourceAwsWafRegionalSizeConstraintSetUpdate,
		Delete: resourceAwsWafRegionalSizeConstraintSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: wafSizeConstraintSetSchema(),
	}
}

func resourceAwsWafRegionalSizeConstraintSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	region := meta.(*AWSClient).region

	name := d.Get("name").(string)

	log.Printf("[INFO] Creating WAF Regional SizeConstraintSet: %s", name)

	wr := newWafRegionalRetryer(conn, region)
	out, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		params := &waf.CreateSizeConstraintSetInput{
			ChangeToken: token,
			Name:        aws.String(name),
		}

		return conn.CreateSizeConstraintSet(params)
	})
	if err != nil {
		return fmt.Errorf("Error creating WAF Regional SizeConstraintSet: %s", err)
	}
	resp := out.(*waf.CreateSizeConstraintSetOutput)

	d.SetId(aws.StringValue(resp.SizeConstraintSet.SizeConstraintSetId))

	return resourceAwsWafRegionalSizeConstraintSetUpdate(d, meta)
}

func resourceAwsWafRegionalSizeConstraintSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn

	log.Printf("[INFO] Reading WAF Regional SizeConstraintSet: %s", d.Get("name").(string))
	params := &waf.GetSizeConstraintSetInput{
		SizeConstraintSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetSizeConstraintSet(params)
	if isAWSErr(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
		log.Printf("[WARN] WAF Regional SizeConstraintSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error getting WAF Regional Size Constraint Set (%s): %s", d.Id(), err)
	}

	d.Set("name", resp.SizeConstraintSet.Name)
	d.Set("size_constraints", flattenWafSizeConstraints(resp.SizeConstraintSet.SizeConstraints))

	return nil
}

func resourceAwsWafRegionalSizeConstraintSetUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient)

	if d.HasChange("size_constraints") {
		o, n := d.GetChange("size_constraints")
		oldConstraints, newConstraints := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateRegionalSizeConstraintSetResource(d.Id(), oldConstraints, newConstraints, client.wafregionalconn, client.region)
		if isAWSErr(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
			log.Printf("[WARN] WAF Regional SizeConstraintSet (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		if err != nil {
			return fmt.Errorf("Error updating WAF Regional SizeConstraintSet(%s): %s", d.Id(), err)
		}
	}

	return resourceAwsWafRegionalSizeConstraintSetRead(d, meta)
}

func resourceAwsWafRegionalSizeConstraintSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafregionalconn
	region := meta.(*AWSClient).region

	oldConstraints := d.Get("size_constraints").(*schema.Set).List()

	if len(oldConstraints) > 0 {
		noConstraints := []interface{}{}
		err := updateRegionalSizeConstraintSetResource(d.Id(), oldConstraints, noConstraints, conn, region)
		if isAWSErr(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
			return nil
		}
		if err != nil {
			return fmt.Errorf("Error deleting WAF Regional SizeConstraintSet(%s): %s", d.Id(), err)
		}
	}

	wr := newWafRegionalRetryer(conn, region)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.DeleteSizeConstraintSetInput{
			ChangeToken:         token,
			SizeConstraintSetId: aws.String(d.Id()),
		}
		return conn.DeleteSizeConstraintSet(req)
	})
	if isAWSErr(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting WAF Regional SizeConstraintSet: %s", err)
	}

	return nil
}

func updateRegionalSizeConstraintSetResource(id string, oldConstraints, newConstraints []interface{}, conn *wafregional.WAFRegional, region string) error {
	wr := newWafRegionalRetryer(conn, region)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.UpdateSizeConstraintSetInput{
			ChangeToken:         token,
			SizeConstraintSetId: aws.String(id),
			Updates:             diffWafSizeConstraints(oldConstraints, newConstraints),
		}

		log.Printf("[INFO] Updating WAF Regional SizeConstraintSet: %s", req)
		return conn.UpdateSizeConstraintSet(req)
	})

	return err
}
