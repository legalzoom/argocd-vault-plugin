# Vault ArgoCD Secret Plugin

Vault ArgoCD Secret plugin is a secrets engine plugin for [HashiCorp Vault](https://www.vaultproject.io/) that allows for the generation and usage
of short-term credentials for ArgoCD.

## Usage
This plugin requires a few setup steps before it will work.

1. Enable the plugin on your Vault infrastructure. See [https://www.vaultproject.io/docs/internals/plugins](https://www.vaultproject.io/docs/internals/plugins) for more information.
2. Create a mount point using the plugin. `vault secrets enable -path=argocd argocd`
3. Configure the mount point with an ArgoCD API Key. The user associated with the key should have access to update all projects in ArgoCD. 
```bash
vault write argocd/config/admin authToken="yourAuthJWT" serverAddress="argo.mydomain.com:443"
```

Once configured, you should be all set to retrieve credentials for ArgoCD.
`vault read argocd/project-name/role-name`
## Developing

All commands can be run using the provided [Makefile](./Makefile). However, it may be instructive to look at the commands to gain a greater understanding of how Vault registers plugins. Using the Makefile will result in running the Vault server in `dev` mode. Do not run Vault in `dev` mode in production. The `dev` server allows you to configure the plugin directory as a flag, and automatically registers plugin binaries in that directory. In production, plugin binaries must be manually registered.

This will build the plugin binary and start the Vault dev server:

```
# Build Vault ArgoCD Secret plugin and start Vault dev server with plugin automatically registered
$ make
```

Now open a new terminal window and run the following commands:

```
# Open a new terminal window and export Vault dev server http address
$ export VAULT_ADDR='http://127.0.0.1:8200'

# Enable the ArgoCD plugin
$ make enable
