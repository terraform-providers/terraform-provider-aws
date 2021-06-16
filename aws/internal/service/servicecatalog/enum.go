package servicecatalog

const (
	// If AWS adds these to the API, we should use those and remove these.

	AcceptLanguageEnglish  = "en"
	AcceptLanguageJapanese = "jp"
	AcceptLanguageChinese  = "zh"

	ConstraintTypeLaunch         = "LAUNCH"
	ConstraintTypeNotification   = "NOTIFICATION"
	ConstraintTypeResourceUpdate = "RESOURCE_UPDATE"
	ConstraintTypeStackset       = "STACKSET"
	ConstraintTypeTemplate       = "TEMPLATE"
)

func AcceptLanguage_Values() []string {
	return []string{
		AcceptLanguageEnglish,
		AcceptLanguageJapanese,
		AcceptLanguageChinese,
	}
}

func ConstraintType_Values() []string {
	return []string{
		ConstraintTypeLaunch,
		ConstraintTypeNotification,
		ConstraintTypeResourceUpdate,
		ConstraintTypeStackset,
		ConstraintTypeTemplate,
	}
}
