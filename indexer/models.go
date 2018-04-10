package indexer

// ProfileTokenFile models a signed Blockstack Profile
type ProfileTokenFile struct {
	Token           string       `json:"token,omitempty"`
	ParentPublicKey string       `json:"parentPublicKey,omitempty"`
	Encrypted       bool         `json:"encrypted,omitempty"`
	DecodedToken    DecodedToken `json:"decodedToken,omitempty"`
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

// Profile contains social proofs and images
type Profile struct {
	Type      string     `json:"@type"`
	Context   string     `json:"@context,omitempty"`
	Name      string     `json:"name,omitempty"`
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
