package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/budgets"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceAwsBudgetsBudget() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},
			"name": {
				Type:          schema.TypeString,
				Computed:      true,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"budget_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"limit_amount": {
				Type:     schema.TypeString,
				Required: true,
			},
			"limit_unit": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cost_types": {
				Type:     schema.TypeList,
				Computed: true,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"include_credit": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"include_discount": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"include_other_subscription": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"include_recurring": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"include_refund": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"include_subscription": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"include_support": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"include_tax": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"include_upfront": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"use_amortized": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"use_blended": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"time_period_start": {
				Type:     schema.TypeString,
				Required: true,
			},
			"time_period_end": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "2087-06-15_00:00",
			},
			"time_unit": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cost_filters": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"notification": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"comparison_operator": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								budgets.ComparisonOperatorEqualTo,
								budgets.ComparisonOperatorGreaterThan,
								budgets.ComparisonOperatorLessThan,
							}, false),
						},
						"threshold": {
							Type:     schema.TypeFloat,
							Required: true,
						},
						"threshold_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								budgets.ThresholdTypeAbsoluteValue,
								budgets.ThresholdTypePercentage,
							}, false),
						},
						"notification_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								budgets.NotificationTypeActual,
								budgets.NotificationTypeForecasted,
							}, false),
						},
						"subscriber_email_addresses": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subscriber_sns_topic_arns": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
		Create: resourceAwsBudgetsBudgetCreate,
		Read:   resourceAwsBudgetsBudgetRead,
		Update: resourceAwsBudgetsBudgetUpdate,
		Delete: resourceAwsBudgetsBudgetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

