package firebase

import (
	"testing"
	"time"

	"google.golang.org/appengine/aetest"
)

func TestCertificateStore(t *testing.T) {
	ctx, done, err := aetest.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	store := newCertificateStore("")

	cert, err := store.Get(ctx, "09712c9531f921fce0118dba9441de0ed4f408f7")
	if err != nil {
		t.Error(err)
	}
	t.Logf("cert 1 %v", cert.AuthorityKeyId)

	nowFunc = func() time.Time {
		return time.Date(2000, 12, 15, 17, 8, 00, 0, time.UTC)
	}

	cert, err = store.Get(ctx, "09712c9531f921fce0118dba9441de0ed4f408f7")
	if err != nil {
		t.Error(err)
	}
	t.Logf("cert 2 %v", cert.AuthorityKeyId)
}
