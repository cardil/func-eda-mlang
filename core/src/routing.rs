use std::ffi::CStr;
use std::os::raw::c_char;

/// Route an event to a handler (noop for PoC)
/// Always returns handler_id 0 - no routing logic implemented yet
pub fn route_event(_event_type: &str) -> u32 {
    0
}

/// FFI-compatible function to route an event
/// Returns handler_id (currently always 0)
#[no_mangle]
pub extern "C" fn eda_route_event(event_type: *const c_char) -> u32 {
    if event_type.is_null() {
        return 0;
    }

    let type_str = unsafe {
        match CStr::from_ptr(event_type).to_str() {
            Ok(s) => s,
            Err(_) => return 0,
        }
    };

    route_event(type_str)
}

#[cfg(target_arch = "wasm32")]
use wasm_bindgen::prelude::*;

#[cfg(target_arch = "wasm32")]
#[wasm_bindgen]
pub fn route_event_wasm(event_type: &str) -> u32 {
    route_event(event_type)
}
