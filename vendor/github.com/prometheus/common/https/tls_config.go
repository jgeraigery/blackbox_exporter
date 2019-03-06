package https

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"

	"gopkg.in/yaml.v2"
)

// Config struct holds information to generate a tls.Config
type Config struct {
	TLSCertPath string    `yaml:"tlsCertPath"`
	TLSKeyPath  string    `yaml:"tlsKeyPath"`
	TLSConfig   TLSStruct `yaml:"tlsConfig"`
}

// TLSStruct forms part of the Config
type TLSStruct struct {
	RootCAs                  string   `yaml:"rootCAs"`
	ServerName               string   `yaml:"serverName"`
	ClientAuth               string   `yaml:"clientAuth"`
	ClientCAs                string   `yaml:"clientCAs"`
	InsecureSkipVerify       bool     `yaml:"insecureSkipVerify"`
	CipherSuites             []uint16 `yaml:"cipherSuites"`
	PreferServerCipherSuites bool     `yaml:"preferServerCipherSuites"`
	MinVersion               uint16   `yaml:"minVersion"`
	MaxVersion               uint16   `yaml:"maxVersion"`
}

// GetTLSConfig take a path to a yml config file and returns a tls.Config based on its values
func GetTLSConfig(configPath string) *tls.Config {
	cfg, _, _ := GetConfigAndPaths(configPath)
	return cfg
}

// GetConfigAndPaths take a path to a yml config file and returns a tls.Config based on its values, as well as paths to the cert and key files
func GetConfigAndPaths(configPath string) (*tls.Config, string, string) {
	config, err := loadConfigFromYaml(configPath)
	if err != nil {
		log.Fatal("Config failed to load from Yaml", err)
	}
	tlsc, err := loadTLSConfig(config)
	if err != nil {
		log.Fatal("Failed to convert Config to tls.Config", err)
	}
	return tlsc, config.TLSCertPath, config.TLSKeyPath
}

func loadConfigFromYaml(fileName string) (*Config, error) {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	c := &Config{}
	err = yaml.Unmarshal(content, c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func loadTLSConfig(c *Config) (*tls.Config, error) {
	cfg := &tls.Config{}
	if len(c.TLSCertPath) > 0 {
		cfg.GetCertificate = func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert, err := tls.LoadX509KeyPair(c.TLSCertPath, c.TLSKeyPath)
			if err != nil {
				return nil, err
			}
			return &cert, nil
		}
		cfg.BuildNameToCertificate()
	}
	if len(c.TLSConfig.ServerName) > 0 {
		cfg.ServerName = c.TLSConfig.ServerName
	}
	if (c.TLSConfig.InsecureSkipVerify) == true {
		cfg.InsecureSkipVerify = true
	}
	if len(c.TLSConfig.CipherSuites) > 0 {
		cfg.CipherSuites = c.TLSConfig.CipherSuites
	}
	if (c.TLSConfig.PreferServerCipherSuites) == true {
		cfg.PreferServerCipherSuites = c.TLSConfig.PreferServerCipherSuites
	}
	if (c.TLSConfig.MinVersion) != 0 {
		cfg.MinVersion = c.TLSConfig.MinVersion
	}
	if (c.TLSConfig.MaxVersion) != 0 {
		cfg.MaxVersion = c.TLSConfig.MaxVersion
	}
	if len(c.TLSConfig.RootCAs) > 0 {
		rootCertPool := x509.NewCertPool()
		rootCAFile, err := ioutil.ReadFile(c.TLSConfig.RootCAs)
		if err != nil {
			return cfg, err
		}
		rootCertPool.AppendCertsFromPEM(rootCAFile)
		cfg.RootCAs = rootCertPool
	}
	if len(c.TLSConfig.ClientCAs) > 0 {
		clientCAPool := x509.NewCertPool()
		clientCAFile, err := ioutil.ReadFile(c.TLSConfig.ClientCAs)
		if err != nil {
			return cfg, err
		}
		clientCAPool.AppendCertsFromPEM(clientCAFile)
		cfg.ClientCAs = clientCAPool
	}
	if len(c.TLSConfig.ClientAuth) > 0 {
		switch s := (c.TLSConfig.ClientAuth); s {
		case "RequestClientCert":
			cfg.ClientAuth = tls.RequestClientCert
		case "RequireClientCert":
			cfg.ClientAuth = tls.RequireAnyClientCert
		case "VerifyClientCertIfGiven":
			cfg.ClientAuth = tls.VerifyClientCertIfGiven
		case "RequireAndVerifyClientCert":
			cfg.ClientAuth = tls.RequireAndVerifyClientCert
		default:
			cfg.ClientAuth = tls.NoClientCert
		}
	}
	return cfg, nil
}

// Listen is a utility function that can be called to start either a TLS or unsecured server
// based on whether the server has a TLSConfig configured.
func Listen(server *http.Server) error {
	if server.TLSConfig != nil {
		return server.ListenAndServeTLS("", "")
	}
	return server.ListenAndServe()
}