func resourceAwsBudgetsBudgetCreate(d *schema.ResourceData, meta interface{}) error {
	budget, err := expandBudgetsBudgetUnmarshal(d)
	if err != nil {
		return fmt.Errorf("failed unmarshalling budget: %v", err)
	}

	if v, ok := d.GetOk("name"); ok {
		budget.BudgetName = aws.String(v.(string))

	} else if v, ok := d.GetOk("name_prefix"); ok {
		budget.BudgetName = aws.String(resource.PrefixedUniqueId(v.(string)))

	} else {
		budget.BudgetName = aws.String(resource.UniqueId())
	}

	conn := meta.(*AWSClient).budgetconn
	var accountID string
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	} else {
		accountID = meta.(*AWSClient).accountid
	}

	_, err = conn.CreateBudget(&budgets.CreateBudgetInput{
		AccountId: aws.String(accountID),
		Budget:    budget,
	})
	if err != nil {
		return fmt.Errorf("create budget failed: %v", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", accountID, *budget.BudgetName))

	notificationsRaw := d.Get("notification").(*schema.Set).List()
	notifications, subscribers := expandBudgetNotificationsUnmarshal(notificationsRaw)

	err = resourceAwsBudgetsBudgetNotificationsCreate(notifications, subscribers, *budget.BudgetName, accountID, meta)

	if err != nil {
		return fmt.Errorf("error creating Budget (%s) Notifications: %s", d.Id(), err)
	}

	return resourceAwsBudgetsBudgetRead(d, meta)
}

func resourceAwsBudgetsBudgetNotificationsCreate(notifications []*budgets.Notification, subscribers [][]*budgets.Subscriber, budgetName string, accountID string, meta interface{}) error {
	conn := meta.(*AWSClient).budgetconn

	for i, notification := range notifications {
		subscribers := subscribers[i]
		if len(subscribers) == 0 {
			return fmt.Errorf("Notification must have at least one subscriber!")
		}
		_, err := conn.CreateNotification(&budgets.CreateNotificationInput{
			BudgetName:   aws.String(budgetName),
			AccountId:    aws.String(accountID),
			Notification: notification,
			Subscribers:  subscribers,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func expandBudgetNotificationsUnmarshal(notificationsRaw []interface{}) ([]*budgets.Notification, [][]*budgets.Subscriber) {

	notifications := make([]*budgets.Notification, len(notificationsRaw))
	subscribersForNotifications := make([][]*budgets.Subscriber, len(notificationsRaw))
	for i, notificationRaw := range notificationsRaw {
		notificationRaw := notificationRaw.(map[string]interface{})
		comparisonOperator := notificationRaw["comparison_operator"].(string)
		threshold := notificationRaw["threshold"].(float64)
		thresholdType := notificationRaw["threshold_type"].(string)
		notificationType := notificationRaw["notification_type"].(string)

		notifications[i] = &budgets.Notification{
			ComparisonOperator: aws.String(comparisonOperator),
			Threshold:          aws.Float64(threshold),
			ThresholdType:      aws.String(thresholdType),
			NotificationType:   aws.String(notificationType),
		}

		emailSubscribers := expandBudgetSubscribers(notificationRaw["subscriber_email_addresses"], budgets.SubscriptionTypeEmail)
		snsSubscribers := expandBudgetSubscribers(notificationRaw["subscriber_sns_topic_arns"], budgets.SubscriptionTypeSns)

		subscribersForNotifications[i] = append(emailSubscribers, snsSubscribers...)
	}
	return notifications, subscribersForNotifications
}

func expandBudgetSubscribers(rawList interface{}, subscriptionType string) []*budgets.Subscriber {
	result := make([]*budgets.Subscriber, 0)
	addrs := expandStringSet(rawList.(*schema.Set))
	for _, addr := range addrs {
		result = append(result, &budgets.Subscriber{
			SubscriptionType: aws.String(subscriptionType),
			Address:          addr,
		})
	}
	return result
}

func resourceAwsBudgetsBudgetRead(d *schema.ResourceData, meta interface{}) error {
	accountID, budgetName, err := decodeBudgetsBudgetID(d.Id())
	if err != nil {
		return err
	}

	conn := meta.(*AWSClient).budgetconn
	describeBudgetOutput, err := conn.DescribeBudget(&budgets.DescribeBudgetInput{
		BudgetName: aws.String(budgetName),
		AccountId:  aws.String(accountID),
	})
	if isAWSErr(err, budgets.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] Budget %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("describe budget failed: %v", err)
	}

	budget := describeBudgetOutput.Budget
	if budget == nil {
		log.Printf("[WARN] Budget %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("account_id", accountID)
	d.Set("budget_type", budget.BudgetType)

	if err := d.Set("cost_filters", convertCostFiltersToStringMap(budget.CostFilters)); err != nil {
		return fmt.Errorf("error setting cost_filters: %s", err)
	}

	if err := d.Set("cost_types", flattenBudgetsCostTypes(budget.CostTypes)); err != nil {
		return fmt.Errorf("error setting cost_types: %s %s", err, budget.CostTypes)
	}

	if budget.BudgetLimit != nil {
		d.Set("limit_amount", budget.BudgetLimit.Amount)
		d.Set("limit_unit", budget.BudgetLimit.Unit)
	}

	d.Set("name", budget.BudgetName)

	if budget.TimePeriod != nil {
		d.Set("time_period_end", aws.TimeValue(budget.TimePeriod.End).Format("2006-01-02_15:04"))
		d.Set("time_period_start", aws.TimeValue(budget.TimePeriod.Start).Format("2006-01-02_15:04"))
	}

	d.Set("time_unit", budget.TimeUnit)

	return resourceAwsBudgetsBudgetNotificationRead(d, meta)
}

func resourceAwsBudgetsBudgetNotificationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).budgetconn

	accountID, budgetName, err := decodeBudgetsBudgetID(d.Id())

	if err != nil {
		return fmt.Errorf("error decoding Budget (%s) ID: %s", d.Id(), err)
	}

	describeNotificationsForBudgetOutput, err := conn.DescribeNotificationsForBudget(&budgets.DescribeNotificationsForBudgetInput{
		BudgetName: aws.String(budgetName),
		AccountId:  aws.String(accountID),
	})

	if err != nil {
		return fmt.Errorf("error describing Budget (%s) Notifications: %s", d.Id(), err)
	}

	notifications := make([]map[string]interface{}, 0)

	for _, notificationOutput := range describeNotificationsForBudgetOutput.Notifications {
		notification := make(map[string]interface{})

		notification["comparison_operator"] = aws.StringValue(notificationOutput.ComparisonOperator)
		notification["threshold"] = aws.Float64Value(notificationOutput.Threshold)
		notification["notification_type"] = aws.StringValue(notificationOutput.NotificationType)

		if notificationOutput.ThresholdType == nil {
			// The AWS API doesn't seem to return a ThresholdType if it's set to PERCENTAGE
			// Set it manually to make behavior more predictable
			notification["threshold_type"] = budgets.ThresholdTypePercentage
		} else {
			notification["threshold_type"] = aws.StringValue(notificationOutput.ThresholdType)
		}

		subscribersOutput, err := conn.DescribeSubscribersForNotification(&budgets.DescribeSubscribersForNotificationInput{
			BudgetName:   aws.String(budgetName),
			AccountId:    aws.String(accountID),
			Notification: notificationOutput,
		})

		if err != nil {
			return fmt.Errorf("error describing Budget (%s) Notification Subscribers: %s", d.Id(), err)
		}

		snsSubscribers := make([]interface{}, 0)
		emailSubscribers := make([]interface{}, 0)

		for _, subscriberOutput := range subscribersOutput.Subscribers {
			if *subscriberOutput.SubscriptionType == budgets.SubscriptionTypeSns {
				snsSubscribers = append(snsSubscribers, *subscriberOutput.Address)
			} else if *subscriberOutput.SubscriptionType == budgets.SubscriptionTypeEmail {
				emailSubscribers = append(emailSubscribers, *subscriberOutput.Address)
			}
		}
		if len(snsSubscribers) > 0 {
			notification["subscriber_sns_topic_arns"] = schema.NewSet(schema.HashString, snsSubscribers)
		}
		if len(emailSubscribers) > 0 {
			notification["subscriber_email_addresses"] = schema.NewSet(schema.HashString, emailSubscribers)
		}
		notifications = append(notifications, notification)
	}

	if err := d.Set("notification", notifications); err != nil {
		return fmt.Errorf("error setting notification: %s %s", err, describeNotificationsForBudgetOutput.Notifications)
	}

	return nil
}

func resourceAwsBudgetsBudgetUpdate(d *schema.ResourceData, meta interface{}) error {
	accountID, _, err := decodeBudgetsBudgetID(d.Id())
	if err != nil {
		return err
	}

	conn := meta.(*AWSClient).budgetconn
	budget, err := expandBudgetsBudgetUnmarshal(d)
	if err != nil {
		return fmt.Errorf("could not create budget: %v", err)
	}

	_, err = conn.UpdateBudget(&budgets.UpdateBudgetInput{
		AccountId: aws.String(accountID),
		NewBudget: budget,
	})
	if err != nil {
		return fmt.Errorf("update budget failed: %v", err)
	}

	err = resourceAwsBudgetsBudgetNotificationsUpdate(d, meta)

	if err != nil {
		return fmt.Errorf("update budget notification failed: %v", err)
	}

	return resourceAwsBudgetsBudgetRead(d, meta)
}
func resourceAwsBudgetsBudgetNotificationsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).budgetconn
	accountID, budgetName, err := decodeBudgetsBudgetID(d.Id())

	if err != nil {
		return err
	}

	if d.HasChange("notification") {
		o, n := d.GetChange("notification")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		removeNotifications, _ := expandBudgetNotificationsUnmarshal(os.Difference(ns).List())
		addNotifications, addSubscribers := expandBudgetNotificationsUnmarshal(ns.Difference(os).List())

		for _, notification := range removeNotifications {
			input := &budgets.DeleteNotificationInput{
				AccountId:    aws.String(accountID),
				BudgetName:   aws.String(budgetName),
				Notification: notification,
			}

			_, err := conn.DeleteNotification(input)

			if err != nil {
				return fmt.Errorf("error deleting Budget (%s) Notification: %s", d.Id(), err)
			}
		}

		err = resourceAwsBudgetsBudgetNotificationsCreate(addNotifications, addSubscribers, budgetName, accountID, meta)

		if err != nil {
			return fmt.Errorf("error creating Budget (%s) Notifications: %s", d.Id(), err)
		}
	}

	return nil
}

func resourceAwsBudgetsBudgetDelete(d *schema.ResourceData, meta interface{}) error {
	accountID, budgetName, err := decodeBudgetsBudgetID(d.Id())
	if err != nil {
		return err
	}

	conn := meta.(*AWSClient).budgetconn
	_, err = conn.DeleteBudget(&budgets.DeleteBudgetInput{
		BudgetName: aws.String(budgetName),
		AccountId:  aws.String(accountID),
	})
	if err != nil {
		if isAWSErr(err, budgets.ErrCodeNotFoundException, "") {
			log.Printf("[INFO] budget %s could not be found. skipping delete.", d.Id())
			return nil
		}

		return fmt.Errorf("delete budget failed: %v", err)
	}

	return nil
}

func flattenBudgetsCostTypes(costTypes *budgets.CostTypes) []map[string]interface{} {
	if costTypes == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"include_credit":             aws.BoolValue(costTypes.IncludeCredit),
		"include_discount":           aws.BoolValue(costTypes.IncludeDiscount),
		"include_other_subscription": aws.BoolValue(costTypes.IncludeOtherSubscription),
		"include_recurring":          aws.BoolValue(costTypes.IncludeRecurring),
		"include_refund":             aws.BoolValue(costTypes.IncludeRefund),
		"include_subscription":       aws.BoolValue(costTypes.IncludeSubscription),
		"include_support":            aws.BoolValue(costTypes.IncludeSupport),
		"include_tax":                aws.BoolValue(costTypes.IncludeTax),
		"include_upfront":            aws.BoolValue(costTypes.IncludeUpfront),
		"use_amortized":              aws.BoolValue(costTypes.UseAmortized),
		"use_blended":                aws.BoolValue(costTypes.UseBlended),
	}
	return []map[string]interface{}{m}
}

func convertCostFiltersToStringMap(costFilters map[string][]*string) map[string]string {
	convertedCostFilters := make(map[string]string)
	for k, v := range costFilters {
		filterValues := make([]string, 0)
		for _, singleFilterValue := range v {
			filterValues = append(filterValues, *singleFilterValue)
		}

		convertedCostFilters[k] = strings.Join(filterValues, ",")
	}

	return convertedCostFilters
}

func expandBudgetsBudgetUnmarshal(d *schema.ResourceData) (*budgets.Budget, error) {
	budgetName := d.Get("name").(string)
	budgetType := d.Get("budget_type").(string)
	budgetLimitAmount := d.Get("limit_amount").(string)
	budgetLimitUnit := d.Get("limit_unit").(string)
	costTypes := expandBudgetsCostTypesUnmarshal(d.Get("cost_types").([]interface{}))
	budgetTimeUnit := d.Get("time_unit").(string)
	budgetCostFilters := make(map[string][]*string)
	for k, v := range d.Get("cost_filters").(map[string]interface{}) {
		filterValue := v.(string)
		budgetCostFilters[k] = append(budgetCostFilters[k], aws.String(filterValue))
	}

	budgetTimePeriodStart, err := time.Parse("2006-01-02_15:04", d.Get("time_period_start").(string))
	if err != nil {
		return nil, fmt.Errorf("failure parsing time: %v", err)
	}

	budgetTimePeriodEnd, err := time.Parse("2006-01-02_15:04", d.Get("time_period_end").(string))
	if err != nil {
		return nil, fmt.Errorf("failure parsing time: %v", err)
	}

	budget := &budgets.Budget{
		BudgetName: aws.String(budgetName),
		BudgetType: aws.String(budgetType),
		BudgetLimit: &budgets.Spend{
			Amount: aws.String(budgetLimitAmount),
			Unit:   aws.String(budgetLimitUnit),
		},
		CostTypes: costTypes,
		TimePeriod: &budgets.TimePeriod{
			End:   &budgetTimePeriodEnd,
			Start: &budgetTimePeriodStart,
		},
		TimeUnit:    aws.String(budgetTimeUnit),
		CostFilters: budgetCostFilters,
	}
	return budget, nil
}

func decodeBudgetsBudgetID(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected AccountID:BudgetName", id)
	}
	return parts[0], parts[1], nil
}

