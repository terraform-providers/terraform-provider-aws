package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsCognitoUserPoolClient() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCognitoUserPoolClientRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_pool_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"client_id": {
				Type:     schema.TypeString,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"client_secret": {
				Type:     schema.TypeString,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAwsCognitoUserPoolClientRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).cognitoidpconn
	name := d.Get("name").(string)
	userPoolId := d.Get("user_pool_id").(string)

	client, err := getCognitoUserPoolClient(conn, userPoolId, name)
	if err != nil {
		return fmt.Errorf("Error getting cognito user pool client: %s", err)
	}
	if client == nil {
		return fmt.Errorf("No cognito user pool client found with name: %s", name)
	}

	d.SetId(*client.ClientId)
	d.Set("client_id", client.ClientId)
	d.Set("client_secret", client.ClientSecret)

	return nil
}

func getCognitoUserPoolClient(conn *cognitoidentityprovider.CognitoIdentityProvider, userPoolId string, userPoolClientName string) (*cognitoidentityprovider.UserPoolClientType, error) {
	var nextToken string

	for {
		input := &cognitoidentityprovider.ListUserPoolClientsInput{
			// MaxResults Valid Range: Minimum value of 1. Maximum value of 60
			MaxResults: aws.Int64(int64(60)),
			UserPoolId: aws.String(userPoolId),
		}
		if nextToken != "" {
			input.NextToken = aws.String(nextToken)
		}
		out, err := conn.ListUserPoolClients(input)
		if err != nil {
			return nil, err
		}

		for _, client := range out.UserPoolClients {

			clientName := aws.StringValue(client.ClientName)

			if clientName == userPoolClientName {

				input := &cognitoidentityprovider.DescribeUserPoolClientInput{
					ClientId:   aws.String(*client.ClientId),
					UserPoolId: aws.String(userPoolId),
				}

				describeClientResponse, err := conn.DescribeUserPoolClient(input)

				if err != nil {
					return nil, err
				}

				return describeClientResponse.UserPoolClient, nil
			}
		}

		if out.NextToken == nil {
			break
		}
		nextToken = aws.StringValue(out.NextToken)
	}

	return nil, nil
}
