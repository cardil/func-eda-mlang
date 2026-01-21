"""Main run function for FFI-based EDA functions."""

from ..consumer import Handler
from ..run import run_with_core, setup_logging
from .core import FFICore


def run(handler: Handler) -> None:
    """Start the EDA consumer using FFI core with the given handler.

    This is the main entry point for FFI-based functions.

    Args:
        handler: User's event handler function (SimpleHandler or OutputHandler).

    Raises:
        RuntimeError: If consumer fails.

    Example:
        >>> from cloudevents.http import CloudEvent
        >>> from eda_sdk.ffi import run
        >>>
        >>> def handle(event: CloudEvent) -> None:
        ...     print(f"Received: {event['type']}")
        >>>
        >>> if __name__ == "__main__":
        ...     run(handle)
    """
    # Setup logging
    setup_logging()

    # Create FFI core
    core = FFICore()

    # Run with core
    try:
        run_with_core(core, handler)
    finally:
        core.close()
