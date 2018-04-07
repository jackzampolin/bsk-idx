package indexer

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/btcsuite/btcd/btcec"
)

// ProfileTokenFile models a signed Blockstack Profile
type ProfileTokenFile struct {
	Token           string       `json:"token,omitempty"`
	ParentPublicKey string       `json:"parentPublicKey,omitempty"`
	Encrypted       bool         `json:"encrypted,omitempty"`
	DecodedToken    DecodedToken `json:"decodedToken,omitempty"`
}

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

// export function extractProfile(token, publicKeyOrAddress = null) {
//   let decodedToken
//   if (publicKeyOrAddress) {
//     decodedToken = verifyProfileToken(token, publicKeyOrAddress)
//   } else {
//     decodedToken = decodeToken(token)
//   }
//
//   let profile = {}
//   if (decodedToken.hasOwnProperty('payload')) {
//     const payload = decodedToken.payload
//     if (payload.hasOwnProperty('claim')) {
//       profile = decodedToken.payload.claim
//     }
//   }
//
//   return profile
// }

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

// DecodedToken contains most of the profile information
type DecodedToken struct {
	Payload   Payload `json:"payload,omitempty"`
	Signature string  `json:"signature,omitempty"`
	Header    Header  `json:"header,omitempty"`
}

// Payload contains social media claim info and other
type Payload struct {
	Claim     Profile   `json:"claim,omitempty"`
	IssuedAt  string    `json:"issuedAt,omitempty"`
	Subject   PublicKey `json:"subject,omitempty"`
	Issuer    PublicKey `json:"issuer,omitempty"`
	ExpiresAt string    `json:"expiresAt,omitempty"`
}

// Account models a social media proof
// TODO: Write method on Account to check proof
type Account struct {
	Type       string `json:"@type,omitempty"`
	Service    string `json:"service,omitempty"`
	ProofType  string `json:"proofType,omitempty"`
	Identifier string `json:"identifier,omitempty"`
	ProofURL   string `json:"proofUrl,omitempty"`
}

// Image models a Profile Image
type Image struct {
	Type       string `json:"@type"`
	ContentURL string `json:"contentUrl,omitempty"`
	Name       string `json:"name,omitempty"`
}

// Claim contains social proofs and images
type Profile struct {
	Type      string     `json:"@type"`
	Context   string     `json:@context,omitempty`
	Image     []Image    `json:"image,omitempty"`
	Account   []Account  `json:"account,omitempty"`
	Website   []Website  `json:"website,omitempty"`
	WorksFor  []WorksFor `json:"worksFor,omitempty"`
	Knows     []Knows    `json:"knows,omitempty"`
	Address   Address    `json:"address,omitempty"`
	BirthDate string     `json:"birthDate,omitempty"`
	TaxID     string     `json:"taxID,omitempty"`
}

// Address models an address return
type Address struct {
	Type            string `json:"@type,omitempty"`
	StreetAddress   string `json:"streetAddress,omitempty"`
	AddressLocality string `json:"addressLocality,omitempty"`
	PostalCode      string `json:"postalCode,omitempty"`
	AddressCountry  string `json:"addressCountry,omitempty"`
}

// Knows models the knows return
type Knows struct {
	Type string `json:"@type"`
	ID   string `json:"@id,omitempty"`
}

// WorksFor models the worksfor return
type WorksFor struct {
	Type string `json:"@type"`
	ID   string `json:"@id,omitempty"`
}

// Website models a website
type Website struct {
	Type string `json:"@type,omitempty"`
	URL  string `json:"url,omitempty"`
}

// PublicKey models a publicKey
type PublicKey struct {
	PublicKey string `json:"publicKey,omitempty"`
}

// Header describes the encryption types
type Header struct {
	Typ string `json:"typ,omitempty"`
	Alg string `json:"alg,omitempty"`
}
