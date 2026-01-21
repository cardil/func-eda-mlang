"""Main run function for WASM-based EDA functions (placeholder)."""

from ..consumer import Handler


def run(handler: Handler) -> None:
    """Start the EDA consumer using WASM core with the given handler.

    Note: WASM support is not yet implemented in the Python SDK.
    The WASM module would be embedded in the package, similar to the FFI library.

    Args:
        handler: User's event handler function (SimpleHandler or OutputHandler).

    Raises:
        NotImplementedError: WASM support not yet implemented.
    """
    raise NotImplementedError(
        "WASM support is not yet implemented in the Python SDK. "
        "Please use the FFI implementation (eda_sdk.ffi) instead."
    )
