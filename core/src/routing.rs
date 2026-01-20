use std::ffi::CStr;
use std::os::raw::c_char;
use std::sync::RwLock;
use std::fs;
use cloudevents::{Event, AttributesReader};
use serde::Deserialize;
use serde_json::Value;

// TODO: Add CESQL (CloudEvents SQL) support for advanced filtering
// The CloudEvents Rust SDK (v0.9) doesn't yet support CESQL filtering.
// Consider:
// 1. Checking if there's an open issue in cloudevents/sdk-rust for CESQL support
// 2. If not, create an issue requesting CESQL implementation
// 3. For now, we implement basic filter dialects (exact, prefix, suffix, all, any, not)
// 4. CESQL would enable complex queries like: "type LIKE 'com.example.%' AND EXISTS priority"
// Reference: https://github.com/cloudevents/spec/blob/main/cesql/spec.md

/// Destination type for output event routing
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum DestinationType {
    Kafka,
    RabbitMQ,
    Http,
    Discard,
}

/// Output destination with type, target, and optional cluster
#[derive(Debug, Clone)]
pub struct OutputDestination {
    pub dest_type: DestinationType,
    pub target: String,
    pub cluster: Option<String>,
}

/// Filter expression for routing rules (JSON-serialized CloudEvents Subscriptions API format)
pub type FilterExpression = String;

/// Routing rule with name, filter, and destination
#[derive(Debug, Clone)]
pub struct RoutingRule {
    pub name: String,
    pub filter: FilterExpression,
    pub destination: OutputDestination,
}

/// Global routing state
static ROUTING_RULES: RwLock<Vec<RoutingRule>> = RwLock::new(Vec::new());
static DEFAULT_DESTINATION: RwLock<Option<OutputDestination>> = RwLock::new(None);

/// Evaluate a filter expression against a CloudEvent
///
/// Implements CloudEvents Subscriptions API filter dialects:
/// - exact: exact match on attributes
/// - prefix: prefix match on attributes
/// - suffix: suffix match on attributes
/// - all: all nested filters must match
/// - any: at least one nested filter must match
/// - not: negates nested filter
fn evaluate_filter(event: &Event, filter_json: &str) -> bool {
    let filter: Value = match serde_json::from_str(filter_json) {
        Ok(f) => f,
        Err(_) => return false,
    };

    evaluate_filter_value(event, &filter)
}

fn evaluate_filter_value(event: &Event, filter: &Value) -> bool {
    if let Some(obj) = filter.as_object() {
        // Check for filter dialect keywords
        if let Some(exact) = obj.get("exact") {
            return evaluate_exact(event, exact);
        }
        if let Some(prefix) = obj.get("prefix") {
            return evaluate_prefix(event, prefix);
        }
        if let Some(suffix) = obj.get("suffix") {
            return evaluate_suffix(event, suffix);
        }
        if let Some(all) = obj.get("all") {
            return evaluate_all(event, all);
        }
        if let Some(any) = obj.get("any") {
            return evaluate_any(event, any);
        }
        if let Some(not) = obj.get("not") {
            return !evaluate_filter_value(event, not);
        }
    }
    false
}

fn evaluate_exact(event: &Event, exact: &Value) -> bool {
    if let Some(obj) = exact.as_object() {
        for (key, value) in obj {
            let event_value = get_event_attribute(event, key);
            if event_value != value.as_str().unwrap_or("") {
                return false;
            }
        }
        return true;
    }
    false
}

fn evaluate_prefix(event: &Event, prefix: &Value) -> bool {
    if let Some(obj) = prefix.as_object() {
        for (key, value) in obj {
            let event_value = get_event_attribute(event, key);
            let prefix_str = value.as_str().unwrap_or("");
            if !event_value.starts_with(prefix_str) {
                return false;
            }
        }
        return true;
    }
    false
}

