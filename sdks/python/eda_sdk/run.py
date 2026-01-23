"""Main run function for EDA SDK."""

import inspect
import logging
import signal
import sys
from pathlib import Path
from typing import Optional

from .consumer import Consumer, Handler
from .core import Core

logger = logging.getLogger(__name__)


def run_with_core(core: Core, handler: Handler) -> None:
    """Start the EDA consumer with an explicit core instance.

    Args:
        core: Core implementation (FFI or WASM).
        handler: User's event handler function (SimpleHandler or OutputHandler).

    Raises:
        RuntimeError: If consumer fails.
    """
    # Try to load routing configuration from the caller's directory
    caller_dir = _get_caller_directory()
    if caller_dir:
        routing_config_path = caller_dir / "routing.yaml"
        if routing_config_path.exists():
            logger.info(f"Loading routing configuration from {routing_config_path}")
            try:
                core.load_routing_config(str(routing_config_path))
            except Exception as e:
                raise RuntimeError(f"Failed to load routing config: {e}") from e

    # Create consumer
    try:
        consumer = Consumer(core, handler)
    except Exception as e:
        raise RuntimeError(f"Failed to create consumer: {e}") from e

    # Setup signal handling for graceful shutdown
    def signal_handler(signum, frame):  # type: ignore
        logger.info("Shutting down...")
        consumer.close()
        sys.exit(0)

    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    # Start consuming
    logger.info("Starting EDA consumer...")
    try:
        consumer.start()
    except Exception as e:
        logger.error(f"Consumer error: {e}")
        raise


def _get_caller_directory() -> Optional[Path]:
    """Get the directory of the caller's script.

    Returns:
        Path to the caller's directory, or None if not found.
    """
    # Walk up the call stack to find the first caller outside the SDK
    for frame_info in inspect.stack():
        frame_path = Path(frame_info.filename)
        # Skip SDK internal files
        if "eda_sdk" not in str(frame_path):
            return frame_path.parent

    return None


def setup_logging(level: int = logging.INFO) -> None:
    """Setup structured logging for the SDK.

    Args:
        level: Logging level (default: INFO).
    """
    logging.basicConfig(
        level=level,
        format='{"time":"%(asctime)s","level":"%(levelname)s","msg":"%(message)s"}',
        datefmt="%Y-%m-%dT%H:%M:%S",
    )
