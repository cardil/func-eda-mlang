#!/usr/bin/env python3
"""Simple FFI example - demonstrates basic event handling."""

from cloudevents.http import CloudEvent

from eda_sdk.ffi import run


def handle(event: CloudEvent) -> None:
    """Handle incoming CloudEvents.

    This is what a developer would write for their EDA function.
    """
    print(f"📨 Received event: {event['id']}")
    print(f"   Type: {event['type']}")
    print(f"   Source: {event['source']}")

    # User's business logic would go here
    # For this example, we just log the event


if __name__ == "__main__":
    run(handle)
