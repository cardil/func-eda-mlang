use std::ffi::CStr;
use std::os::raw::c_char;

/// Determine if an error should be retried (noop for PoC)
/// Always returns false - no retry logic implemented yet
pub fn should_retry(_error: &str, _attempt: u32) -> bool {
    false
}

/// Calculate backoff duration in milliseconds (noop for PoC)
/// Always returns 0 - no backoff logic implemented yet
pub fn calculate_backoff(_attempt: u32) -> u64 {
    0
}

/// FFI-compatible function to check if error should be retried
/// Returns 1 for true, 0 for false
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

/// FFI-compatible function to calculate backoff duration
/// Returns backoff duration in milliseconds
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
