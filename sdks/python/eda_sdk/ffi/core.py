"""FFI-based Core implementation using cffi."""

from typing import Any, Optional

from ..types import DestinationType, KafkaConfig, OutputDestination
from .loader import ffi, load_library


class FFICore:
    """Core implementation using FFI (cffi) to call Rust shared library."""

    def __init__(self) -> None:
        """Initialize the FFI core.

        Raises:
            RuntimeError: If library loading fails.
        """
        self._lib: Any = load_library()

    def get_kafka_config(self) -> KafkaConfig:
        """Retrieve the Kafka connection configuration.

        Returns:
            KafkaConfig with broker, topic, and group settings.

        Raises:
            RuntimeError: If configuration cannot be retrieved.
        """
        # Get broker
        broker_ptr = self._lib.eda_get_kafka_broker()
        if broker_ptr == ffi.NULL:
            raise RuntimeError("Failed to get Kafka broker")
        broker = ffi.string(broker_ptr).decode("utf-8")  # type: ignore[union-attr]
        self._lib.eda_free_string(broker_ptr)

        # Get topic
        topic_ptr = self._lib.eda_get_kafka_topic()
        if topic_ptr == ffi.NULL:
            raise RuntimeError("Failed to get Kafka topic")
        topic = ffi.string(topic_ptr).decode("utf-8")  # type: ignore[union-attr]
        self._lib.eda_free_string(topic_ptr)

        # Get group
        group_ptr = self._lib.eda_get_kafka_group()
        if group_ptr == ffi.NULL:
            raise RuntimeError("Failed to get Kafka group")
        group = ffi.string(group_ptr).decode("utf-8")  # type: ignore[union-attr]
        self._lib.eda_free_string(group_ptr)

        return KafkaConfig(broker=broker, topic=topic, group=group)

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
        error_bytes = error.encode("utf-8")
        result = self._lib.eda_should_retry(error_bytes, attempt)
        return bool(result)  # type: ignore[no-any-return]

    def calculate_backoff(self, attempt: int) -> int:
        """Calculate backoff duration in milliseconds.

        Args:
            attempt: The current attempt number (1-indexed).

        Returns:
            Backoff duration in milliseconds.

        Raises:
            RuntimeError: If backoff calculation fails.
        """
        result = self._lib.eda_calculate_backoff(attempt)
        return int(result)

    def get_output_destination(self, event_json: str) -> OutputDestination:
        """Route an output event to its destination.

        Args:
            event_json: JSON-serialized CloudEvent.

        Returns:
            OutputDestination specifying where to send the event.

        Raises:
            RuntimeError: If routing fails.
        """
        event_bytes = event_json.encode("utf-8")
        dest_ptr = self._lib.eda_get_output_destination(event_bytes)
        if dest_ptr == ffi.NULL:
            raise RuntimeError("Failed to get output destination")

        try:
            # Extract destination type
            dest_type = DestinationType(dest_ptr.dest_type)

            # Extract target string
            if dest_ptr.target == ffi.NULL:
                raise RuntimeError("Destination target is NULL")
            target = ffi.string(dest_ptr.target).decode("utf-8")  # type: ignore[union-attr]

            # Extract cluster string (optional)
            cluster: Optional[str] = None
            if dest_ptr.cluster != ffi.NULL:
                cluster = ffi.string(dest_ptr.cluster).decode("utf-8")  # type: ignore[union-attr]

            return OutputDestination(type=dest_type, target=target, cluster=cluster)
        finally:
            self._lib.eda_free_output_destination(dest_ptr)

    def load_routing_config(self, file_path: str) -> None:
        """Load routing configuration from a YAML file.

        Args:
            file_path: Path to the routing configuration YAML file.

        Raises:
            RuntimeError: If loading fails.
        """
        path_bytes = file_path.encode("utf-8")
        success = self._lib.eda_load_routing_config(path_bytes)
        if not success:
            raise RuntimeError(f"Failed to load routing config from {file_path}")

    def close(self) -> None:
        """Release resources (no-op for FFI implementation)."""
        pass
