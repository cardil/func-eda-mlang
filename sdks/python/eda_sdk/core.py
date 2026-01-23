"""Core interface for EDA SDK implementations."""

from typing import Protocol

from .types import KafkaConfig, OutputDestination


class Core(Protocol):
    """Protocol defining the interface for EDA core implementations.

    This can be implemented by FFI or WASM backends.
    """

    def get_kafka_config(self) -> KafkaConfig:
        """Retrieve the Kafka connection configuration.

        Returns:
            KafkaConfig with broker, topic, and group settings.

        Raises:
            RuntimeError: If configuration cannot be retrieved.
        """
        ...

    def should_retry(self, error: str, attempt: int) -> bool:
        """Check if an error should be retried.

        Args:
            error: The error message to check.
            attempt: The current attempt number (1-indexed).

        Returns:
            True if the error should be retried, False otherwise.

        Raises:
            RuntimeError: If retry check fails.
        """
        ...

    def calculate_backoff(self, attempt: int) -> int:
        """Calculate backoff duration in milliseconds.

        Args:
            attempt: The current attempt number (1-indexed).

        Returns:
            Backoff duration in milliseconds.

        Raises:
            RuntimeError: If backoff calculation fails.
        """
        ...

    def get_output_destination(self, event_json: str) -> OutputDestination:
        """Route an output event to its destination.

        Args:
            event_json: JSON-serialized CloudEvent.

        Returns:
            OutputDestination specifying where to send the event.

        Raises:
            RuntimeError: If routing fails.
        """
        ...

    def load_routing_config(self, file_path: str) -> None:
        """Load routing configuration from a YAML file.

        Args:
            file_path: Path to the routing configuration YAML file.

        Raises:
            RuntimeError: If loading fails.
        """
        ...

    def close(self) -> None:
        """Release resources held by the Core implementation."""
        ...
