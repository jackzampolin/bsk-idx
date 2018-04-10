package indexer

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcec"
)

// returns the publicKey from this signed profile in the following order
//   - pt.ParentPublicKey
//   - pt.DecodedToken.Payload.Issuer.PublicKey
//   - pt.DecodedToken.Payload.Subject.PublicKey
func (pt *ProfileTokenFile) getPubKey() string {
	if pt.ParentPublicKey != "" {
		return pt.ParentPublicKey
	} else if pt.DecodedToken.Payload.Issuer.PublicKey != "" {
		return pt.DecodedToken.Payload.Issuer.PublicKey
	} else if pt.DecodedToken.Payload.Subject.PublicKey != "" {
		return pt.DecodedToken.Payload.Subject.PublicKey
	}
	return ""
}

// Validate if the profile is valid
func (pt *ProfileTokenFile) Validate() error {
	dt := &Payload{}
	out := strings.Split(pt.Token, ".")
	err := json.Unmarshal([]byte(out[1]), dt)

	if err != nil {
		return fmt.Errorf("Error unmarshalling tokenPayload %s", err)
	}
	if dt.Subject.PublicKey == "" {
		return fmt.Errorf("Token doesn't have a subject public key")
	}
	if dt.Issuer.PublicKey == "" {
		return fmt.Errorf("Token doesn't have an issuer public key")
	}
	if dt.Claim.Type == "" {
		return fmt.Errorf("Token doesn't have a claim")
	}

	// Decode hex-encoded serialized public key.
	pubKeyBytes, err := hex.DecodeString(dt.Issuer.PublicKey)
	if err != nil {
		return err
	}

	// const issuerPublicKey = payload.issuer.publicKey
	// const publicKeyBuffer = new Buffer(issuerPublicKey, 'hex')
	//
	// const Q = ecurve.Point.decodeFrom(secp256k1, publicKeyBuffer)
	// const compressedKeyPair = new ECPair(null, Q, { compressed: true })
	// const compressedAddress = compressedKeyPair.getAddress()
	// const uncompressedKeyPair = new ECPair(null, Q, { compressed: false })
	// const uncompressedAddress = uncompressedKeyPair.getAddress()

	pubKey, err := btcec.ParsePubKey(pubKeyBytes, btcec.S256())
	if err != nil {
		return err
	}

	if pt.getPubKey() == string(pubKey.SerializeCompressed()) {
		// pass
	} else if pt.getPubKey() == string(pubKey.SerializeUncompressed()) {
		// pass
	} else {
		return fmt.Errorf("Token issuer public key does not match the verifying value")
	}

	// // Decode hex-encoded serialized signature.
	// sigBytes, err := hex.DecodeString("30450220090ebfb3690a0ff115bb1b38b" +
	// 	"8b323a667b7653454f1bccb06d4bbdca42c2079022100ec95778b51e707" +
	// 	"1cb1205f8bde9af6592fc978b0452dafe599481c46d6b2e479")
	//
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// signature, err := btcec.ParseSignature(sigBytes, btcec.S256())
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	return nil
}

// JSON Marshals things
func (pt *ProfileTokenFile) JSON() string {
	byt, err := json.Marshal(pt)
	if err != nil {
		// We Unmarshal this JSON so there should be no errors Marshaling
		panic(fmt.Sprintf("error marshalling profile %s", err))
	}
	return string(byt)
}
