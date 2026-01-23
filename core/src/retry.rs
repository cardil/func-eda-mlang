use std::ffi::CStr;
use std::os::raw::c_char;

/// Error categories for retry decision making
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ErrorCategory {
    Transient,
    Permanent,
    Unknown,
}

/// Retry decision with backoff and DLQ information
#[derive(Debug, Clone)]
pub struct RetryDecision {
    pub should_retry: bool,
    pub backoff_ms: u64,
    pub send_to_dlq: bool,
}

/// Classify an error message into a category (placeholder)
pub fn classify_error(_error_message: &str) -> ErrorCategory {
    // TODO: For production, implement pattern matching on error messages:
    // - Transient: "timeout", "connection refused", "503", "429", "rate limit"
    // - Permanent: "validation", "unauthorized", "404", "400", "invalid"
    // - Unknown: everything else (conservative retry)
    ErrorCategory::Unknown
}

/// Get retry decision based on error category and attempt number (placeholder)
pub fn get_retry_decision(
    _error_category: ErrorCategory,
    _attempt: u32,
    _max_attempts: u32,
) -> RetryDecision {
    // TODO: For production, implement retry logic:
    // - Permanent errors: never retry, send to DLQ immediately
    // - Transient/Unknown: retry if attempt < max_attempts
    // - Calculate backoff using exponential backoff with jitter
    // - Send to DLQ when max attempts exceeded
    // - Record telemetry span for retry decision
    RetryDecision {
        should_retry: false,
        backoff_ms: 0,
        send_to_dlq: false,
    }
}

/// Calculate exponential backoff (placeholder)
pub fn calculate_backoff(_attempt: u32) -> u64 {
    // TODO: For production, implement exponential backoff with jitter:
    // - Formula: min(base * 2^attempt * (1 + jitter), max_backoff)
    // - base: 100ms, max_backoff: 30000ms (30 seconds)
    // - jitter: random value between -0.25 and +0.25 to avoid thundering herd
    // - Record telemetry span for backoff calculation
    0
}

/// Legacy function for backward compatibility
pub fn should_retry(_error: &str, _attempt: u32) -> bool {
    false
}

// FFI exports

#[allow(clippy::not_unsafe_ptr_arg_deref)]
#[no_mangle]
pub extern "C" fn eda_classify_error(error: *const c_char) -> u32 {
    if error.is_null() {
        return 2; // Unknown
    }

    let error_str = unsafe {
        match CStr::from_ptr(error).to_str() {
            Ok(s) => s,
            Err(_) => return 2,
        }
    };

    match classify_error(error_str) {
        ErrorCategory::Transient => 0,
        ErrorCategory::Permanent => 1,
        ErrorCategory::Unknown => 2,
    }
}

#[no_mangle]
pub extern "C" fn eda_get_retry_decision(
    error_category: u32,
    attempt: u32,
    max_attempts: u32,
) -> u64 {
    let category = match error_category {
        0 => ErrorCategory::Transient,
        1 => ErrorCategory::Permanent,
        _ => ErrorCategory::Unknown,
    };

    let decision = get_retry_decision(category, attempt, max_attempts);

    let mut result: u64 = decision.backoff_ms & 0xFFFFFFFF;
    if decision.should_retry {
        result |= 1u64 << 32;
    }
    if decision.send_to_dlq {
        result |= 1u64 << 33;
    }

    result
}

#[allow(clippy::not_unsafe_ptr_arg_deref)]
#[no_mangle]
pub extern "C" fn eda_should_retry(error: *const c_char, attempt: u32) -> i32 {
    if error.is_null() {
        return 0;
    }

    let error_str = unsafe {
        match CStr::from_ptr(error).to_str() {
            Ok(s) => s,
            Err(_) => return 0,
        }
    };

    if should_retry(error_str, attempt) {
        1
    } else {
        0
    }
}

#[no_mangle]
pub extern "C" fn eda_calculate_backoff(attempt: u32) -> u64 {
    calculate_backoff(attempt)
}

#[cfg(target_arch = "wasm32")]
use wasm_bindgen::prelude::*;

#[cfg(target_arch = "wasm32")]
#[wasm_bindgen]
pub fn should_retry_wasm(error: &str, attempt: u32) -> bool {
    should_retry(error, attempt)
}

#[cfg(target_arch = "wasm32")]
#[wasm_bindgen]
pub fn calculate_backoff_wasm(attempt: u32) -> u64 {
    calculate_backoff(attempt)
}

#[cfg(target_arch = "wasm32")]
#[wasm_bindgen]
pub fn classify_error_wasm(error: &str) -> u32 {
    match classify_error(error) {
        ErrorCategory::Transient => 0,
        ErrorCategory::Permanent => 1,
        ErrorCategory::Unknown => 2,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_classify_error_placeholder() {
        assert_eq!(classify_error("any error"), ErrorCategory::Unknown);
    }

    #[test]
    fn test_retry_decision_placeholder() {
        let decision = get_retry_decision(ErrorCategory::Transient, 1, 5);
        assert!(!decision.should_retry);
        assert_eq!(decision.backoff_ms, 0);
    }

    #[test]
    fn test_backoff_placeholder() {
        assert_eq!(calculate_backoff(1), 0);
        assert_eq!(calculate_backoff(10), 0);
    }
}
