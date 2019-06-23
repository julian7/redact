package keys

// KeyHandler interface
type KeyHandler interface {
	Version() uint32
	Generate() error
	AES() []byte
	HMAC() []byte
}
