FROM golang:1.16 as builder

WORKDIR /workspace

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download
COPY cmd/ cmd/
COPY backend.go backend.go
COPY path_config.go path_config.go
COPY secret_token.go secret_token.go

RUN CGO_ENABLED=0 go build -o vault/plugins/argocd cmd/argocd-vault-plugin/main.go

FROM alpine
COPY --from=builder /workspace/vault/plugins/argocd /vault-plugin 

