package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	route53ResolverRuleAssociationStatusDeleted = "DELETED"
)

func resourceAwsRoute53ResolverRuleAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53ResolverRuleAssociationCreate,
		Read:   resourceAwsRoute53ResolverRuleAssociationRead,
		Delete: resourceAwsRoute53ResolverRuleAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"resolver_rule_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},

			"vpc_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},

			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateRoute53ResolverName,
			},
		},
	}
}

func resourceAwsRoute53ResolverRuleAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	req := &route53resolver.AssociateResolverRuleInput{
		ResolverRuleId: aws.String(d.Get("resolver_rule_id").(string)),
		VPCId:          aws.String(d.Get("vpc_id").(string)),
	}
	if v, ok := d.GetOk("name"); ok {
		req.Name = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Route 53 Resolver rule association: %s", req)
	resp, err := conn.AssociateResolverRule(req)
	if err != nil {
		return fmt.Errorf("error creating Route 53 Resolver rule association: %s", err)
	}

	d.SetId(aws.StringValue(resp.ResolverRuleAssociation.Id))

	err = route53ResolverRuleAssociationWaitUntilTargetState(conn, d.Id(), d.Timeout(schema.TimeoutCreate),
		[]string{route53resolver.ResolverRuleAssociationStatusCreating},
		[]string{route53resolver.ResolverRuleAssociationStatusComplete})
	if err != nil {
		return err
	}

	return resourceAwsRoute53ResolverRuleAssociationRead(d, meta)
}

func resourceAwsRoute53ResolverRuleAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	assocRaw, state, err := route53ResolverRuleAssociationRefresh(conn, d.Id())()
	if err != nil {
		return fmt.Errorf("error getting Route53 Resolver rule association (%s): %s", d.Id(), err)
	}
	if state == route53ResolverRuleAssociationStatusDeleted {
		log.Printf("[WARN] Route53 Resolver rule association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	assoc := assocRaw.(*route53resolver.ResolverRuleAssociation)

	d.Set("name", assoc.Name)
	d.Set("resolver_rule_id", assoc.ResolverRuleId)
	d.Set("vpc_id", assoc.VPCId)

	return nil
}

func resourceAwsRoute53ResolverRuleAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).route53resolverconn

	log.Printf("[DEBUG] Deleting Route53 Resolver rule association: %s", d.Id())
	_, err := conn.DisassociateResolverRule(&route53resolver.DisassociateResolverRuleInput{
		ResolverRuleId: aws.String(d.Get("resolver_rule_id").(string)),
		VPCId:          aws.String(d.Get("vpc_id").(string)),
	})
	if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting Route 53 Resolver rule association (%s): %s", d.Id(), err)
	}

	err = route53ResolverRuleAssociationWaitUntilTargetState(conn, d.Id(), d.Timeout(schema.TimeoutDelete),
		[]string{route53resolver.ResolverRuleAssociationStatusDeleting},
		[]string{route53ResolverRuleAssociationStatusDeleted})
	if err != nil {
		return err
	}

	return nil
}

func route53ResolverRuleAssociationRefresh(conn *route53resolver.Route53Resolver, assocId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.GetResolverRuleAssociation(&route53resolver.GetResolverRuleAssociationInput{
			ResolverRuleAssociationId: aws.String(assocId),
		})
		if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
			return "", route53ResolverRuleAssociationStatusDeleted, nil
		}
		if err != nil {
			return nil, "", err
		}

		if statusMessage := aws.StringValue(resp.ResolverRuleAssociation.StatusMessage); statusMessage != "" {
			log.Printf("[INFO] Route 53 Resolver rule association (%s) status message: %s", assocId, statusMessage)
		}

		return resp.ResolverRuleAssociation, aws.StringValue(resp.ResolverRuleAssociation.Status), nil
	}
}

func route53ResolverRuleAssociationWaitUntilTargetState(conn *route53resolver.Route53Resolver, assocId string, timeout time.Duration, pending, target []string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    route53ResolverRuleAssociationRefresh(conn, assocId),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error waiting for Route53 Resolver rule association (%s) to reach target state: %s", assocId, err)
	}

	return nil
}
