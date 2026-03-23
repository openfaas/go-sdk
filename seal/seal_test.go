package seal

import (
	"bytes"
	"testing"
)

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

	sealed, err := Seal(pub, values, "builder-2026-03")
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
	}, "key-1")
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

	sealed, err := Seal(pub, map[string][]byte{"a": []byte("b")}, "")
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

	sealed, err := Seal(pub, map[string][]byte{"token": []byte("secret")}, "")
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

	// Fake PNG: magic header + null bytes + random binary
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

	sealed, err := Seal(pub, values, "test")
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

func TestKeyID(t *testing.T) {
	pub, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	sealed, err := Seal(pub, map[string][]byte{"k": []byte("v")}, "my-rotation-key")
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	got, err := KeyID(sealed)
	if err != nil {
		t.Fatalf("KeyID: %v", err)
	}
	if got != "my-rotation-key" {
		t.Fatalf("want keyID %q, got %q", "my-rotation-key", got)
	}
}

func TestKeyIDEmpty(t *testing.T) {
	pub, priv, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	sealed, err := Seal(pub, map[string][]byte{"k": []byte("v")}, "")
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	keyID, err := KeyID(sealed)
	if err != nil {
		t.Fatalf("KeyID: %v", err)
	}
	if keyID != "" {
		t.Fatalf("want empty keyID, got %q", keyID)
	}

	got, err := Unseal(priv, sealed)
	if err != nil {
		t.Fatalf("Unseal: %v", err)
	}
	if string(got["k"]) != "v" {
		t.Fatalf("want %q, got %q", "v", got["k"])
	}
}

func TestSealInvalidPublicKey(t *testing.T) {
	_, err := Seal([]byte("not-base64!!!"), map[string][]byte{"k": []byte("v")}, "")
	if err == nil {
		t.Fatal("expected error for invalid base64 key")
	}
}

func TestSealWrongLengthKey(t *testing.T) {
	// 16 bytes instead of 32
	_, err := Seal([]byte("AAAAAAAAAAAAAAAAAAAAAA=="), map[string][]byte{"k": []byte("v")}, "")
	if err == nil {
		t.Fatal("expected error for wrong-length key")
	}
}

func TestUnsealCorruptedCiphertext(t *testing.T) {
	pub, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	sealed, err := Seal(pub, map[string][]byte{"k": []byte("v")}, "")
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	// Corrupt the ciphertext by flipping bytes in the sealed output
	corrupted := make([]byte, len(sealed))
	copy(corrupted, sealed)
	// Find "secrets:" and corrupt a few bytes after it
	for i := range corrupted {
		if i > len(corrupted)-20 {
			corrupted[i] ^= 0xFF
			break
		}
	}

	_, priv, _ := GenerateKeyPair()
	_, err = Unseal(priv, corrupted)
	if err == nil {
		t.Fatal("expected error for corrupted sealed data")
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
