"""Pytest fixtures for e2e tests."""

import os
import subprocess
import pytest
from pathlib import Path

# Change to root directory
ROOT_DIR = Path(__file__).parent.parent.parent
os.chdir(ROOT_DIR)


@pytest.fixture(scope="session")
def build_project():
    """Build the entire project before running tests."""
    print("\n🔨 Building project...")
    subprocess.run(["make", "build"], check=True)
    yield


@pytest.fixture(scope="session")
def kafka_infra():
    """Start and stop Kafka infrastructure once for all tests."""
    print("\n🚀 Starting Kafka infrastructure...")
    subprocess.run(["make", "-C", "infra", "up"], check=True)
    
    yield
    
    print("\n🛑 Stopping Kafka infrastructure...")
    subprocess.run(["make", "-C", "infra", "down"], check=False)


@pytest.fixture
def send_test_events(kafka_infra):
    """Send test events to Kafka (depends on kafka_infra to ensure it's running)."""
    def _send():
        print("\n📤 Sending test events...")
        subprocess.run(["make", "-C", "infra", "send-events"], check=True)
    return _send
