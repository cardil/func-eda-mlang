//! EDA Core - Shared multi-language library for Event-Driven Architecture
//!
//! This library provides core functionality for EDA consumers across multiple languages:
//! - Configuration management (Kafka broker, topic, consumer group)
//! - Retry logic with exponential backoff
//! - Event routing to handlers
//!
//! The library is designed to be consumed via FFI (Go, Python, Java) or WASM (JavaScript).

pub mod config;
pub mod retry;
pub mod routing;

// Re-export main types for convenience
pub use config::{get_kafka_config, KafkaConfig};
pub use retry::{calculate_backoff, should_retry};
pub use routing::route_event;

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
    fn test_retry_noop() {
        assert_eq!(should_retry("some error", 1), false);
        assert_eq!(should_retry("another error", 5), false);
    }

    #[test]
    fn test_backoff_noop() {
        assert_eq!(calculate_backoff(1), 0);
        assert_eq!(calculate_backoff(10), 0);
    }

    #[test]
    fn test_routing_noop() {
        assert_eq!(route_event("user.created"), 0);
        assert_eq!(route_event("order.placed"), 0);
    }
}
