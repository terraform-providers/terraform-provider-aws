package encryption

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/terraform-plugin-sdk/internal/vault/helper/pgpkeys"
)

// RetrieveGPGKey returns the PGP key specified as the pgpKey parameter, or queries
// the public key from the keybase service if the parameter is a keybase username
// prefixed with the phrase "keybase:"
//
// Deprecated: This function will be removed in v2 without replacement. Please
// see https://www.terraform.io/docs/extend/best-practices/sensitive-state.html#don-39-t-encrypt-state
// for more information.
func RetrieveGPGKey(pgpKey string) (string, error) {
	const keybasePrefix = "keybase:"

	encryptionKey := pgpKey
	if strings.HasPrefix(pgpKey, keybasePrefix) {
		publicKeys, err := pgpkeys.FetchKeybasePubkeys([]string{pgpKey})
		if err != nil {
			return "", errwrap.Wrapf(fmt.Sprintf("Error retrieving Public Key for %s: {{err}}", pgpKey), err)
		}
		encryptionKey = publicKeys[pgpKey]
	}

	return encryptionKey, nil
}

// EncryptValue encrypts the given value with the given encryption key. Description
// should be set such that errors return a meaningful user-facing response.
//
// Deprecated: This function will be removed in v2 without replacement. Please
// see https://www.terraform.io/docs/extend/best-practices/sensitive-state.html#don-39-t-encrypt-state
// for more information.
func EncryptValue(encryptionKey, value, description string) (string, string, error) {
	fingerprints, encryptedValue, err :=
		pgpkeys.EncryptShares([][]byte{[]byte(value)}, []string{encryptionKey})
	if err != nil {
		return "", "", errwrap.Wrapf(fmt.Sprintf("Error encrypting %s: {{err}}", description), err)
	}

	return fingerprints[0], base64.StdEncoding.EncodeToString(encryptedValue[0]), nil
}
