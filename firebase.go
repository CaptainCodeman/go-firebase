package firebase // import "github.com/captaincodeman/go-firebase"

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

type App struct {
	name  string
	creds *Credentials
}

const (
	defaultAppName = "[DEFAULT]"
)

var apps = struct {
	sync.RWMutex
	m map[string]*App
}{
	m: make(map[string]*App),
}

func GetApp(name string) (*App, error) {
	name = normalizeName(name)
	if name == "" {
		name = defaultAppName
	}

	apps.RLock()
	defer apps.RUnlock()

	app, ok := apps.m[name]
	if !ok {
		return nil, fmt.Errorf("App %s not yet initialized!", name)
	}
	return app, nil
}

func New(options ...Option) (*App, error) {
	cfg := defaultConfig()
	for _, option := range options {
		if err := option(cfg); err != nil {
			return nil, err
		}
	}

	if cfg.Credentials == nil {
		r, err := os.Open(cfg.CredentialsPath)
		if err != nil {
			return nil, err
		}
		c, err := loadCredential(r)
		if err != nil {
			return nil, err
		}
		cfg.Credentials = c
	}

	app := &App{
		name:  cfg.Name,
		creds: cfg.Credentials,
	}

	apps.Lock()
	defer apps.Unlock()

	if _, ok := apps.m[app.name]; ok {
		return nil, fmt.Errorf("App %s already exists!", app.name)
	}

	apps.m[app.name] = app
	return app, nil
}

func (a *App) Auth() *Auth {
	return &Auth{
		app: a,
	}
}

func (a *App) Name() string {
	return a.name
}

func normalizeName(name string) string {
	return strings.TrimSpace(name)
}
