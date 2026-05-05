# Temporal Cost Optimizer

Go backend bootstrap for the Temporal Cost Copilot MVP described in `spec.md`.

## Backend

Create a local `.env` file:

```sh
cp .env.example .env
```

Set `TEMPORAL_CLOUD_API_KEY` in `.env` before starting the server. The backend requires a real Temporal Cloud API key because it creates the SDK client during startup.

Run the API server:

```sh
go run ./cmd/api
```

The server reads configuration from `.env` in the current working directory and listens on `:8080` by default. Override it in `.env` with:

```dotenv
HTTP_ADDR=:9090
```

## Environment

Temporal Cloud usage is accessed through the experimental Temporal Cloud Go SDK and is intentionally limited to the Cloud Usage API summary records from `temporal/api/cloud/usage/v1/message.proto`.

Supported `.env` variables:

- `HTTP_ADDR`
- `TEMPORAL_CLOUD_API_KEY`
- `TEMPORAL_CLOUD_API_HOST_PORT`
- `TEMPORAL_CLOUD_API_VERSION`
- `TEMPORAL_USAGE_PAGE_SIZE`

Only `GET /namespaces?top=5` is backed by Temporal Cloud usage data today. Workflow-type drilldown and workflow execution analysis still return `501 Not Implemented` because the Usage API groups records by namespace only.

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
