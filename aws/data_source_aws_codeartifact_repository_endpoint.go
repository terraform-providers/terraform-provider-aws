package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsCodeArtifactRepositoryEndpoint() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCodeArtifactRepositoryEndpointRead,

		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"repository": {
				Type:     schema.TypeString,
				Required: true,
			},
			"format": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(codeartifact.PackageFormat_Values(), false),
			},
			"domain_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateAwsAccountId,
			},
			"repository_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsCodeArtifactRepositoryEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).codeartifactconn
	domainOwner := meta.(*AWSClient).accountid
	domain := d.Get("domain").(string)
	repo := d.Get("repository").(string)
	format := d.Get("format").(string)
	params := &codeartifact.GetRepositoryEndpointInput{
		Domain:     aws.String(domain),
		Repository: aws.String(repo),
		Format:     aws.String(format),
	}

	if v, ok := d.GetOk("domain_owner"); ok {
		params.DomainOwner = aws.String(v.(string))
		domainOwner = v.(string)
	}

	log.Printf("[DEBUG] Getting CodeArtifact Repository Endpoint")
	out, err := conn.GetRepositoryEndpoint(params)
	if err != nil {
		return fmt.Errorf("error getting CodeArtifact Repository Endpoint: %w", err)
	}
	log.Printf("[DEBUG] CodeArtifact Repository Endpoint: %#v", out)

	d.SetId(fmt.Sprintf("%s:%s:%s:%s", domainOwner, domain, repo, format))
	d.Set("repository_endpoint", out.RepositoryEndpoint)
	d.Set("domain_owner", domainOwner)

	return nil
}
