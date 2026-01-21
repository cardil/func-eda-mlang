"""WASM-based Core implementation (placeholder)."""

from ..types import KafkaConfig, OutputDestination


class WASMCore:
    """Core implementation using WASM (placeholder).

    Note: Full WASM Component Model support in Python is still evolving.
    This is a placeholder for future implementation using wasmtime-py.
    """

    def __init__(self) -> None:
        """Initialize the WASM core.

        Note: The WASM module would be embedded in the package.

        Raises:
            NotImplementedError: WASM support not yet implemented.
        """
        raise NotImplementedError(
            "WASM support is not yet implemented in the Python SDK. "
            "Please use the FFI implementation (eda_sdk.ffi) instead."
        )

    def get_kafka_config(self) -> KafkaConfig:
        """Retrieve the Kafka connection configuration."""
        raise NotImplementedError("WASM support not yet implemented")

    def should_retry(self, error: str, attempt: int) -> bool:
        """Check if an error should be retried."""
        raise NotImplementedError("WASM support not yet implemented")

    def calculate_backoff(self, attempt: int) -> int:
        """Calculate backoff duration in milliseconds."""
        raise NotImplementedError("WASM support not yet implemented")

    def get_output_destination(self, event_json: str) -> OutputDestination:
        """Route an output event to its destination."""
        raise NotImplementedError("WASM support not yet implemented")

    def load_routing_config(self, file_path: str) -> None:
        """Load routing configuration from a YAML file."""
        raise NotImplementedError("WASM support not yet implemented")

    def close(self) -> None:
        """Release resources."""
        pass
