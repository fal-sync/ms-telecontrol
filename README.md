# ms-telecontrol

A Go clean‑architecture based telecontrol service. This service receives command requests over HTTP, stores the command state, and publishes the command to MQTT using Eclipse Paho MQTT.

## Structure

```text
.
|-- cmd/api                                      # application entry point
|-- internal/app                                 # dependency wiring and bootstrap integration
|-- internal/config                              # configuration/environment loader
|-- internal/delivery/http                       # HTTP handlers and router
|-- internal/domain                              # entities, domain errors, domain events
|-- internal/usecase                             # business logic
|-- internal/usecase/port                        # outbound ports for external integrations
|-- internal/infrastructure/persistence          # repository adapters
|-- internal/infrastructure/external/httpclient  # shared HTTP client for service‑to‑service calls
|-- internal/infrastructure/external/service     # client/hook for other services
`-- internal/infrastructure/external/messaging   # messaging adapters, including MQTT
```

## Endpoints

- `GET /health`
- `POST /telecontrol/commands`
- `GET /telecontrol/commands`
- `GET /telecontrol/commands/{id}`

Example command request:

```bash
curl --location 'http://localhost:8080/telecontrol/commands' \
  --header 'Content-Type: application/json' \
  --data '{
    "device_id": "device-001",
    "command": "relay.set",
    "payload": {
      "state": "on"
    },
    "correlation_id": "req-001",
    "requested_by": "operator",
    "ttl_seconds": 30
  }'
```

A successful response uses HTTP status `202 Accepted`. The command will be in `published` status if MQTT publishing succeeds, or `failed` if the broker rejects the publish.

## MQTT

Enable MQTT via environment variables:

```bash
MQTT_ENABLED=true
MQTT_BROKERS=tcp://localhost:1883
MQTT_CLIENT_ID=ms-telecontrol
MQTT_COMMAND_TOPIC_PATTERN=telecontrol/{device_id}/commands
MQTT_QOS=1
```

Available placeholder topics:

- `{device_id}`
- `{command}`
- `{command_id}`

The MQTT payload published contains a `TelecontrolCommandIssuedEvent` with fields `id`, `device_id`, `command`, `payload`, `correlation_id`, `requested_by`, and a command timestamp.

## Integration with Other Services

The use‑case layer only depends on ports defined in `internal/usecase/port`:

- `CommandPublisher` for publishing commands to transports such as MQTT.
- `CommandIssuedHook` for integration after a command has been successfully published.

An HTTP hook to a telemetry gateway can be enabled with:

```bash
GATEWAY_TELEMETRY_BASE_URL=http://localhost:9000
GATEWAY_TELEMETRY_COMMAND_PATH=/internal/telecontrol/commands
GATEWAY_TELEMETRY_TIMEOUT=3s
```

## Environment Variables

- `APP_NAME=ms-telecontrol`
- `APP_PORT=8080`
- `APP_SHUTDOWN_TIMEOUT=10s`
- `MQTT_ENABLED=false`
- `MQTT_BROKERS=tcp://localhost:1883`
- `MQTT_CLIENT_ID=ms-telecontrol`
- `MQTT_USERNAME=`
- `MQTT_PASSWORD=`
- `MQTT_COMMAND_TOPIC_PATTERN=telecontrol/{device_id}/commands`
- `MQTT_QOS=1`
- `MQTT_RETAINED=false`
- `MQTT_CONNECT_TIMEOUT=5s`
- `MQTT_PUBLISH_TIMEOUT=3s`
- `GATEWAY_TELEMETRY_BASE_URL=`
- `GATEWAY_TELEMETRY_COMMAND_PATH=/internal/telecontrol/commands`
- `GATEWAY_TELEMETRY_TIMEOUT=3s`

If `MQTT_ENABLED=false`, the API still starts, but `POST /telecontrol/commands` will return `503` because the publisher is not configured.

## Running the Project

```bash
go run ./cmd/api
```

With a local MQTT broker:

```bash
MQTT_ENABLED=true go run ./cmd/api
```

In PowerShell:

```powershell
$env:MQTT_ENABLED="true"
go run ./cmd/api
```

## Architectural Notes

- The MQTT adapter based on Paho resides in `internal/infrastructure/external/messaging/mqtt`.
- Direct publish from the request cycle is fine for an initial prototype. For mission‑critical flows, consider an outbox pattern to avoid command loss when the broker or network fails.
- The current repository implementation is in‑memory. The persistence layer can be swapped without changing the HTTP delivery or use‑case layers as long as the `TelecontrolCommandRepository` contract remains satisfied.