fn evaluate_suffix(event: &Event, suffix: &Value) -> bool {
    if let Some(obj) = suffix.as_object() {
        for (key, value) in obj {
            let event_value = get_event_attribute(event, key);
            let suffix_str = value.as_str().unwrap_or("");
            if !event_value.ends_with(suffix_str) {
                return false;
            }
        }
        return true;
    }
    false
}

fn evaluate_all(event: &Event, all: &Value) -> bool {
    if let Some(arr) = all.as_array() {
        for filter in arr {
            if !evaluate_filter_value(event, filter) {
                return false;
            }
        }
        return true;
    }
    false
}

fn evaluate_any(event: &Event, any: &Value) -> bool {
    if let Some(arr) = any.as_array() {
        for filter in arr {
            if evaluate_filter_value(event, filter) {
                return true;
            }
        }
        return false;
    }
    false
}

fn get_event_attribute(event: &Event, key: &str) -> String {
    match key {
        "type" => event.ty().to_string(),
        "source" => event.source().to_string(),
        "id" => event.id().to_string(),
        "subject" => event.subject().unwrap_or("").to_string(),
        _ => {
            // Try to get from extensions
            if let Some(ext) = event.extension(key) {
                ext.to_string()
            } else {
                String::new()
            }
        }
    }
}

/// Get the output destination for an event based on routing rules
///
/// Routes output events from handlers to their destinations based on configured rules.
/// This is NOT for routing incoming events to handlers - it's for routing handler output.
///
/// # Arguments
/// * `event_json` - Serialized CloudEvent in JSON format
///
/// # Returns
/// The destination where the output event should be published
pub fn get_output_destination(event_json: &str) -> OutputDestination {
    // Parse the CloudEvent
    let event: Event = match serde_json::from_str(event_json) {
        Ok(e) => e,
        Err(_) => return get_default_destination(),
    };
    
    let rules = ROUTING_RULES.read().unwrap();
    
    // Evaluate each rule's filter against the event
    for rule in rules.iter() {
        if evaluate_filter(&event, &rule.filter) {
            return rule.destination.clone();
        }
    }
    
    // No matching rule, return default destination
    get_default_destination()
}

/// Add a routing rule
pub fn add_routing_rule(rule: RoutingRule) {
    let mut rules = ROUTING_RULES.write().unwrap();
    rules.push(rule);
}

/// Clear all routing rules
pub fn clear_routing_rules() {
    let mut rules = ROUTING_RULES.write().unwrap();
    rules.clear();
}

/// Get the default destination used when no rule matches
pub fn get_default_destination() -> OutputDestination {
    let default = DEFAULT_DESTINATION.read().unwrap();
    default.clone().unwrap_or_else(|| {
        // Fallback: return default Kafka destination
        OutputDestination {
            dest_type: DestinationType::Kafka,
            target: "events".to_string(),
            cluster: Some("default".to_string()),
        }
    })
}

/// Set the default destination for unmatched events
pub fn set_default_destination(dest: OutputDestination) {
    let mut default = DEFAULT_DESTINATION.write().unwrap();
    *default = Some(dest);
}

// YAML configuration structures
#[derive(Debug, Deserialize)]
struct RoutingConfig {
    routing: RoutingConfigInner,
}

#[derive(Debug, Deserialize)]
struct RoutingConfigInner {
    default: Option<DestinationConfig>,
    rules: Option<Vec<RuleConfig>>,
}

#[derive(Debug, Deserialize)]
struct RuleConfig {
    name: String,
    filter: Value,
    destination: DestinationConfig,
}

#[derive(Debug, Deserialize)]
struct DestinationConfig {
    #[serde(rename = "type")]
    dest_type: String,
    #[serde(default)]
    target: String,
    cluster: Option<String>,
}

