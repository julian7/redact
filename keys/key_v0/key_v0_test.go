package keyv0

import "testing"

const (
	sampleAES  = "0123456789abcdefghijklmnopqrstuv"
	sampleHMAC = "0123456789abcdefghijklmnopqrstuv0123456789abcdefghijklmnopqrstuv"
)

func TestNewKey(t *testing.T) {
	epoch := uint32(5)
	key := NewKey(epoch)
	if key == nil {
		t.Error("No key returned")
		return
	}
	if key.Epoch != epoch {
		t.Errorf(
			"invalid epoch\nExpected: %d\nReceived: %d",
			epoch,
			key.Epoch,
		)
	}
}

func TestKeyVersion(t *testing.T) {
	key := KeyV0{}
	expected := uint32(0)
	if key.Version() != expected {
		t.Errorf(
			"invalid version\nExpected: %d\nReceived: %d",
			expected,
			key.Version(),
		)
	}
}

func TestGenerate(t *testing.T) {
	key := KeyV0{Epoch: 1}
	err := key.Generate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if len(key.AES()) != AESKeySize {
		t.Errorf(
			"Invalid AES key size\nExpected: %d\nReceived: %d",
			AESKeySize,
			len(key.AES()),
		)
	}
	if len(key.HMAC()) != HMACKeySize {
		t.Errorf(
			"Invalid HMAC key size\nExpected: %d\nReceived: %d",
			HMACKeySize,
			len(key.HMAC()),
		)
	}
}

func TestString(t *testing.T) {
	key := KeyV0{
		Epoch: 1,
	}
	copy(key.AESData[:], sampleAES)
	copy(key.HMACData[:], sampleHMAC)
	strval := key.String()
	expected := "#1 b3ec9ddd"
	if strval != expected {
		t.Errorf(
			"Invalid string representation of key\nExpected: %s\nReceived: %s",
			expected,
			strval,
		)
	}
}
