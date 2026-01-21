"""EDA SDK for Python - Event-Driven Architecture Functions.

This SDK provides a unified interface for building EDA functions with
shared Rust core logic via FFI or WASM backends.

Usage:
    from cloudevents.http import CloudEvent
    from eda_sdk.ffi import run

    def handle(event: CloudEvent) -> None:
        print(f"Received: {event['type']}")

    if __name__ == "__main__":
        run(handle)
"""

__version__ = "0.1.0"