/// Load routing configuration from a YAML file
pub fn load_routing_config(file_path: &str) -> Result<(), String> {
    // Read the YAML file
    let yaml_content = fs::read_to_string(file_path)
        .map_err(|e| format!("Failed to read routing config file: {}", e))?;
    
    // Parse YAML
    let config: RoutingConfig = serde_yaml::from_str(&yaml_content)
        .map_err(|e| format!("Failed to parse routing config YAML: {}", e))?;
    
    // Set default destination if provided
    if let Some(default_config) = config.routing.default {
        let dest_type = parse_destination_type(&default_config.dest_type);
        let default_dest = OutputDestination {
            dest_type,
            target: default_config.target,
            cluster: default_config.cluster,
        };
        set_default_destination(default_dest);
    }
    
    // Add routing rules if provided
    if let Some(rules) = config.routing.rules {
        for rule_config in rules {
            let dest_type = parse_destination_type(&rule_config.destination.dest_type);
            let destination = OutputDestination {
                dest_type,
                target: rule_config.destination.target,
                cluster: rule_config.destination.cluster,
            };
            
            let rule = RoutingRule {
                name: rule_config.name,
                filter: rule_config.filter.to_string(),
                destination,
            };
            
            add_routing_rule(rule);
        }
    }
    
    Ok(())
}

fn parse_destination_type(type_str: &str) -> DestinationType {
    match type_str.to_lowercase().as_str() {
        "kafka" => DestinationType::Kafka,
        "rabbitmq" | "amqp" => DestinationType::RabbitMQ,
        "http" | "https" => DestinationType::Http,
        "discard" => DestinationType::Discard,
        _ => DestinationType::Kafka,
    }
}

// FFI exports

/// FFI-compatible output destination structure
#[repr(C)]
pub struct COutputDestination {
    pub dest_type: u32,
    pub target: *mut c_char,
    pub cluster: *mut c_char,
}

use std::ffi::CString;

#[no_mangle]
pub extern "C" fn eda_get_output_destination(event_json: *const c_char) -> *mut COutputDestination {
    if event_json.is_null() {
        return std::ptr::null_mut();
    }

    let json_str = unsafe {
        match CStr::from_ptr(event_json).to_str() {
            Ok(s) => s,
            Err(_) => return std::ptr::null_mut(),
        }
    };

    let dest = get_output_destination(json_str);
    
    let dest_type = match dest.dest_type {
        DestinationType::Kafka => 0,
        DestinationType::RabbitMQ => 1,
        DestinationType::Http => 2,
        DestinationType::Discard => 3,
    };
    
    let target = CString::new(dest.target).unwrap().into_raw();
    let cluster = dest.cluster
        .map(|c| CString::new(c).unwrap().into_raw())
        .unwrap_or(std::ptr::null_mut());
    
    Box::into_raw(Box::new(COutputDestination {
        dest_type,
        target,
        cluster,
    }))
}

#[no_mangle]
pub extern "C" fn eda_free_output_destination(dest: *mut COutputDestination) {
    if dest.is_null() {
        return;
    }
    
    unsafe {
        let dest_box = Box::from_raw(dest);
        if !dest_box.target.is_null() {
            let _ = CString::from_raw(dest_box.target);
        }
        if !dest_box.cluster.is_null() {
            let _ = CString::from_raw(dest_box.cluster);
        }
    }
}

/// Load routing configuration from a YAML file via FFI
#[no_mangle]
pub extern "C" fn eda_load_routing_config(file_path: *const c_char) -> bool {
    if file_path.is_null() {
        return false;
    }

    let path_str = unsafe {
        match CStr::from_ptr(file_path).to_str() {
            Ok(s) => s,
            Err(_) => return false,
        }
    };

    load_routing_config(path_str).is_ok()
}

#[cfg(target_arch = "wasm32")]
use wasm_bindgen::prelude::*;

