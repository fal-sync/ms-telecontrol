# ms-telecontrol

Service telecontrol berbasis Go clean architecture. Service ini menerima request command lewat HTTP, menyimpan state command, lalu mem-publish command ke MQTT memakai Eclipse Paho MQTT.

## Struktur

```text
.
|-- cmd/api                                      # entry point aplikasi
|-- internal/app                                 # wiring dependency dan bootstrap integration
|-- internal/config                              # loader config/env
|-- internal/delivery/http                       # HTTP handler dan router
|-- internal/domain                              # entity, domain error, domain event
|-- internal/usecase                             # business logic
|-- internal/usecase/port                        # outbound port untuk external integration
|-- internal/infrastructure/persistence          # adapter repository
|-- internal/infrastructure/external/httpclient  # shared HTTP client untuk service-to-service
|-- internal/infrastructure/external/service     # client/hook ke service lain
`-- internal/infrastructure/external/messaging   # adapter messaging, termasuk MQTT
```

## Endpoint

- `GET /health`
- `POST /telecontrol/commands`
- `GET /telecontrol/commands`
- `GET /telecontrol/commands/{id}`

Contoh issue command:

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

Response sukses memakai status HTTP `202 Accepted`. Command akan berada di status `published` jika publish MQTT berhasil, atau `failed` jika broker menolak publish.

## MQTT

Aktifkan MQTT dengan env:

```bash
MQTT_ENABLED=true
MQTT_BROKERS=tcp://localhost:1883
MQTT_CLIENT_ID=ms-telecontrol
MQTT_COMMAND_TOPIC_PATTERN=telecontrol/{device_id}/commands
MQTT_QOS=1
```

Placeholder topic yang tersedia:

- `{device_id}`
- `{command}`
- `{command_id}`

Payload MQTT yang dipublish berisi event `TelecontrolCommandIssuedEvent`, termasuk `id`, `device_id`, `command`, `payload`, `correlation_id`, `requested_by`, dan timestamp command.

## Integrasi Service Lain

Usecase hanya bergantung pada port di `internal/usecase/port`:

- `CommandPublisher` untuk publish command ke transport seperti MQTT.
- `CommandIssuedHook` untuk integrasi setelah command berhasil dipublish.

Hook HTTP ke gateway telemetry dapat diaktifkan dengan:

```bash
GATEWAY_TELEMETRY_BASE_URL=http://localhost:9000
GATEWAY_TELEMETRY_COMMAND_PATH=/internal/telecontrol/commands
GATEWAY_TELEMETRY_TIMEOUT=3s
```

## Environment Variable

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

Jika `MQTT_ENABLED=false`, API tetap start, tetapi `POST /telecontrol/commands` akan mengembalikan `503` karena publisher belum dikonfigurasi.

## Menjalankan Project

```bash
go run ./cmd/api
```

Dengan MQTT lokal:

```bash
MQTT_ENABLED=true go run ./cmd/api
```

Di PowerShell:

```powershell
$env:MQTT_ENABLED="true"
go run ./cmd/api
```

## Catatan Arsitektur

- Adapter MQTT berbasis Paho ada di `internal/infrastructure/external/messaging/mqtt`.
- Flow publish langsung dari request cycle cocok untuk tahap awal. Untuk flow mission critical, pertimbangkan outbox pattern agar command tidak hilang saat broker atau jaringan bermasalah.
- Repository saat ini masih memory. Layer persistence bisa diganti tanpa mengubah delivery HTTP atau usecase selama kontrak `TelecontrolCommandRepository` tetap dipenuhi.
