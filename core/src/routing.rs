use std::ffi::CStr;
use std::os::raw::c_char;

/// Destination type for event routing
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum DestinationType {
    Kafka,
    RabbitMQ,
    Http,
}

/// Event destination with type, cluster, and topic
#[derive(Debug, Clone)]
pub struct EventDestination {
    pub dest_type: DestinationType,
    pub cluster: String,
    pub topic: String,
}

/// Route an event to a destination (placeholder)
pub fn route_event(_event_type: &str) -> EventDestination {
    // TODO: For production, implement routing for return event forwarding:
    // - When a function returns an event, determine where to send it
    // - Support pattern matching for event type to destination mapping
    // - Example: "user.created" -> Kafka,cluster:default,topic:user-events
    // - Example: "order.*" -> Kafka,cluster:default,topic:order-events
    // - Example: error events -> Kafka,cluster:default,topic:dlq
    // - Return full destination specification (type, cluster, topic)
    // - Record telemetry span for routing decision
    EventDestination {
        dest_type: DestinationType::Kafka,
        cluster: "default".to_string(),
        topic: "events".to_string(),
    }
}

// FFI exports

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

    // For FFI, just return destination type as u32
    let dest = route_event(type_str);
    match dest.dest_type {
        DestinationType::Kafka => 0,
        DestinationType::RabbitMQ => 1,
        DestinationType::Http => 2,
    }
}

#[cfg(target_arch = "wasm32")]
use wasm_bindgen::prelude::*;

#[cfg(target_arch = "wasm32")]
#[wasm_bindgen]
pub fn route_event_wasm(event_type: &str) -> u32 {
    let dest = route_event(event_type);
    match dest.dest_type {
        DestinationType::Kafka => 0,
        DestinationType::RabbitMQ => 1,
        DestinationType::Http => 2,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_routing_placeholder() {
        let result = route_event("user.created");
        assert_eq!(result.dest_type, DestinationType::Kafka);
        assert_eq!(result.cluster, "default");
        assert_eq!(result.topic, "events");
    }
}