#[cfg(target_arch = "wasm32")]
#[wasm_bindgen]
pub fn get_output_destination_wasm(event_json: &str) -> u32 {
    let dest = get_output_destination(event_json);
    match dest.dest_type {
        DestinationType::Kafka => 0,
        DestinationType::RabbitMQ => 1,
        DestinationType::Http => 2,
        DestinationType::Discard => 3,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    // Helper to reset state before each test
    fn reset_routing_state() {
        clear_routing_rules();
        // Reset default destination to initial state
        let default = OutputDestination {
            dest_type: DestinationType::Kafka,
            target: "events".to_string(),
            cluster: Some("default".to_string()),
        };
        set_default_destination(default);
    }

    #[test]
    fn test_default_destination() {
        reset_routing_state();
        let result = get_output_destination(r#"{"specversion":"1.0","type":"test","source":"test","id":"1"}"#);
        assert_eq!(result.dest_type, DestinationType::Kafka);
        assert_eq!(result.target, "events");
        assert_eq!(result.cluster, Some("default".to_string()));
    }

    #[test]
    fn test_set_default_destination() {
        reset_routing_state();
        let custom_dest = OutputDestination {
            dest_type: DestinationType::Http,
            target: "https://example.com/webhook".to_string(),
            cluster: None,
        };
        set_default_destination(custom_dest);
        
        let result = get_default_destination();
        assert_eq!(result.dest_type, DestinationType::Http);
        assert_eq!(result.target, "https://example.com/webhook");
        assert_eq!(result.cluster, None);
    }

    #[test]
    fn test_exact_filter_match() {
        reset_routing_state();
        
        let rule = RoutingRule {
            name: "test-rule".to_string(),
            filter: r#"{"exact":{"type":"com.example.test"}}"#.to_string(),
            destination: OutputDestination {
                dest_type: DestinationType::Kafka,
                target: "test-topic".to_string(),
                cluster: Some("test-cluster".to_string()),
            },
        };
        
        add_routing_rule(rule);
        
        let event_json = r#"{"specversion":"1.0","type":"com.example.test","source":"test","id":"1"}"#;
        let result = get_output_destination(event_json);
        assert_eq!(result.dest_type, DestinationType::Kafka);
        assert_eq!(result.target, "test-topic");
    }

    #[test]
    fn test_prefix_filter_match() {
        reset_routing_state();
        
        let rule = RoutingRule {
            name: "prefix-rule".to_string(),
            filter: r#"{"prefix":{"type":"com.example."}}"#.to_string(),
            destination: OutputDestination {
                dest_type: DestinationType::Kafka,
                target: "example-events".to_string(),
                cluster: Some("default".to_string()),
            },
        };
        
        add_routing_rule(rule);
        
        let event_json = r#"{"specversion":"1.0","type":"com.example.order.created","source":"test","id":"1"}"#;
        let result = get_output_destination(event_json);
        assert_eq!(result.dest_type, DestinationType::Kafka);
        assert_eq!(result.target, "example-events");
    }

    #[test]
    fn test_suffix_filter_match() {
        reset_routing_state();
        
        let rule = RoutingRule {
            name: "suffix-rule".to_string(),
            filter: r#"{"suffix":{"type":".created"}}"#.to_string(),
            destination: OutputDestination {
                dest_type: DestinationType::Kafka,
                target: "created-events".to_string(),
                cluster: Some("default".to_string()),
            },
        };
        
        add_routing_rule(rule);
        
        let event_json = r#"{"specversion":"1.0","type":"order.created","source":"test","id":"1"}"#;
        let result = get_output_destination(event_json);
        assert_eq!(result.dest_type, DestinationType::Kafka);
        assert_eq!(result.target, "created-events");
    }

    #[test]
    fn test_no_match_returns_default() {
        reset_routing_state();
        
        let rule = RoutingRule {
            name: "specific-rule".to_string(),
            filter: r#"{"exact":{"type":"specific.type"}}"#.to_string(),
            destination: OutputDestination {
                dest_type: DestinationType::Http,
                target: "http://example.com".to_string(),
                cluster: None,
            },
        };
        
        add_routing_rule(rule);
        
        let event_json = r#"{"specversion":"1.0","type":"different.type","source":"test","id":"1"}"#;
        let result = get_output_destination(event_json);
        // Should return default destination
        assert_eq!(result.dest_type, DestinationType::Kafka);
        assert_eq!(result.target, "events");
    }
}
