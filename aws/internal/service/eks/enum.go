package eks

const (
	ResourcesSecrets = "secrets"
)

func Resources_Values() []string {
	return []string{
		ResourcesSecrets,
	}
}
