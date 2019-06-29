package keys

// KeyHandler interface
type KeyHandler interface {
	// Type returns key format type
	Type() uint32
	// Version returns epoch version number
	// Each time a new key is created, it must be stored with a different
	// epoch number. New encryptions must use latest key version.
	Version() uint32
	// Generate creates a new key, which is completely out of control of
	// the user. It is using secure random for generating the keys.
	Generate() error
	// AES returns the AES key
	AES() []byte
	// HMAC returns the HMAC key
	HMAC() []byte
	// String provides a string representation of the key. It is safe to show
	// it publicly.
	String() string
}
