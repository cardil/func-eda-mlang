#!/bin/bash
# Usage: ./send-test-event.sh [event-type] [data]

EVENT_TYPE=${1:-"test"}
DATA=${2:-"Hello from test"}

echo "Building CloudEvent and sending to Kafka topic 'events'..."
echo ""

# Build CloudEvent using kn-event and send to Kafka via rpk
go run knative.dev/kn-plugin-event/cmd/kn-event@latest build \
  --type "$EVENT_TYPE" \
  --source "/test" \
  --field "message=$DATA" \
  --output json | podman exec -i redpanda rpk topic produce events

echo ""
echo "Event sent successfully!"
