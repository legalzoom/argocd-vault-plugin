package argocd

import (
	"context"
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/project"
	"github.com/docker/distribution/uuid"
	"github.com/hashicorp/go-hclog"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient"
)

// backend wraps the backend framework and adds a map for storing key value pairs
type backend struct {
	*framework.Backend
	configMutex sync.RWMutex
	store       map[string][]byte
}

type AdminConfig struct {
	ServerAddress string `json:"serverAddress"`
	AuthToken     string `json:"authToken"`
}

var _ logical.Factory = Factory

// Factory configures and returns Mock backends
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b, err := newBackend(conf.Logger)
	if err != nil {
		return nil, err
	}

	if conf == nil {
		return nil, fmt.Errorf("configuration passed into backend is nil")
	}

	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}

	return b, nil
}

func newBackend(logger hclog.Logger) (*backend, error) {
	logger.Info("init backend")
	b := &backend{
		store: make(map[string][]byte),
	}

	b.Backend = &framework.Backend{
		Help:        strings.TrimSpace(mockHelp),
		BackendType: logical.TypeLogical,
		PathsSpecial: &logical.Paths{
			SealWrapStorage: []string{"config/admin"},
		},
		Paths: append([]*framework.Path{b.pathConfig()},
			b.paths()...,
		),
		Secrets: []*framework.Secret{b.secretAccessToken()},
	}

	return b, nil
}

func (b *backend) paths() []*framework.Path {
	return []*framework.Path{
		{
			Pattern: framework.MatchAllRegex("path"),

			Fields: map[string]*framework.FieldSchema{
				"path": {
					Type:        framework.TypeString,
					Description: "Specifies the path of the secret.",
				},
			},

			Operations: map[logical.Operation]framework.OperationHandler{
				logical.ReadOperation: &framework.PathOperation{
					Callback: b.handleRead,
					Summary:  "Retrieve the secret from the map.",
				},
			},

			ExistenceCheck: b.handleExistenceCheck,
		},
	}
}

func (b *backend) getArgoAdminConfig(ctx context.Context, storage logical.Storage) (*AdminConfig, error) {

	// Read in the backend configuration
	entry, err := storage.Get(ctx, "config/admin")
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	var argoConfig = AdminConfig{}

	if err := entry.DecodeJSON(&argoConfig); err != nil {
		return nil, err
	}
	return &argoConfig, nil
}

func (b *backend) getArgoProjectClient(ctx context.Context, storage logical.Storage) (project.ProjectServiceClient, error) {

	// Read in the backend configuration
	entry, err := storage.Get(ctx, "config/admin")
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	var argoConfig = AdminConfig{}

	if err := entry.DecodeJSON(&argoConfig); err != nil {
		return nil, err
	}

	argoClient, err := apiclient.NewClient(&apiclient.ClientOptions{
		ServerAddr: argoConfig.ServerAddress,
		UserAgent:  "Vault ArgoCD Plugin",
		AuthToken:  argoConfig.AuthToken,
	})
	if err != nil {
		return nil, err
	}
	_, projClient, err := argoClient.NewProjectClient()
	if err != nil {
		return nil, err
	}

	return projClient, nil
}

func (b *backend) handleExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	return true, nil
}

func (b *backend) handleRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if req.ClientToken == "" {
		return nil, fmt.Errorf("client token empty")
	}

	path := data.Get("path").(string)
	subpaths := strings.Split(path, "/")

	projClient, err := b.getArgoProjectClient(ctx, req.Storage)
	if err != nil {
		return nil, err
	}

	var tokenId = uuid.Generate().String()

	createToken, err := projClient.CreateToken(ctx, &project.ProjectTokenCreateRequest{
		Project:     subpaths[0],
		Description: "Dynamically generated vault secret",
		Role:        subpaths[1],
		ExpiresIn:   3600,
		Id:          tokenId,
	})
	if err != nil {
		return nil, err
	}
	secret := b.Secret(SecretTokenType)
	secret.DefaultDuration, _ = time.ParseDuration("1h")
	return secret.Response(map[string]interface{}{
		"authToken": createToken.Token,
	}, map[string]interface{}{
		"authToken": createToken.Token,
		"role":      subpaths[1],
		"project":   subpaths[0],
		"id":        tokenId,
	}), nil
}

const mockHelp = `
The Kubernetes backend is a secrets backend that vends Argo CD tokens.
`
