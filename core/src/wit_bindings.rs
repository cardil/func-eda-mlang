//! WIT bindings for WASM Component Model
//!
//! This module implements the WIT world defined in wit/eda-core.wit
//! It bridges the WIT interface to our Rust implementation.

#[cfg(target_arch = "wasm32")]
wit_bindgen::generate!({
    world: "eda-core",
    path: "../wit"
});

#[cfg(target_arch = "wasm32")]
use crate::{config, retry, routing, telemetry};

#[cfg(target_arch = "wasm32")]
use exports::eda::core::{config::Guest as ConfigGuest, retry::Guest as RetryGuest, routing::Guest as RoutingGuest, telemetry::Guest as TelemetryGuest};

#[cfg(target_arch = "wasm32")]
use eda::core::types::*;

#[cfg(target_arch = "wasm32")]
struct Component;

#[cfg(target_arch = "wasm32")]
impl ConfigGuest for Component {
    fn get_kafka_config() -> KafkaConfig {
        let cfg = config::get_kafka_config();
        KafkaConfig {
            broker: cfg.broker,
            topic: cfg.topic,
            group_id: cfg.group,
        }
    }
}

#[cfg(target_arch = "wasm32")]
impl RetryGuest for Component {
    fn classify_error(error_message: String) -> ErrorCategory {
        match retry::classify_error(&error_message) {
            retry::ErrorCategory::Transient => ErrorCategory::Transient,
            retry::ErrorCategory::Permanent => ErrorCategory::Permanent,
            retry::ErrorCategory::Unknown => ErrorCategory::Unknown,
        }
    }

    fn get_retry_decision(
        error_category: ErrorCategory,
        attempt: u32,
        max_attempts: u32,
    ) -> RetryDecision {
        let cat = match error_category {
            ErrorCategory::Transient => retry::ErrorCategory::Transient,
            ErrorCategory::Permanent => retry::ErrorCategory::Permanent,
            ErrorCategory::Unknown => retry::ErrorCategory::Unknown,
        };

        let decision = retry::get_retry_decision(cat, attempt, max_attempts);
        RetryDecision {
            should_retry: decision.should_retry,
            backoff_ms: decision.backoff_ms,
            send_to_dlq: decision.send_to_dlq,
        }
    }
}

#[cfg(target_arch = "wasm32")]
impl RoutingGuest for Component {
    fn route_event(event_type: String) -> EventDestination {
        let dest = routing::route_event(&event_type);
        EventDestination {
            dest_type: match dest.dest_type {
                routing::DestinationType::Kafka => DestinationType::Kafka,
                routing::DestinationType::RabbitMQ => DestinationType::Rabbitmq,
                routing::DestinationType::Http => DestinationType::Http,
            },
            cluster: dest.cluster,
            topic: dest.topic,
        }
    }
}

#[cfg(target_arch = "wasm32")]
impl TelemetryGuest for Component {
    fn record_event_received(event_type: String) {
        telemetry::record_event_received(&event_type);
    }

    fn record_event_processed(event_type: String, success: bool, duration_ms: u64) {
        telemetry::record_event_processed(&event_type, success, duration_ms);
    }

    fn record_retry_attempt(attempt: u32, backoff_ms: u64) {
        telemetry::record_retry_attempt(attempt, backoff_ms);
    }

    fn get_event_count() -> u64 {
        telemetry::get_event_count()
    }
}

#[cfg(target_arch = "wasm32")]
export!(Component);
