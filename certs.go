package firebase

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"

	"golang.org/x/net/context"
)

const (
	defaultCertsCacheTime = 1 * time.Hour

	// URL containing the public keys for the Google certs
	clientCertURL = "https://www.googleapis.com/robot/v1/metadata/x509/securetoken@system.gserviceaccount.com"
)

type (
	certificateStore struct {
		sync.RWMutex
		url   string
		certs map[string]*x509.Certificate
		exp   time.Time
	}
)

var (
	certs *certificateStore
)

func init() {
	certs = newCertificateStore("")
}

func newCertificateStore(url string) *certificateStore {
	if url == "" {
		url = clientCertURL
	}
	return &certificateStore{
		url: url,
	}
}

func (c *certificateStore) Get(ctx context.Context, kid string) (*x509.Certificate, error) {
	if err := c.ensureLoaded(ctx); err != nil {
		return nil, err
	}

	c.RLock()
	defer c.RUnlock()

	cert, found := c.certs[kid]
	if !found {
		return nil, fmt.Errorf("certificate not found for key ID: %s", kid)
	}
	return cert, nil
}

func (c *certificateStore) ensureLoaded(ctx context.Context) error {
	c.RLock()
	if c.exp.After(clock.Now()) {
		c.RUnlock()
		return nil
	}
	c.RUnlock()

	certs, cacheTime, err := c.download(ctx)
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()
	c.certs = certs
	c.exp = clock.Now().Add(cacheTime)
	return nil
}

// TODO: pass in transport, provide appengine stub to automatically get it from context
func (c *certificateStore) download(ctx context.Context) (map[string]*x509.Certificate, time.Duration, error) {
	client, err := ContextClient(ctx)
	if err != nil {
		return nil, 0, err
	}

	resp, err := client.Get(c.url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("download %s fails: %s", c.url, resp.Status)
	}
	certs, err := parse(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	return certs, cacheTime(resp), nil
}

// parse parses the certificates response in JSON format.
// The response has the format:
// {
//   "kid1": "-----BEGIN CERTIFICATE-----...-----END CERTIFICATE-----",
//   "kid2": "-----BEGIN CERTIFICATE-----...-----END CERTIFICATE-----",
// }
func parse(r io.Reader) (map[string]*x509.Certificate, error) {
	m := make(map[string]string)
	dec := json.NewDecoder(r)
	if err := dec.Decode(&m); err != nil {
		return nil, err
	}
	certs := make(map[string]*x509.Certificate)
	for k, v := range m {
		block, _ := pem.Decode([]byte(v))
		c, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		certs[k] = c
	}
	return certs, nil
}

// cacheTime extracts the cache time from the HTTP response header.
// A default cache time is returned if extraction fails.
func cacheTime(resp *http.Response) time.Duration {
	cc := strings.Split(resp.Header.Get("Cache-Control"), ",")
	const maxAge = "max-age="
	for _, c := range cc {
		c = strings.TrimSpace(c)
		if strings.HasPrefix(c, maxAge) {
			if d, err := strconv.Atoi(c[len(maxAge):]); err == nil {
				return time.Duration(d) * time.Second
			}
		}
	}
	return defaultCertsCacheTime
}
