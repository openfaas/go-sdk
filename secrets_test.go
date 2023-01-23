package sdk

import (
	"fmt"
	"os"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
)

func Test_ReadSecret_FromTemporaryDirectory(t *testing.T) {

	tmpDir := t.TempDir()

	os.Setenv("secret_mount_path", tmpDir)
	value := "test1234"
	if err := os.WriteFile(tmpDir+"/api-key", []byte(value), 0644); err != nil {
		t.Fatal(err)
	}

	secret, err := ReadSecret("api-key")
	if err != nil {
		t.Fatal(err)
	}

	if secret != value {
		t.Fatalf("want %s, but got %s", value, secret)
	}
}

func Test_ReadSecrets_FromTemporaryDirectory(t *testing.T) {

	tmpDir := t.TempDir()

	os.Setenv("secret_mount_path", tmpDir)
	value1 := "test1234"
	if err := os.WriteFile(tmpDir+"/api-key", []byte(value1), 0644); err != nil {
		t.Fatal(err)
	}

	value2 := "1234test"

	if err := os.WriteFile(tmpDir+"/secret-key", []byte(value2), 0644); err != nil {
		t.Fatal(err)
	}

	secrets, err := ReadSecrets()
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"api-key":    value1,
		"secret-key": value2,
	}

	if len(secrets.values) != len(want) {
		t.Fatalf("want %d secrets, but got %d", len(want), len(secrets.values))
	}

	if !cmp.Equal(want, secrets.values) {
		t.Fatalf("want %v, but got %v", want, secrets)
	}
}

func Test_SecretMap_Exists_Present(t *testing.T) {
	m := newSecretMap(map[string]string{"key1": "value1"})

	if !m.Exists("key1") {
		t.Fatalf("key1 should exist")
	}
}

func Test_SecretMap_Exists_NotPresent(t *testing.T) {
	m := newSecretMap(map[string]string{"key2": "value1"})

	if m.Exists("key1") {
		t.Fatalf("key1 should not exist")
	}
}

func Test_SecretMap_GetValue_Exists(t *testing.T) {
	m := newSecretMap(map[string]string{"key1": "value1"})

	v, err := m.Get("key1")
	if err != nil {
		t.Fatalf("key1 should exist")
	}

	if v != m.values["key1"] {
		t.Fatalf("want %s, but got %s", m.values["key1"], v)
	}
}

func Test_SecretMap_GetValue_NotFound(t *testing.T) {
	m := newSecretMap(map[string]string{})
	secret := "key1"
	_, err := m.Get(secret)
	wantErr := fmt.Errorf("secret %s not found", secret)
	if err.Error() != wantErr.Error() {
		t.Fatalf("want %q, but got %q", wantErr, err)
	}

}