func expandBudgetsCostTypesUnmarshal(budgetCostTypes []interface{}) *budgets.CostTypes {
	costTypes := &budgets.CostTypes{
		IncludeCredit:            aws.Bool(true),
		IncludeDiscount:          aws.Bool(true),
		IncludeOtherSubscription: aws.Bool(true),
		IncludeRecurring:         aws.Bool(true),
		IncludeRefund:            aws.Bool(true),
		IncludeSubscription:      aws.Bool(true),
		IncludeSupport:           aws.Bool(true),
		IncludeTax:               aws.Bool(true),
		IncludeUpfront:           aws.Bool(true),
		UseAmortized:             aws.Bool(false),
		UseBlended:               aws.Bool(false),
	}
	if len(budgetCostTypes) == 1 {
		costTypesMap := budgetCostTypes[0].(map[string]interface{})
		for k, v := range map[string]*bool{
			"include_credit":             costTypes.IncludeCredit,
			"include_discount":           costTypes.IncludeDiscount,
			"include_other_subscription": costTypes.IncludeOtherSubscription,
			"include_recurring":          costTypes.IncludeRecurring,
			"include_refund":             costTypes.IncludeRefund,
			"include_subscription":       costTypes.IncludeSubscription,
			"include_support":            costTypes.IncludeSupport,
			"include_tax":                costTypes.IncludeTax,
			"include_upfront":            costTypes.IncludeUpfront,
			"use_amortized":              costTypes.UseAmortized,
			"use_blended":                costTypes.UseBlended,
		} {
			if val, ok := costTypesMap[k]; ok {
				*v = val.(bool)
			}
		}
	}

	return costTypes
}
