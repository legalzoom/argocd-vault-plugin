package main

import (
	argoPlugin "github.com/backjo/argocd-vault-plugin"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/plugin"
	"os"
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{Output: os.Stdout})
	apiClientMeta := &api.PluginAPIClientMeta{}
	flags := apiClientMeta.FlagSet()
	flags.Parse(os.Args[1:])

	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := api.VaultPluginTLSProvider(tlsConfig)

	err := plugin.Serve(&plugin.ServeOpts{
		BackendFactoryFunc: argoPlugin.Factory,
		TLSProviderFunc:    tlsProviderFunc,
	})
	logger.Info("plugin bootstrap caPath={}, caCert={}", tlsConfig.CAPath, tlsConfig.CACert)
	if err != nil {
		logger.Error("plugin shutting down", "error", err)
		os.Exit(1)
	}
}
