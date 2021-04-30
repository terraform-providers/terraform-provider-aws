package servicecatalog

const (
	ServiceCatalogAcceptLanguageEnglish  = "en"
	ServiceCatalogAcceptLanguageJapanese = "jp"
	ServiceCatalogAcceptLanguageChinese  = "zh"
)

func AcceptLanguage_Values() []string {
	return []string{
		ServiceCatalogAcceptLanguageEnglish,
		ServiceCatalogAcceptLanguageJapanese,
		ServiceCatalogAcceptLanguageChinese,
	}
}
