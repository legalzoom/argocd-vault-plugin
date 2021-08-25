package argocd

import (
	"context"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/project"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const SecretTokenType = "argo_cd_token"

func (b *backend) secretAccessToken() *framework.Secret {
	return &framework.Secret{
		Type: SecretTokenType,
		Fields: map[string]*framework.FieldSchema{
			"authToken": {
				Type:        framework.TypeString,
				Description: `ArgoCD Token`,
			},
			"id": {
				Type:        framework.TypeString,
				Description: `ArgoCD Token Id`,
			},
			"role": {
				Type:        framework.TypeString,
				Description: `ArgoCD Token Role`,
			},
			"project": {
				Type:        framework.TypeString,
				Description: `ArgoCD Token Project`,
			},
		},

		Revoke: b.secretAccessTokenRevoke,
	}
}

func (b *backend) secretAccessTokenRevoke(ctx context.Context, req *logical.Request, _ *framework.FieldData) (*logical.Response, error) {
	client, err := b.getArgoProjectClient(ctx, req.Storage)

	if err != nil {
		return nil, err
	}

	if client == nil {
		return logical.ErrorResponse("backend not configured"), nil
	}

	role := req.Secret.InternalData["role"].(string)
	proj := req.Secret.InternalData["project"].(string)
	id := req.Secret.InternalData["id"].(string)

	_, err = client.DeleteToken(ctx, &project.ProjectTokenDeleteRequest{
		Project: proj,
		Role:    role,
		Iat:     0,
		Id:      id,
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}
