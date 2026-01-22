"""E2E tests for all examples."""

import subprocess
import time
import pytest
import os
from pathlib import Path


def run_example_with_monitoring(command: list[str], expect_output: bool = False) -> tuple[bool, str]:
    """
    Run an example and monitor output in real-time, terminating early when success criteria met.
    
    Args:
        command: Command to run
        expect_output: Whether to expect output events (for output examples)
    
    Returns:
        tuple: (success, output) where success is True if expected output found
    """
    # Override timeout for tests and ensure Python is unbuffered
    env = os.environ.copy()
    env['EXAMPLE_TIMEOUT'] = '10'
    env['PYTHONUNBUFFERED'] = '1'
    
    # Run process and capture output
    process = subprocess.Popen(
        command,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
        env=env,
        bufsize=1
    )
    
    output_lines = []
    received_count = 0
    published_count = 0
    
    try:
        # Read output line by line
        for line in process.stdout:
            output_lines.append(line)
            
            # Count success indicators
            if "📨 Received event:" in line:
                received_count += 1
            if "Published output event to" in line:
                published_count += 1
            
            # Check if we've met success criteria
            if expect_output:
                if received_count >= 10 and published_count >= 10:
                    # Success! Terminate the process gracefully
                    # Give enough time for Kafka consumer group LeaveGroup protocol
                    process.terminate()
                    try:
                        process.wait(timeout=60)
                    except subprocess.TimeoutExpired:
                        process.kill()
                        process.wait()
                    break
            else:
                if received_count >= 10:
                    # Success! Terminate the process gracefully
                    # Give enough time for Kafka consumer group LeaveGroup protocol
                    process.terminate()
                    try:
                        process.wait(timeout=60)
                    except subprocess.TimeoutExpired:
                        process.kill()
                        process.wait()
                    break
        
        # Wait for process to complete if it hasn't already
        process.wait(timeout=60)
        
    except Exception:
        process.kill()
        process.wait()
    
    output = ''.join(output_lines)
    
    # Final check
    if expect_output:
        success = received_count >= 10 and published_count >= 10
    else:
        success = received_count >= 10
    
    return success, output


@pytest.mark.e2e
def test_go_ffi_example(build_project, send_test_events):
    """Test Go FFI example."""
    send_test_events()
    time.sleep(2)  # Give Kafka time to have events ready
    
    success, output = run_example_with_monitoring(
        ["make", "-C", "sdks/go/examples/ffi-example", "run"],
        expect_output=False
    )
    
    print(f"\n📋 Go FFI Example Output:\n{output}")
    assert success, "Go FFI example did not process 10+ events successfully"


@pytest.mark.e2e
def test_go_ffi_output_example(build_project, send_test_events):
    """Test Go FFI output example."""
    send_test_events()
    time.sleep(2)
    
    success, output = run_example_with_monitoring(
        ["make", "-C", "sdks/go/examples/ffi-output-example", "run"],
        expect_output=True
    )
    
    print(f"\n📋 Go FFI Output Example Output:\n{output}")
    assert success, "Go FFI output example did not process 10+ events and publish 10+ outputs successfully"


@pytest.mark.e2e
def test_python_ffi_example(build_project, send_test_events):
    """Test Python FFI example."""
    send_test_events()
    time.sleep(2)
    
    success, output = run_example_with_monitoring(
        ["make", "-C", "sdks/python/examples/ffi", "run"],
        expect_output=False
    )
    
    print(f"\n📋 Python FFI Example Output:\n{output}")
    assert success, "Python FFI example did not process 10+ events successfully"


@pytest.mark.e2e
def test_python_ffi_output_example(build_project, send_test_events):
    """Test Python FFI output example."""
    send_test_events()
    time.sleep(2)
    
    success, output = run_example_with_monitoring(
        ["make", "-C", "sdks/python/examples/ffi-output", "run"],
        expect_output=True
    )
    
    print(f"\n📋 Python FFI Output Example Output:\n{output}")
    assert success, "Python FFI output example did not process 10+ events and publish 10+ outputs successfully"
