//! EDA Core - Shared multi-language library for Event-Driven Architecture
//!
//! This library provides core functionality for EDA consumers across multiple languages:
//! - Configuration management (Kafka broker, topic, consumer group)
//! - Retry logic with error classification and backoff
//! - Event routing to destinations
//! - Telemetry recording
//!
//! The library is designed to be consumed via FFI (Go, Python, Java) or WASM (JavaScript).

pub mod config;
pub mod retry;
pub mod routing;
pub mod telemetry;

#[cfg(target_arch = "wasm32")]
pub mod wit_bindings;

// Re-export main types for convenience
pub use config::{get_kafka_config, KafkaConfig};
pub use retry::{calculate_backoff, classify_error, get_retry_decision, should_retry, ErrorCategory, RetryDecision};
pub use routing::{
    get_output_destination, add_routing_rule, clear_routing_rules,
    get_default_destination, set_default_destination, load_routing_config,
    DestinationType, OutputDestination, RoutingRule, FilterExpression
};
pub use telemetry::{get_event_count, record_event_processed, record_event_received, record_retry_attempt};

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_config() {
        let config = get_kafka_config();
        assert_eq!(config.broker, "localhost:9092");
        assert_eq!(config.topic, "events");
        assert_eq!(config.group, "poc");
    }

    #[test]
    fn test_retry_placeholder() {
        assert_eq!(should_retry("some error", 1), false);
        assert_eq!(should_retry("another error", 5), false);
    }

    #[test]
    fn test_backoff_placeholder() {
        assert_eq!(calculate_backoff(1), 0);
        assert_eq!(calculate_backoff(10), 0);
    }

    #[test]
    fn test_routing_default() {
        // Reset routing state to ensure clean test
        clear_routing_rules();
        set_default_destination(OutputDestination {
            dest_type: DestinationType::Kafka,
            target: "events".to_string(),
            cluster: Some("default".to_string()),
        });
        
        let event_json = r#"{"specversion":"1.0","type":"user.created","source":"test","id":"1"}"#;
        let dest = get_output_destination(event_json);
        assert_eq!(dest.dest_type, DestinationType::Kafka);
        assert_eq!(dest.target, "events");
    }

    #[test]
    fn test_telemetry_counting() {
        let before = get_event_count();
        record_event_received("test.event");
        let after = get_event_count();
        assert!(after > before);
    }
}
