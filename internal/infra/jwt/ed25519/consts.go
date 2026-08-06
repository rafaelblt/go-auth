package ed25519

import "github.com/golang-jwt/jwt/v5"

var keyAlgorithm = jwt.SigningMethodEdDSA.Alg()

const (
	keyCurve = "Ed25519"
	keyType  = "OKP"
)
