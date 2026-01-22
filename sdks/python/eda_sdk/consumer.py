"""Kafka consumer with CloudEvents support."""

import json
import logging
from typing import Callable, Optional, Union

from cloudevents.http import CloudEvent
from cloudevents.kafka import KafkaMessage, from_binary, from_structured
from confluent_kafka import Consumer as KafkaConsumer, KafkaError, KafkaException, Message, Producer

from .core import Core
from .types import DestinationType

# Type aliases for handler functions
SimpleHandler = Callable[[CloudEvent], None]
OutputHandler = Callable[[CloudEvent], Optional[CloudEvent]]
Handler = Union[SimpleHandler, OutputHandler]

logger = logging.getLogger(__name__)


class Consumer:
    """Manages Kafka consumption and event processing."""

    def __init__(self, core: Core, handler: Handler) -> None:
        """Initialize the consumer.

        Args:
            core: Core implementation (FFI or WASM).
            handler: User's event handler function.

        Raises:
            RuntimeError: If initialization fails.
        """
        self.core = core
        self.handler = handler
        self._consumer: Optional[KafkaConsumer] = None
        self._producer: Optional[Producer] = None

        # Detect handler signature
        self._is_output_handler = self._detect_output_handler(handler)

        # Get Kafka config from core
        config = self.core.get_kafka_config()
        logger.info(
            "Kafka configuration loaded",
            extra={"broker": config.broker, "topic": config.topic, "group": config.group},
        )

        # Create Kafka consumer
        self._consumer = KafkaConsumer(
            {
                "bootstrap.servers": config.broker,
                "group.id": config.group,
                "auto.offset.reset": "earliest",
                "enable.auto.commit": True,
            }
        )

        # Subscribe to topic
        self._consumer.subscribe([config.topic])
        logger.info(f"Subscribed to topic: {config.topic}")

        # Create Kafka producer for output events (if handler returns events)
        if self._is_output_handler:
            self._producer = Producer({"bootstrap.servers": config.broker})
            logger.info("Kafka producer initialized for output events")

    def _detect_output_handler(self, handler: Handler) -> bool:
        """Detect if handler returns output events.

        Args:
            handler: The handler function to inspect.

        Returns:
            True if handler returns CloudEvent, False otherwise.
        """
        # Check type hints if available
        if hasattr(handler, "__annotations__"):
            return_type = handler.__annotations__.get("return")
            if return_type is not None:
                # Check if return type includes Optional[CloudEvent] or CloudEvent
                return "CloudEvent" in str(return_type)
        # Default to simple handler
        return False

    def start(self) -> None:
        """Start consuming events (blocking).

        Raises:
            RuntimeError: If consumer fails.
        """
        if self._consumer is None:
            raise RuntimeError("Consumer not initialized")

        logger.info("Starting consumer")
        consecutive_errors = 0
        max_consecutive_errors = 5

        try:
            while True:
                # Poll for messages with timeout
                msg = self._consumer.poll(timeout=0.1)

                if msg is None:
                    continue

                if msg.error():
                    error = msg.error()
                    if error is not None and error.code() == KafkaError._PARTITION_EOF:  # type: ignore[attr-defined]
                        # End of partition, not an error
                        continue
                    else:
                        logger.error(f"Consumer error: {msg.error()}")
                        consecutive_errors += 1
                        if consecutive_errors >= max_consecutive_errors:
                            raise KafkaException(msg.error())
                        continue

                # Reset error counter on successful read
                consecutive_errors = 0

                # Parse CloudEvent
                event = self._parse_cloud_event(msg)
                if event is None:
                    continue

                # Call user handler
                try:
                    self._invoke_handler(event)
                except Exception as e:
                    logger.error(f"Handler error: {e}", extra={"event_type": event["type"]})

                    # Check if we should retry using core
                    try:
                        should_retry = self.core.should_retry(str(e), 1)
                        if should_retry:
                            backoff = self.core.calculate_backoff(1)
                            logger.warning(
                                "Would retry after backoff (not implemented in PoC)",
                                extra={"backoff_ms": backoff},
                            )
                    except Exception as retry_err:
                        logger.error(f"Error checking retry: {retry_err}")

        except KeyboardInterrupt:
            logger.info("Consumer interrupted")
        finally:
            self.close()

    def _parse_cloud_event(self, msg: Message) -> Optional[CloudEvent]:
        """Convert Kafka message to CloudEvent.

        Tries multiple parsing strategies in order:
        1. Structured mode: CloudEvent attributes in Kafka headers, data in body
        2. Binary mode: Full CloudEvent as JSON in Kafka value (both CE fields and data)

        Args:
            msg: Kafka message.

        Returns:
            CloudEvent or None if parsing fails.
        """
        value = msg.value()
        if value is None:
            logger.warning("Received message with None value, skipping")
            return None

        # Convert confluent_kafka Message to cloudevents KafkaMessage
        headers: dict[str, bytes] = {}
        msg_headers = msg.headers()
        if msg_headers is not None:
            for key, val in msg_headers:  # type: ignore[misc]
                if val is not None:
                    # Ensure val is bytes
                    if isinstance(val, str):
                        headers[key] = val.encode("utf-8")
                    elif isinstance(val, bytes):
                        headers[key] = val

        kafka_msg = KafkaMessage(headers=headers, key=msg.key(), value=value)

        # Try 1: Structured mode (CE attributes in headers, data in body)
        try:
            event = from_structured(kafka_msg)
            return event
        except Exception:
            pass  # Not structured mode, try next method

        # Try 2: Binary mode (full CloudEvent JSON in body)
        try:
            event = from_binary(kafka_msg)
            return event
        except Exception as e:
            logger.warning(f"Failed to parse CloudEvent: {e}")
            return None

    def _invoke_handler(self, event: CloudEvent) -> None:
        """Call the user's handler function and handle output events.

        Args:
            event: The CloudEvent to process.

        Raises:
            Exception: If handler or output publishing fails.
        """
        if self._is_output_handler:
            # Handler returns Optional[CloudEvent]
            output_event = self.handler(event)  # type: ignore
            if output_event is not None:
                self._publish_output_event(output_event)
        else:
            # Simple handler returns None
            self.handler(event)  # type: ignore

    def _publish_output_event(self, event: CloudEvent) -> None:
        """Route and publish an output event.

        Args:
            event: The CloudEvent to publish.

        Raises:
            RuntimeError: If publishing fails.
        """
        if self._producer is None:
            raise RuntimeError("Producer not initialized")

        # Serialize event to JSON for routing
        event_dict = {
            "specversion": event["specversion"],
            "type": event["type"],
            "source": event["source"],
            "id": event["id"],
        }
        if "time" in event:
            event_dict["time"] = event["time"]
        if "datacontenttype" in event:
            event_dict["datacontenttype"] = event["datacontenttype"]
        if "data" in event:
            event_dict["data"] = event["data"]

        event_json = json.dumps(event_dict)

        # Get output destination from core routing
        dest = self.core.get_output_destination(event_json)

        logger.info(
            "Routing output event",
            extra={
                "event_type": event["type"],
                "dest_type": dest.type.name,
                "dest_target": dest.target,
            },
        )

        # Handle different destination types
        if dest.type == DestinationType.KAFKA:
            self._publish_to_kafka(event, dest.target)
        elif dest.type == DestinationType.DISCARD:
            logger.info(f"Discarding output event: {event['type']}")
        elif dest.type in (DestinationType.HTTP, DestinationType.RABBITMQ):
            # TODO: Implement HTTP and RabbitMQ publishing
            logger.warning(
                f"Destination type {dest.type.name} not yet implemented, discarding event"
            )
        else:
            raise RuntimeError(f"Unknown destination type: {dest.type}")

    def _publish_to_kafka(self, event: CloudEvent, topic: str) -> None:
        """Publish an event to a Kafka topic.

        Args:
            event: The CloudEvent to publish.
            topic: The Kafka topic name.

        Raises:
            RuntimeError: If publishing fails.
        """
        if self._producer is None:
            raise RuntimeError("Producer not initialized")

        # Serialize event to JSON
        event_dict = {
            "specversion": event["specversion"],
            "type": event["type"],
            "source": event["source"],
            "id": event["id"],
        }
        if "time" in event:
            event_dict["time"] = event["time"]
        if "datacontenttype" in event:
            event_dict["datacontenttype"] = event["datacontenttype"]
        if "data" in event:
            event_dict["data"] = event["data"]

        event_json = json.dumps(event_dict).encode("utf-8")

        # Produce to Kafka
        self._producer.produce(
            topic=topic,
            value=event_json,
            key=event["id"].encode("utf-8"),
        )
        self._producer.flush()

        logger.info(f"Published output event to Kafka topic: {topic}")

    def close(self) -> None:
        """Release resources."""
        if self._producer is not None:
            self._producer.flush(timeout=5)
            logger.info("Producer flushed")

        if self._consumer is not None:
            self._consumer.close()
            logger.info("Consumer closed")

        if self.core is not None:
            self.core.close()
            logger.info("Core closed")
