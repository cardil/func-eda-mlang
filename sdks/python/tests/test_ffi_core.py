"""Unit tests for FFI Core implementation."""

import pytest
from eda_sdk.ffi.core import FFICore
from eda_sdk.types import KafkaConfig


def test_ffi_core_construction():
    """Test that FFICore can be constructed."""
    core = FFICore()
    assert core is not None
    assert core._lib is not None


def test_ffi_core_get_kafka_config():
    """Test getting Kafka configuration from FFI core."""
    core = FFICore()
    config = core.get_kafka_config()
    
    assert isinstance(config, KafkaConfig)
    assert isinstance(config.broker, str)
    assert isinstance(config.topic, str)
    assert isinstance(config.group, str)
    assert len(config.broker) > 0
    assert len(config.topic) > 0
    assert len(config.group) > 0


def test_ffi_core_calculate_backoff():
    """Test backoff calculation."""
    core = FFICore()
    
    # Test backoff for different attempts
    backoff1 = core.calculate_backoff(1)
    backoff2 = core.calculate_backoff(2)
    backoff3 = core.calculate_backoff(3)
    
    assert isinstance(backoff1, int)
    assert isinstance(backoff2, int)
    assert isinstance(backoff3, int)
    assert backoff1 >= 0
    assert backoff2 >= 0
    assert backoff3 >= 0


def test_ffi_core_close():
    """Test that close() can be called without errors."""
    core = FFICore()
    core.close()  # Should not raise any exceptions
