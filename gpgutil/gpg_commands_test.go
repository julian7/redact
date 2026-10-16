package gpgutil

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestParseSecretKeyListing(t *testing.T) {
	const (
		primary   = "0123456789abcdef0123456789abcdef01234567"
		other     = "89abcdef0123456789abcdef0123456789abcdef"
		subkey    = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		keyRecord = "sec:u:255:22:89ABCDEF01234567:1700000000:::u:::cEC:::#::ed25519:::0:\n"
	)

	primaryRecord := "fpr:::::::::" + strings.ToUpper(primary) + ":\n"
	otherRecord := "fpr:::::::::" + strings.ToUpper(other) + ":\n"
	subkeyRecords := "ssb:u:255:18:AAAAAAAAAAAAAAAA:1700000000:::::e:::+::cv25519::\n" +
		"fpr:::::::::" + strings.ToUpper(subkey) + ":\n"

	tests := []struct {
		name     string
		listing  string
		keys     []string
		warnings int
	}{
		{name: "empty"},
		{name: "primary", listing: keyRecord + primaryRecord, keys: []string{primary}},
		{
			name:    "no final newline",
			listing: keyRecord + strings.TrimSuffix(primaryRecord, "\n"),
			keys:    []string{primary},
		},
		{name: "ten fields", listing: keyRecord + "fpr:::::::::" + primary, keys: []string{primary}},
		{name: "nine fields", listing: keyRecord + "fpr::::::::\n", warnings: 1},
		{name: "too few fields", listing: keyRecord + "fpr:\n", warnings: 1},
		{name: "empty fingerprint", listing: keyRecord + "fpr::::::::::\n", warnings: 1},
		{name: "invalid hex", listing: keyRecord + "fpr:::::::::0123GG:\n", warnings: 1},
		{name: "odd hex", listing: keyRecord + "fpr:::::::::012:\n", warnings: 1},
		{
			name:     "continue after invalid fingerprint",
			listing:  keyRecord + "fpr:::::::::0123GG:\n" + keyRecord + otherRecord,
			keys:     []string{other},
			warnings: 1,
		},
		{
			name:    "primary with secret subkey",
			listing: keyRecord + primaryRecord + "grp:::::::::keygrip:\nuid:::::::::recipient:\n" + subkeyRecords,
			keys:    []string{primary},
		},
		{
			name:    "multiple primary keys",
			listing: keyRecord + primaryRecord + subkeyRecords + keyRecord + otherRecord,
			keys:    []string{primary, other},
		},
		{name: "orphan fingerprint", listing: primaryRecord},
		{name: "public key", listing: "pub:u:\n" + primaryRecord + "sub:u:\n" + otherRecord},
		{
			name:    "unknown records",
			listing: "tru::1:\n" + keyRecord + "future:ignored:\n" + primaryRecord,
			keys:    []string{primary},
		},
		{
			name:    "longer fingerprint",
			listing: keyRecord + "fpr:::::::::" + strings.Repeat("AB", 32) + ":\n",
			keys:    []string{strings.Repeat("ab", 32)},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			keys, warnings := parseSecretKeyListing([]byte(tc.listing))

			got := make([]string, len(keys))
			for i, key := range keys {
				got[i] = fmt.Sprintf("%x", key)
			}

			if !slices.Equal(got, tc.keys) {
				t.Fatalf("fingerprints: got %v, want %v", got, tc.keys)
			}

			if len(warnings) != tc.warnings {
				t.Fatalf("expected %d warnings, got %v", tc.warnings, warnings)
			}
		})
	}
}
