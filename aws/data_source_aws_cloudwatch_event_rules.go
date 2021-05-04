package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsCloudwatchEventRules() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsCloudwatchEventRulesRead,

		Schema: map[string]*schema.Schema{
			"name_prefix": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceAwsCloudwatchEventRulesRead(d *schema.ResourceData, meta interface{}) error {
	namePrefix := d.Get("name_prefix").(string)
	conn := meta.(*AWSClient).cloudwatcheventsconn

	ruleNames := []string{}
	var nextToken string
	for {
		input := cloudwatchevents.ListRulesInput{NamePrefix: aws.String(namePrefix)}
		if nextToken != "" {
			input.SetNextToken(nextToken)
		}

		resp, err := conn.ListRules(&input)
		if err != nil {
			return fmt.Errorf("error listing CloudWatch Event Rules with prefix (%s): %s", namePrefix, err)
		}

		for _, rule := range resp.Rules {
			ruleNames = append(ruleNames, aws.StringValue(rule.Name))
		}

		if resp.NextToken != nil {
			nextToken = *resp.NextToken
		} else {
			break
		}
	}

	d.SetId(namePrefix)
	d.Set("rule_names", ruleNames)

	return nil
}
