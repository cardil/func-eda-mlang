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
pub use routing::{route_event, DestinationType, EventDestination};
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
    fn test_routing_placeholder() {
        let dest = route_event("user.created");
        assert_eq!(dest.dest_type, DestinationType::Kafka);
    }

    #[test]
    fn test_telemetry_counting() {
        let before = get_event_count();
        record_event_received("test.event");
        let after = get_event_count();
        assert!(after > before);
    }
}
