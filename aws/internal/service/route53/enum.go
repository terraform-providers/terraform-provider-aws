package route53

const (
	KeySigningKeyStatusActionNeeded    = "ACTION_NEEDED"
	KeySigningKeyStatusActive          = "ACTIVE"
	KeySigningKeyStatusDeleting        = "DELETING"
	KeySigningKeyStatusInactive        = "INACTIVE"
	KeySigningKeyStatusInternalFailure = "INTERNAL_FAILURE"

	ServeSignatureActionNeeded    = "ACTION_NEEDED"
	ServeSignatureDeleting        = "DELETING"
	ServeSignatureInternalFailure = "INTERNAL_FAILURE"
	ServeSignatureNotSigning      = "NOT_SIGNING"
	ServeSignatureSigning         = "SIGNING"
)
