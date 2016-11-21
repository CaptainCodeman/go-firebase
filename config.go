package firebase

type (
	// Config stores firebase app configuration settings
	Config struct {
		Name            string
		Credentials     *Credentials
		CredentialsPath string
	}

	// Option is the signature for configuration options
	Option func(*Config) error
)

func defaultConfig() *Config {
	return &Config{
		Name:            defaultAppName,
		CredentialsPath: "firebase-credentials.json",
	}
}

// WithName sets the name of the app
func WithName(name string) func(*Config) error {
	return func(c *Config) error {
		c.Name = normalizeName(name)
		return nil
	}
}

// WithCredentialsPath sets the path to load credentials from
func WithCredentialsPath(path string) func(*Config) error {
	return func(c *Config) error {
		c.CredentialsPath = path
		return nil
	}
}

// WithCredentials sets the credentials
func WithCredentials(creds *Credentials) func(*Config) error {
	return func(c *Config) error {
		c.Credentials = creds
		return nil
	}
}
