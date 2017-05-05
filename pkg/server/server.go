package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/openshift/oscanner/pkg/configuration"
	"github.com/openshift/oscanner/pkg/instance"
)

// A Server represents a complete instance of the oscanner.
type Server struct {
	config *configuration.Configuration
	server *http.Server
}

// NewServer creates a new oscanner server.
func NewServer(in *instance.Instance) (*Server, error) {
	handler := NewHTTPHandler(in)
	handler = panicHandler(handler)

	return &Server{
		config: in.Config,
		server: &http.Server{
			Handler: handler,
		},
	}, nil
}

// ListenAndServe runs the registry's HTTP server.
func (srv *Server) ListenAndServe() error {
	config := srv.config

	ln, err := net.Listen("tcp", config.HTTP.Addr)
	if err != nil {
		return err
	}

	if config.HTTP.TLS.Certificate != "" {
		tlsConf := &tls.Config{
			ClientAuth:               tls.NoClientCert,
			NextProtos:               []string{"h2", "http/1.1"},
			MinVersion:               tls.VersionTLS10,
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
		}

		tlsConf.Certificates = make([]tls.Certificate, 1)
		tlsConf.Certificates[0], err = tls.LoadX509KeyPair(config.HTTP.TLS.Certificate, config.HTTP.TLS.Key)
		if err != nil {
			return err
		}

		if len(config.HTTP.TLS.ClientCAs) != 0 {
			pool := x509.NewCertPool()

			for _, ca := range config.HTTP.TLS.ClientCAs {
				caPem, err := ioutil.ReadFile(ca)
				if err != nil {
					return err
				}

				if ok := pool.AppendCertsFromPEM(caPem); !ok {
					return fmt.Errorf("Could not add CA to pool")
				}
			}

			for _, subj := range pool.Subjects() {
				log.Debugf("CA Subject: %s", string(subj))
			}

			tlsConf.ClientAuth = tls.RequireAndVerifyClientCert
			tlsConf.ClientCAs = pool
		}

		ln = tls.NewListener(ln, tlsConf)
		log.Infof("listening on %v, tls", ln.Addr())
	} else {
		log.Infof("listening on %v", ln.Addr())
	}

	return srv.server.Serve(ln)
}

func panicHandler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Panic(fmt.Sprintf("%v", err))
			}
		}()
		handler.ServeHTTP(w, r)
	})
}

func resolveConfiguration(args []string) (*configuration.Configuration, error) {
	var configFile string

	if len(args) > 0 {
		configFile = args[0]
	} else if os.Getenv("OSCANNER_CONFIGURATION_FILE") != "" {
		configFile = os.Getenv("OSCANNER_CONFIGURATION_FILE")
	}

	if configFile == "" {
		return nil, fmt.Errorf("configuration file unspecified")
	}

	config, err := configuration.ParseFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error parsing %s: %v", configFile, err)
	}

	return config, nil
}

func configureLogging(config *configuration.Configuration) error {
	level := string(config.Log.Level)

	l, err := log.ParseLevel(level)
	if err != nil {
		l = log.InfoLevel
		log.Warnf("error parsing level %q: %v, using %q	", level, err, l)
	}

	log.SetLevel(l)

	formatter := config.Log.Formatter
	if formatter == "" {
		// default formatter
		formatter = "text"
	}

	switch formatter {
	case "json":
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	case "text":
		log.SetFormatter(&log.TextFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	default:
		return fmt.Errorf("unsupported logging formatter: %q", config.Log.Formatter)
	}

	log.Debugf("using %q logging formatter", formatter)

	return nil
}
