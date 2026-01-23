"""FFI library loader with embedded library support."""

import atexit
import os
import platform
import shutil
import tempfile
from pathlib import Path
from typing import Any, Optional

from cffi import FFI

# Define the C interface from the bundled header file
ffi = FFI()

# Path to the bundled header file (copied during build)
_header_path = Path(__file__).parent / "include" / "eda_core.h"


def _load_c_definitions() -> str:
    """Load C definitions from the bundled header file.

    Returns:
        C definitions suitable for cffi.cdef().

    Raises:
        RuntimeError: If header file not found.
    """
    if not _header_path.exists():
        raise RuntimeError(
            f"Header file not found at {_header_path}. "
            f"The SDK package may be corrupted or incomplete."
        )

    # Read and parse the header file
    with open(_header_path, "r") as f:
        content = f.read()

    # Extract only the parts we need (between extern "C" blocks)
    # Remove preprocessor directives, comments, and namespace declarations
    lines = []
    in_extern_c = False
    in_struct = False

    for line in content.split("\n"):
        stripped = line.strip()

        # Skip preprocessor directives and empty lines
        if stripped.startswith("#") or not stripped:
            continue

        # Skip single-line comments
        if stripped.startswith("//"):
            continue

        # Track extern "C" blocks
        if 'extern "C"' in stripped:
            in_extern_c = True
            continue

        # Track struct definitions
        if "typedef struct" in stripped:
            in_struct = True
            lines.append(line)
            continue

        # End of struct
        if in_struct and "} " in stripped and ";" in stripped:
            lines.append(line)
            in_struct = False
            continue

        # Inside struct or extern C block
        if in_struct or in_extern_c:
            # Skip closing braces of extern C
            if stripped == "}" or stripped == '}  // extern "C"':
                if not in_struct:
                    in_extern_c = False
                continue
            # Skip namespace declarations
            if "namespace" in stripped:
                continue
            # Keep the line
            if stripped:
                lines.append(line)

    return "\n".join(lines)


ffi.cdef(_load_c_definitions())

# Global library handle
_lib: Any = None
_temp_dir: Optional[str] = None


def get_lib_name() -> str:
    """Get the platform-specific library name.

    Returns:
        Library filename for the current platform.
    """
    system = platform.system()
    if system == "Windows":
        return "eda_core.dll"
    elif system == "Darwin":
        return "libeda_core.dylib"
    else:
        return "libeda_core.so"


def get_embedded_lib_path() -> Path:
    """Get the path to the embedded library for the current platform.

    Returns:
        Path to the embedded library file.

    Raises:
        RuntimeError: If platform is not supported.
    """
    system = platform.system()
    machine = platform.machine()

    # Map platform to directory name
    if system == "Linux":
        if machine in ("x86_64", "AMD64"):
            platform_dir = "linux_amd64"
        elif machine in ("aarch64", "arm64"):
            platform_dir = "linux_arm64"
        else:
            raise RuntimeError(f"Unsupported Linux architecture: {machine}")
    elif system == "Darwin":
        if machine == "x86_64":
            platform_dir = "darwin_amd64"
        elif machine == "arm64":
            platform_dir = "darwin_arm64"
        else:
            raise RuntimeError(f"Unsupported macOS architecture: {machine}")
    elif system == "Windows":
        if machine in ("x86_64", "AMD64"):
            platform_dir = "windows_amd64"
        else:
            raise RuntimeError(f"Unsupported Windows architecture: {machine}")
    else:
        raise RuntimeError(f"Unsupported operating system: {system}")

    # Get the path to the embedded library
    lib_dir = Path(__file__).parent / "libs" / platform_dir
    lib_path = lib_dir / get_lib_name()

    if not lib_path.exists():
        raise RuntimeError(
            f"Embedded library not found at {lib_path}. "
            f"Please build the library with 'make core-build-ffi' and copy it to the SDK."
        )

    return lib_path


def extract_embedded_lib() -> str:
    """Extract the embedded library to a temporary directory.

    Returns:
        Path to the extracted library file.

    Raises:
        RuntimeError: If extraction fails.
    """
    global _temp_dir

    # Create a temporary directory
    _temp_dir = tempfile.mkdtemp(prefix="eda-core-")

    # Get the embedded library path
    embedded_lib_path = get_embedded_lib_path()

    # Copy to temp directory
    lib_name = get_lib_name()
    temp_lib_path = os.path.join(_temp_dir, lib_name)

    try:
        shutil.copy2(embedded_lib_path, temp_lib_path)
        # Make it executable
        os.chmod(temp_lib_path, 0o755)
    except Exception as e:
        shutil.rmtree(_temp_dir, ignore_errors=True)
        raise RuntimeError(f"Failed to extract embedded library: {e}") from e

    return temp_lib_path


def load_library() -> Any:
    """Load the FFI library.

    Returns:
        The loaded library handle.

    Raises:
        RuntimeError: If loading fails.
    """
    global _lib

    if _lib is not None:
        return _lib

    # Extract embedded library to temp file
    lib_path = extract_embedded_lib()

    # Load the library
    try:
        _lib = ffi.dlopen(lib_path)
    except Exception as e:
        raise RuntimeError(f"Failed to load library from {lib_path}: {e}") from e

    return _lib


def cleanup() -> None:
    """Clean up temporary files."""
    global _temp_dir
    if _temp_dir and os.path.exists(_temp_dir):
        shutil.rmtree(_temp_dir, ignore_errors=True)
        _temp_dir = None


# Register cleanup on exit
atexit.register(cleanup)
