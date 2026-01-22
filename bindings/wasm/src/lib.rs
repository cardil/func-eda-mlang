//! WASM bindings for EDA Core
//!
//! This module provides JavaScript-compatible bindings for the EDA core library.

use wasm_bindgen::prelude::*;

// Re-export all WASM functions from the core library
pub use eda_core::config::{get_kafka_broker, get_kafka_group, get_kafka_topic};
pub use eda_core::retry::{calculate_backoff_wasm as calculate_backoff, should_retry_wasm as should_retry};
pub use eda_core::routing::route_event_wasm as route_event;

/// Initialize the WASM module (called automatically on load)
#[wasm_bindgen(start)]
pub fn initialize() {
    // Set panic hook for better error messages in browser console
    #[cfg(feature = "console_error_panic_hook")]
    console_error_panic_hook::set_once();
}
