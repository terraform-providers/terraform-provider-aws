package aws

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/identitystore"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceAwsIdentityStoreUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsIdentityStoreUserRead,

		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute_path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"attribute_value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},

			"identity_store_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]*$`), "must match [a-zA-Z0-9-]"),
				),
			},

			"user_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 47),
					validation.StringMatch(regexp.MustCompile(`^([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}$`), "must match ([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}"),
				),
			},

			"user_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsIdentityStoreUserRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).identitystoreconn

	input := &identitystore.ListUsersInput{
		IdentityStoreId: aws.String(d.Get("identity_store_id").(string)),
		Filters:         expandIdentityStoreFilters(d.Get("filter").(*schema.Set).List()),
	}

	var results []*identitystore.User

	err := conn.ListUsersPages(input, func(page *identitystore.ListUsersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, user := range page.Users {
			if user == nil {
				continue
			}

			if v, ok := d.GetOk("user_id"); ok && v.(string) != aws.StringValue(user.UserId) {
				continue
			}

			results = append(results, user)
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing Identity Store Users: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("no Identity Store User found matching criteria; try different search")
	}

	if len(results) > 1 {
		return fmt.Errorf("multiple Identity Store Users found matching criteria; try different search")
	}

	user := results[0]

	d.SetId(aws.StringValue(user.UserId))
	d.Set("user_id", user.UserId)
	d.Set("user_name", user.UserName)

	return nil
}
