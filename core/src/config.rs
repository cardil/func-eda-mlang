use serde::{Deserialize, Serialize};
use std::ffi::CString;
use std::os::raw::c_char;

/// Kafka configuration for EDA consumers
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct KafkaConfig {
    pub broker: String,
    pub topic: String,
    pub group: String,
}

impl Default for KafkaConfig {
    fn default() -> Self {
        Self {
            broker: "localhost:9092".to_string(),
            topic: "events".to_string(),
            group: "poc".to_string(),
        }
    }
}

/// Get default Kafka configuration (static/mock implementation for PoC)
pub fn get_kafka_config() -> KafkaConfig {
    KafkaConfig::default()
}

/// FFI-compatible function to get Kafka broker
/// Returns a C string that must be freed by the caller
#[no_mangle]
pub extern "C" fn eda_get_kafka_broker() -> *mut c_char {
    let config = get_kafka_config();
    match CString::new(config.broker) {
        Ok(c_str) => c_str.into_raw(),
        Err(_) => std::ptr::null_mut(),
    }
}

/// FFI-compatible function to get Kafka topic
/// Returns a C string that must be freed by the caller
#[no_mangle]
pub extern "C" fn eda_get_kafka_topic() -> *mut c_char {
    let config = get_kafka_config();
    match CString::new(config.topic) {
        Ok(c_str) => c_str.into_raw(),
        Err(_) => std::ptr::null_mut(),
    }
}

/// FFI-compatible function to get Kafka consumer group
/// Returns a C string that must be freed by the caller
#[no_mangle]
pub extern "C" fn eda_get_kafka_group() -> *mut c_char {
    let config = get_kafka_config();
    match CString::new(config.group) {
        Ok(c_str) => c_str.into_raw(),
        Err(_) => std::ptr::null_mut(),
    }
}

/// FFI-compatible function to free C strings returned by this library
#[allow(clippy::not_unsafe_ptr_arg_deref)]
#[no_mangle]
pub extern "C" fn eda_free_string(s: *mut c_char) {
    if !s.is_null() {
        unsafe {
            let _ = CString::from_raw(s);
        }
    }
}

#[cfg(target_arch = "wasm32")]
use wasm_bindgen::prelude::*;

#[cfg(target_arch = "wasm32")]
#[wasm_bindgen]
pub fn get_kafka_broker() -> String {
    get_kafka_config().broker
}

#[cfg(target_arch = "wasm32")]
#[wasm_bindgen]
pub fn get_kafka_topic() -> String {
    get_kafka_config().topic
}

#[cfg(target_arch = "wasm32")]
#[wasm_bindgen]
pub fn get_kafka_group() -> String {
    get_kafka_config().group
}
