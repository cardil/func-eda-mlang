//! WASM Component bindings for EDA Core
//!
//! This module provides WASM Component Model bindings using WIT interfaces.
//! The WIT bindings are defined in the core library.

// Re-export the WIT bindings from core
// The actual implementation is in core/src/wit_bindings.rs
pub use eda_core::wit_bindings::*;
