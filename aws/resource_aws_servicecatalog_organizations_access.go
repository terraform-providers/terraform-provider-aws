package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/servicecatalog"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicecatalog/waiter"
)

func resourceAwsServiceCatalogOrganizationsAccess() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceCatalogOrganizationsAccessCreate,
		Read:   resourceAwsServiceCatalogOrganizationsAccessRead,
		Delete: resourceAwsServiceCatalogOrganizationsAccessDelete,

		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsServiceCatalogOrganizationsAccessCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	d.SetId(meta.(*AWSClient).accountid)

	// During create, if enabled = "true", then Enable Access and vice versa
	// During delete, the opposite

	if _, ok := d.GetOk("enabled"); ok {
		_, err := conn.EnableAWSOrganizationsAccess(&servicecatalog.EnableAWSOrganizationsAccessInput{})

		if err != nil {
			return fmt.Errorf("error enabling Service Catalog AWS Organizations Access: %w", err)
		}

		return resourceAwsServiceCatalogOrganizationsAccessRead(d, meta)
	}

	_, err := conn.DisableAWSOrganizationsAccess(&servicecatalog.DisableAWSOrganizationsAccessInput{})

	if err != nil {
		return fmt.Errorf("error disabling Service Catalog AWS Organizations Access: %w", err)
	}

	return resourceAwsServiceCatalogOrganizationsAccessRead(d, meta)
}

func resourceAwsServiceCatalogOrganizationsAccessRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	output, err := waiter.OrganizationsAccessStable(conn)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, servicecatalog.ErrCodeResourceNotFoundException) {
		// theoretically this should not be possible
		log.Printf("[WARN] Service Catalog Organizations Access (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing Service Catalog AWS Organizations Access (%s): %w", d.Id(), err)
	}

	if output == "" {
		return fmt.Errorf("error getting Service Catalog AWS Organizations Access (%s): empty response", d.Id())
	}

	if output == servicecatalog.AccessStatusEnabled {
		d.Set("enabled", true)
		return nil
	}

	d.Set("enabled", false)
	return nil
}

func resourceAwsServiceCatalogOrganizationsAccessDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).scconn

	// During create, if enabled = "true", then Enable Access and vice versa
	// During delete, the opposite

	if _, ok := d.GetOk("enabled"); !ok {
		_, err := conn.EnableAWSOrganizationsAccess(&servicecatalog.EnableAWSOrganizationsAccessInput{})

		if err != nil {
			return fmt.Errorf("error enabling Service Catalog AWS Organizations Access: %w", err)
		}

		return nil
	}

	_, err := conn.DisableAWSOrganizationsAccess(&servicecatalog.DisableAWSOrganizationsAccessInput{})

	if err != nil {
		return fmt.Errorf("error disabling Service Catalog AWS Organizations Access: %w", err)
	}

	return nil
}
