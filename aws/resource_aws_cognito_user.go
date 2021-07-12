package aws

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsCognitoUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsCognitoUserCreate,
		Read:   resourceAwsCognitoUserRead,
		Update: resourceAwsCognitoUserUpdate,
		Delete: resourceAwsCognitoUserDelete,

		Importer: &schema.ResourceImporter{
			State: resourceAwsCognitoUserImport,
		},

		// https://docs.aws.amazon.com/cognito-user-identity-pools/latest/APIReference/API_AdminCreateUser.html
		Schema: map[string]*schema.Schema{
			"user_attribute": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
					},
				},
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsCognitoUserCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.AdminCreateUserInput{
		Username:   aws.String(d.Get("username").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	if v, ok := d.GetOk("user_attribute"); ok {
		attributes := v.(*schema.Set)
		params.UserAttributes = expandCognitoUserAttributes(attributes)
	}

	log.Print("[DEBUG] Creating Cognito User")

	resp, err := conn.AdminCreateUser(params)
	if err != nil {
		return fmt.Errorf("Error creating Cognito User: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s", *params.UserPoolId, *resp.User.Username))

	return resourceAwsCognitoUserRead(d, meta)
}

func resourceAwsCognitoUserRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	log.Println("[DEBUG] Creating request struct")
	params := &cognitoidentityprovider.AdminGetUserInput{
		Username:   aws.String(d.Get("username").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}
	log.Println("[DEBUG] Request input: ", params)
	log.Println("[DEBUG] Reading Cognito User")

	user, err := conn.AdminGetUser(params)
	if err != nil {
		log.Println("[ERROR] Error reading Cognito User: ", err)
		if isAWSErr(err, "ResourceNotFoundException", "") {
			log.Printf("[WARN] Cognito User %s is already gone", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading Cognito User: %s", err)
	}

	if err := d.Set("user_attribute", flattenCognitoUserAttributes(user.UserAttributes)); err != nil {
		return fmt.Errorf("failed setting user_attributes: %w", err)
	}

	return nil
}

func resourceAwsCognitoUserUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	log.Println("[DEBUG] Updating Cognito User")

	if d.HasChange("user_attribute") {
		o, n := d.GetChange("user_attribute")

		upd, del := computeCognitoUserAttributesUpdate(o, n)

		if upd.Len() > 0 {
			params := &cognitoidentityprovider.AdminUpdateUserAttributesInput{
				Username:       aws.String(d.Get("username").(string)),
				UserPoolId:     aws.String(d.Get("user_pool_id").(string)),
				UserAttributes: expandCognitoUserAttributes(upd),
			}
			_, err := conn.AdminUpdateUserAttributes(params)
			if err != nil {
				if isAWSErr(err, "ResourceNotFoundException", "") {
					log.Printf("[WARN] Cognito User %s is already gone", d.Id())
					d.SetId("")
					return nil
				}
				return fmt.Errorf("Error updating Cognito User Attributes: %s", err)
			}
		}
		if len(del) > 0 {
			params := &cognitoidentityprovider.AdminDeleteUserAttributesInput{
				Username:           aws.String(d.Get("username").(string)),
				UserPoolId:         aws.String(d.Get("user_pool_id").(string)),
				UserAttributeNames: del,
			}
			_, err := conn.AdminDeleteUserAttributes(params)
			if err != nil {
				if isAWSErr(err, "ResourceNotFoundException", "") {
					log.Printf("[WARN] Cognito User %s is already gone", d.Id())
					d.SetId("")
					return nil
				}
				return fmt.Errorf("Error updating Cognito User Attributes: %s", err)
			}
		}
	}

	return resourceAwsCognitoUserRead(d, meta)
}

func resourceAwsCognitoUserDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn

	params := &cognitoidentityprovider.AdminDeleteUserInput{
		Username:   aws.String(d.Get("username").(string)),
		UserPoolId: aws.String(d.Get("user_pool_id").(string)),
	}

	log.Print("[DEBUG] Deleting Cognito User")

	_, err := conn.AdminDeleteUser(params)
	if err != nil {
		return fmt.Errorf("Error deleting Cognito User: %s", err)
	}

	return nil
}

func resourceAwsCognitoUserImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idSplit := strings.Split(d.Id(), "/")
	if len(idSplit) != 2 {
		return nil, errors.New("Error importing Cognito User. Must specify user_pool_id/username")
	}
	userPoolId := idSplit[0]
	name := idSplit[1]
	d.Set("user_pool_id", userPoolId)
	d.Set("username", name)
	return []*schema.ResourceData{d}, nil
}

func expandCognitoUserAttributes(tfSet *schema.Set) []*cognitoidentityprovider.AttributeType {
	if tfSet.Len() == 0 {
		return nil
	}

	apiList := make([]*cognitoidentityprovider.AttributeType, 0, tfSet.Len())

	for _, tfAttribute := range tfSet.List() {
		apiAttribute := tfAttribute.(map[string]interface{})
		apiList = append(apiList, &cognitoidentityprovider.AttributeType{
			Name:  aws.String(apiAttribute["name"].(string)),
			Value: aws.String(apiAttribute["value"].(string)),
		})
	}

	return apiList
}

func flattenCognitoUserAttributes(apiList []*cognitoidentityprovider.AttributeType) *schema.Set {
	if len(apiList) == 1 {
		return nil
	}

	tfList := []interface{}{}

	for _, apiAttribute := range apiList {
		// not sure if this is the best way to deal with this system attrubute
		if *apiAttribute.Name == "sub" {
			continue
		}

		tfAttribute := map[string]interface{}{}

		if apiAttribute.Name != nil {
			tfAttribute["name"] = aws.StringValue(apiAttribute.Name)
		}

		if apiAttribute.Value != nil {
			tfAttribute["value"] = aws.StringValue(apiAttribute.Value)
		}

		tfList = append(tfList, tfAttribute)
	}

	tfSet := schema.NewSet(cognitoUserAttributeHash, tfList)

	return tfSet
}

// computeCognitoUserAttributesUpdate computes which userattributes should be updated and which ones should be deleted.
// We should do it like this because we cannot explicitly set a list of user attributes in cognito. We can either perform
// an update or delete operation.
func computeCognitoUserAttributesUpdate(old interface{}, new interface{}) (*schema.Set, []*string) {
	oldMap := map[string]interface{}{}

	oldList := old.(*schema.Set).List()
	newList := new.(*schema.Set).List()

	upd := schema.NewSet(cognitoUserAttributeHash, []interface{}{})
	del := []*string{}

	for _, v := range oldList {
		vMap := v.(map[string]interface{})
		oldMap[vMap["name"].(string)] = vMap["value"]
	}

	for _, v := range newList {
		vMap := v.(map[string]interface{})
		if oldV, ok := oldMap[vMap["name"].(string)]; ok {
			if oldV != vMap["value"] {
				upd.Add(map[string]interface{}{
					"name":  vMap["name"].(string),
					"value": vMap["value"],
				})
			}
			delete(oldMap, vMap["name"].(string))
		} else {
			upd.Add(map[string]interface{}{
				"name":  vMap["name"].(string),
				"value": vMap["value"],
			})
		}
	}

	for k := range oldMap {
		del = append(del, &k)
	}

	return upd, del
}

func cognitoUserAttributeHash(attr interface{}) int {
	attrMap := attr.(map[string]interface{})

	return schema.HashString(attrMap["name"])
}
