package main

import (
	"crypto/x509"
	argoPlugin "github.com/backjo/argocd-vault-plugin"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/plugin"
	"io/ioutil"
	"os"
)

func main() {
	apiClientMeta := &api.PluginAPIClientMeta{}
	flags := apiClientMeta.FlagSet()
	flags.Parse(os.Args[1:])

	tlsConfig := apiClientMeta.GetTLSConfig()
	if tlsConfig != nil {
		tlsConfig.Insecure = true
	} else {
		tlsConfig = &api.TLSConfig{Insecure: true}
	}
	tlsProviderFunc := api.VaultPluginTLSProvider(tlsConfig)

	dat, _ := ioutil.ReadFile("/vault/secrets/ca.crt")
	certPool, _ := x509.SystemCertPool()
	certPool.AppendCertsFromPEM(dat)
	err := plugin.Serve(&plugin.ServeOpts{
		BackendFactoryFunc: argoPlugin.Factory,
		TLSProviderFunc:    tlsProviderFunc,
	})
	if err != nil {
		logger := hclog.New(&hclog.LoggerOptions{Output: os.Stdout})
		logger.Error("plugin shutting down", "error", err)
		os.Exit(1)
	}
}
