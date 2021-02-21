package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsCodeArtifactAuthorizationToken() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCodeArtifactAuthorizationTokenRead,

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"domain_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateAwsAccountId,
			},
			"duration_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
				ValidateFunc: validation.Any(
					validation.IntBetween(900, 43200),
					validation.IntInSlice([]int{0}),
				),
			},
			"authorization_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"expiration": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsCodeArtifactAuthorizationTokenRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn
	domain := d.Get("domain").(string)
	domainOwner := meta.(*AWSClient).accountid
	params := &codeartifact.GetAuthorizationTokenInput{
		Domain: aws.String(domain),
	}

	if v, ok := d.GetOk("domain_owner"); ok {
		params.DomainOwner = aws.String(v.(string))
		domainOwner = v.(string)
	}

	if v, ok := d.GetOkExists("duration_seconds"); ok {
		params.DurationSeconds = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Getting CodeArtifact authorization token")
	out, err := conn.GetAuthorizationToken(params)
	if err != nil {
		return fmt.Errorf("error getting CodeArtifact authorization token: %w", err)
	}
	log.Printf("[DEBUG] CodeArtifact authorization token: %#v", out)

	d.SetId(fmt.Sprintf("%s:%s", domainOwner, domain))
	d.Set("authorization_token", aws.StringValue(out.AuthorizationToken))
	d.Set("expiration", aws.TimeValue(out.Expiration).Format(time.RFC3339))
	d.Set("domain_owner", domainOwner)

	return nil
}
