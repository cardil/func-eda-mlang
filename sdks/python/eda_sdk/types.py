"""Type definitions for the EDA SDK."""

from dataclasses import dataclass
from enum import IntEnum
from typing import Optional


class DestinationType(IntEnum):
    """Destination type for output events."""

    KAFKA = 0
    RABBITMQ = 1
    HTTP = 2
    DISCARD = 3


@dataclass
class KafkaConfig:
    """Kafka connection configuration."""

    broker: str
    topic: str
    group: str


@dataclass
class OutputDestination:
    """Output destination for routing events."""

    type: DestinationType
    target: str
    cluster: Optional[str] = None
