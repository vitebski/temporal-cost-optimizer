# Temporal Cost Optimizer

Go backend bootstrap for the Temporal Cost Copilot MVP described in `spec.md`.

## Backend

Run the API server:

```sh
go run ./cmd/api
```

The server listens on `:8080` by default. Override it with:

```sh
HTTP_ADDR=:9090 go run ./cmd/api
```

## Environment

The Temporal Cloud integration points are wired through configuration, but the API methods intentionally return `501 Not Implemented` until the real usage, metrics, and workflow history clients are added.

Supported environment variables:

- `HTTP_ADDR`
- `TEMPORAL_CLOUD_API_KEY`
- `TEMPORAL_CLOUD_ACCOUNT_ID`
- `TEMPORAL_CLOUD_REGION`
- `TEMPORAL_NAMESPACE`

## API Surface

- `GET /healthz`
- `GET /namespaces?top=5`
- `GET /namespaces/{name}/workflow-types?top=5`
- `GET /workflow-types/{workflowType}/usage?namespace={name}`
- `GET /workflows/{workflowId}/analyze`

Run tests:

```sh
go test ./...
```
