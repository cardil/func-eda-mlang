"""FFI implementation of the EDA SDK using cffi."""

from .core import FFICore
from .run import run

__all__ = ["FFICore", "run"]
