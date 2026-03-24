package seal

import (
	"bytes"
	"testing"
)

func TestDeriveKeyID(t *testing.T) {
	pub, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	id1, err := DeriveKeyID(pub)
	if err != nil {
		t.Fatalf("DeriveKeyID: %v", err)
	}
	if len(id1) != 8 {
		t.Fatalf("want 8 chars, got %d: %q", len(id1), id1)
	}

	// Deterministic
	id2, _ := DeriveKeyID(pub)
	if id1 != id2 {
		t.Fatalf("not deterministic: %q != %q", id1, id2)
	}

	// Different key → different ID
	pub2, _, _ := GenerateKeyPair()
	id3, _ := DeriveKeyID(pub2)
	if id1 == id3 {
		t.Fatalf("different keys produced same ID: %q", id1)
	}
}

func TestSealDeriveKeyID(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	sealed, err := Seal(pub, map[string][]byte{"k": []byte("v")})
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	// Verify the envelope has the derived key_id
	keyID, err := KeyID(sealed)
	if err != nil {
		t.Fatalf("KeyID: %v", err)
	}

	wantID, _ := DeriveKeyID(pub)
	if keyID != wantID {
		t.Fatalf("want derived key_id %q, got %q", wantID, keyID)
	}

	// Still decrypts fine
	got, err := Unseal(priv, sealed)
	if err != nil {
		t.Fatalf("Unseal: %v", err)
	}
	if string(got["k"]) != "v" {
		t.Fatalf("want %q, got %q", "v", got["k"])
	}
}

func TestSealUnsealRoundTrip(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	values := map[string][]byte{
		"pip_token":      []byte("s3cr3t"),
		"aws_secret_key": []byte("wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
		"ca.crt":         []byte("-----BEGIN CERTIFICATE-----\nMIIBkTCB...\n-----END CERTIFICATE-----\n"),
	}

	sealed, err := Seal(pub, values)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	got, err := Unseal(priv, sealed)
	if err != nil {
		t.Fatalf("Unseal: %v", err)
	}

	for k, want := range values {
		if !bytes.Equal(got[k], want) {
			t.Fatalf("key %q: want %q, got %q", k, want, got[k])
		}
	}

	if len(got) != len(values) {
		t.Fatalf("want %d keys, got %d", len(values), len(got))
	}
}

func TestUnsealSingleKey(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	sealed, err := Seal(pub, map[string][]byte{
		"pip_token": []byte("s3cr3t"),
		"npm_token": []byte("npm_4F4k3"),
	})
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	got, err := UnsealKey(priv, sealed, "pip_token")
	if err != nil {
		t.Fatalf("UnsealKey: %v", err)
	}

	if string(got) != "s3cr3t" {
		t.Fatalf("want %q, got %q", "s3cr3t", got)
	}
}

func TestUnsealKeyMissing(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	sealed, err := Seal(pub, map[string][]byte{"a": []byte("b")})
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	_, err = UnsealKey(priv, sealed, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestUnsealWrongKey(t *testing.T) {
	pub, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	sealed, err := Seal(pub, map[string][]byte{"token": []byte("secret")})
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	_, wrongPriv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	_, err = Unseal(wrongPriv, sealed)
	if err == nil {
		t.Fatal("expected error when unsealing with wrong key")
	}
}

func TestMixedBinaryAndText(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	fakePNG := append([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, make([]byte, 1024)...)
	for i := 8; i < len(fakePNG); i++ {
		fakePNG[i] = byte(i % 256)
	}

	values := map[string][]byte{
		"logo.png":   fakePNG,
		"api_token":  []byte("sk-1234567890abcdef"),
		"ca.crt":     []byte("-----BEGIN CERTIFICATE-----\nMIIBkTCB+w==\n-----END CERTIFICATE-----\n"),
		"null_bytes": {0x00, 0x00, 0xFF, 0x00, 0xFE},
	}

	sealed, err := Seal(pub, values)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	got, err := Unseal(priv, sealed)
	if err != nil {
		t.Fatalf("Unseal: %v", err)
	}

	for k, want := range values {
		if !bytes.Equal(got[k], want) {
			t.Fatalf("key %q: byte mismatch (want %d bytes, got %d)", k, len(want), len(got[k]))
		}
	}
}

func TestSealInvalidPublicKey(t *testing.T) {
	_, err := Seal([]byte("not-base64!!!"), map[string][]byte{"k": []byte("v")})
	if err == nil {
		t.Fatal("expected error for invalid base64 key")
	}
}

func TestSealWrongLengthKey(t *testing.T) {
	_, err := Seal([]byte("AAAAAAAAAAAAAAAAAAAAAA=="), map[string][]byte{"k": []byte("v")})
	if err == nil {
		t.Fatal("expected error for wrong-length key")
	}
}

func TestUnsealBadVersion(t *testing.T) {
	_, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	_, err = Unseal(priv, []byte("version: v99\nalgorithm: nacl/box\n"))
	if err == nil {
		t.Fatal("expected error for unsupported version")
	}
}

func TestUnsealBadAlgorithm(t *testing.T) {
	_, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	_, err = Unseal(priv, []byte("version: v1\nalgorithm: aes-gcm\n"))
	if err == nil {
		t.Fatal("expected error for unsupported algorithm")
	}
}
