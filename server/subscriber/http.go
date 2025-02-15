package subscriber

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"time"

	"github.com/cnosdatabase/cnosdb/client"
	"github.com/cnosdatabase/cnosdb/server/coordinator"
)

// HTTP supports writing points over HTTP using the line protocol.
type HTTP struct {
	c client.Client
}

// NewHTTP returns a new HTTP points writer with default options.
func NewHTTP(addr string, timeout time.Duration) (*HTTP, error) {
	return NewHTTPS(addr, timeout, false, "", nil)
}

// NewHTTPS returns a new HTTPS points writer with default options and HTTPS configured.
func NewHTTPS(addr string, timeout time.Duration, unsafeSsl bool, caCerts string, tlsConfig *tls.Config) (*HTTP, error) {
	tlsConfig, err := createTLSConfig(caCerts, tlsConfig)
	if err != nil {
		return nil, err
	}

	conf := client.HTTPConfig{
		Addr:               addr,
		Timeout:            timeout,
		InsecureSkipVerify: unsafeSsl,
		TLSConfig:          tlsConfig,
	}

	c, err := client.NewHTTPClient(conf)
	if err != nil {
		return nil, err
	}
	return &HTTP{c: c}, nil
}

// WritePoints writes points over HTTP transport.
func (h *HTTP) WritePoints(p *coordinator.WritePointsRequest) (err error) {
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:   p.Database,
		TimeToLive: p.TimeToLive,
	})
	for _, pt := range p.Points {
		bp.AddPoint(client.NewPointFrom(pt))
	}
	err = h.c.Write(bp)
	return
}

func createTLSConfig(caCerts string, tlsConfig *tls.Config) (*tls.Config, error) {
	if caCerts == "" {
		if tlsConfig != nil {
			return tlsConfig.Clone(), nil
		}
		return nil, nil
	}
	return loadCaCerts(caCerts, tlsConfig)
}

func loadCaCerts(caCerts string, tlsConfig *tls.Config) (*tls.Config, error) {
	caCert, err := ioutil.ReadFile(caCerts)
	if err != nil {
		return nil, err
	}

	out := new(tls.Config)
	if tlsConfig != nil {
		out = tlsConfig.Clone()
	}

	out.RootCAs = x509.NewCertPool()
	out.RootCAs.AppendCertsFromPEM(caCert)
	return out, nil
}
