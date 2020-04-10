package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
)

func expandMediaConvertReservationPlanSettings(config map[string]interface{}) *mediaconvert.ReservationPlanSettings {
	reservationPlanSettings := &mediaconvert.ReservationPlanSettings{}

	if v, ok := config["commitment"]; ok {
		reservationPlanSettings.Commitment = aws.String(v.(string))
	}

	if v, ok := config["renewal_type"]; ok {
		reservationPlanSettings.RenewalType = aws.String(v.(string))
	}

	if v, ok := config["reserved_slots"]; ok {
		reservationPlanSettings.ReservedSlots = aws.Int64(int64(v.(int)))
	}

	return reservationPlanSettings
}

func flattenMediaConvertReservationPlan(reservationPlan *mediaconvert.ReservationPlan) []interface{} {
	if reservationPlan == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"commitment":     aws.StringValue(reservationPlan.Commitment),
		"renewal_type":   aws.StringValue(reservationPlan.RenewalType),
		"reserved_slots": aws.Int64Value(reservationPlan.ReservedSlots),
	}

	return []interface{}{m}
}
