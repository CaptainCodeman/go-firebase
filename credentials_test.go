package firebase

import (
	"os"
	"testing"
)

func TestCredentials(t *testing.T) {
	r, err := os.Open("app/credentials.json")
	if err != nil {
		t.Error(err)
	}
	c, err := loadCredential(r)
	if err != nil {
		t.Error(err)
	}

	t.Logf("client email %s", c.ClientEmail)
	// t.Logf("private key %v", c.PrivateKey)
	// t.Logf("project id %s", c.ProjectID)
}
