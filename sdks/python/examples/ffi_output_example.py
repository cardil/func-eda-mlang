#!/usr/bin/env python3
"""FFI output example - demonstrates output event routing."""

from datetime import datetime, timezone
from typing import Optional

from cloudevents.http import CloudEvent

from eda_sdk.ffi import run


def handle(event: CloudEvent) -> Optional[CloudEvent]:
    """Handle incoming CloudEvents and produce output events.

    This demonstrates the output event routing capability.
    """
    print(f"📨 Received event: {event['id']}")
    print(f"   Type: {event['type']}")
    print(f"   Source: {event['source']}")

    # Only process events of type "kafka.message"
    # Return None for other types (no output event)
    if event["type"] != "kafka.message":
        return None

    # Create an output event
    output_event = CloudEvent(
        {
            "specversion": "1.0",
            "type": "com.example.processed",
            "source": "ffi-output-example",
            "id": f"processed-{event['id']}",
            "time": datetime.now(timezone.utc).isoformat(),
        },
        data={
            "original_id": event["id"],
            "original_type": event["type"],
            "processed_at": datetime.now(timezone.utc).isoformat(),
            "message": "Event processed successfully",
        },
    )

    print(f"✅ Producing output event: {output_event['id']} (type: {output_event['type']})")

    return output_event


if __name__ == "__main__":
    run(handle)
