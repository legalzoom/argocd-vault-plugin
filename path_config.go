package argocd

import (
	"context"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

func (b *backend) pathConfig() *framework.Path {
	return &framework.Path{
		Pattern: "config/admin",
		Fields: map[string]*framework.FieldSchema{
			"authToken": {
				Type:        framework.TypeString,
				Required:    true,
				Description: "Administrator token to access ArgoCD",
			},
			"serverAddress": {
				Type:        framework.TypeString,
				Required:    true,
				Description: "Address of the ArgoCD instance",
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathConfigUpdate,
				Summary:  "Configure the ArgoCD secrets backend.",
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.pathConfigDelete,
				Summary:  "Delete the ArgoCD secrets configuration.",
			},
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathConfigRead,
				Summary:  "Examine the ArgoCD secrets configuration.",
			},
		},
		HelpSynopsis:    `Interact with the ArgoCD secrets configuration.`,
		HelpDescription: ``,
	}
}

func (b *backend) pathConfigUpdate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	b.configMutex.Lock()
	defer b.configMutex.Unlock()

	config := &AdminConfig{}
	config.AuthToken = data.Get("authToken").(string)
	config.ServerAddress = data.Get("serverAddress").(string)

	if config.AuthToken == "" {
		return logical.ErrorResponse("authToken is required"), nil
	}

	if config.ServerAddress == "" {
		return logical.ErrorResponse("serverAddress is required"), nil
	}

	entry, err := logical.StorageEntryJSON("config/admin", config)
	if err != nil {
		return nil, err
	}

	err = req.Storage.Put(ctx, entry)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathConfigDelete(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	b.configMutex.Lock()
	defer b.configMutex.Unlock()

	if err := req.Storage.Delete(ctx, "config/admin"); err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathConfigRead(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	b.configMutex.RLock()
	defer b.configMutex.RUnlock()

	config, err := b.getArgoAdminConfig(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	if config == nil {
		return logical.ErrorResponse("backend not configured"), nil
	}

	configMap := map[string]interface{}{
		"serverAddress": config.ServerAddress,
	}

	return &logical.Response{
		Data: configMap,
	}, nil
}
