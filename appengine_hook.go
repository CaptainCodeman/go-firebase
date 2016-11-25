// +build appengine

// App Engine hooks.

package firebase

import (
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

func init() {
	RegisterRequestContextFunc(requestContextAppEngine)
	RegisterContextClientFunc(contextClientAppEngine)
}

func requestContextAppEngine(req *http.Request) (context.Context, error) {
	return appengine.NewContext(req), nil
}

func contextClientAppEngine(ctx context.Context) (*http.Client, error) {
	return urlfetch.Client(ctx), nil
}
