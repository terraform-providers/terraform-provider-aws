package aws

import ( // nosemgrep: aws-sdk-go-multiple-service-imports
	"fmt"
	"log"
	"path/filepath" // filepath for glob matching

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAwsIAMRoles() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsIAMRolesRead,

		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"path_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
	}
}

func dataSourceAwsIAMRolesRead(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn
	req := &iam.ListRolesInput{}

	if pathPrefix, hasPathPrefix := d.GetOk("path_prefix"); hasPathPrefix {
		req.PathPrefix = aws.String(pathPrefix.(string))
	}

	filters, hasFilters := d.GetOk("filter")
	filtersSet := []*ec2.Filter{}

	if hasFilters {
		filtersSet = buildAwsDataSourceFilters(filters.(*schema.Set))
		log.Printf("[DEBUG] Has filters : %s", filtersSet)
		// Only filters using the name "role-name" are currently supported
		for _, f := range filtersSet {
			if "role-name" != aws.StringValue(f.Name) {
				return fmt.Errorf("Provided filters does not match supported names. See the documentation of this data source for supported filters.")
			}
		}
	} else {
		log.Printf("[DEBUG] No filter")
	}

	roles := []*iam.Role{}

	err := iamconn.ListRolesPages(
		req,
		func(page *iam.ListRolesOutput, lastPage bool) bool {
			for _, role := range page.Roles {
				if hasFilters {
					log.Printf("[DEBUG] Found Role '%s' to be checked against filters", *role.RoleName)
					matchAllFilters := true
					for _, f := range filtersSet {
						for _, filterValue := range f.Values {
							// must match all values
							if matched, _ := filepath.Match(*filterValue, *role.RoleName); !matched {
								log.Printf("[DEBUG] RoleName '%s' does not match filter '%s'", *role.RoleName, *filterValue)
								matchAllFilters = false
							}
						}
					}
					if matchAllFilters {
						roles = append(roles, role)
					}
				} else {
					log.Printf("[DEBUG] Found Role '%s'", *role.RoleName)
					roles = append(roles, role)
				}
			}
			return !lastPage
		},
	)
	if err != nil {
		return fmt.Errorf("error reading IAM roles : %w", err)
	}

	if len(roles) == 0 {
		log.Printf("[WARN] couldn't find any IAM role matching the provided parameters")
	}

	arns := []string{}
	names := []string{}
	for _, v := range roles {
		arns = append(arns, aws.StringValue(v.Arn))
		names = append(names, aws.StringValue(v.RoleName))
	}

	if err := d.Set("arns", arns); err != nil {
		return fmt.Errorf("error setting arns: %w", err)
	}
	if err := d.Set("names", names); err != nil {
		return fmt.Errorf("error setting names: %w", err)
	}

	d.SetId(meta.(*AWSClient).region)

	return nil
}
