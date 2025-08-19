// Test file demonstrating good Rust code practices

use std::collections::HashMap;
use std::error::Error;
use std::fmt;

/// Well-documented public struct following PascalCase naming
#[derive(Debug, Clone)]
pub struct UserProfile {
    /// User's unique identifier
    id: UserId,
    /// User's display name
    name: String,
    /// User's age in years
    age: u32,
}

/// Type alias following PascalCase convention
type UserId = u64;

/// Well-named trait following PascalCase
pub trait DataProcessor {
    /// Process data with proper error handling
    fn process(&self, data: &str) -> Result<String, ProcessError>;
}

/// Custom error type for better error handling
#[derive(Debug)]
pub struct ProcessError {
    message: String,
}

impl fmt::Display for ProcessError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "Process error: {}", self.message)
    }
}

impl Error for ProcessError {}

/// Well-named constant following SCREAMING_SNAKE_CASE
const MAX_RETRY_COUNT: u32 = 3;
const DEFAULT_TIMEOUT: u64 = 30;

/// Static variable with proper naming
static GLOBAL_CONFIG: &str = "production";

/// Well-structured enum following naming conventions
#[derive(Debug, PartialEq)]
pub enum RequestStatus {
    Pending,
    Approved,
    Rejected,
}

impl UserProfile {
    /// Constructor with proper documentation
    pub fn new(id: UserId, name: String, age: u32) -> Self {
        UserProfile { id, name, age }
    }

    /// Method following snake_case naming
    pub fn get_user_age(&self) -> u32 {
        self.age
    }

    /// Mutable method with proper naming
    pub fn update_name(&mut self, name: String) {
        self.name = name;
    }

    /// Method with proper error handling using Result
    pub fn validate(&self) -> Result<(), ProcessError> {
        if self.name.is_empty() {
            return Err(ProcessError {
                message: "Name cannot be empty".to_string(),
            });
        }
        if self.age > 150 {
            return Err(ProcessError {
                message: "Invalid age".to_string(),
            });
        }
        Ok(())
    }
}

/// Function with proper snake_case naming and error propagation
pub fn process_user_data(data: &str) -> Result<UserProfile, Box<dyn Error>> {
    let parts: Vec<&str> = data.split(',').collect();
    
    if parts.len() != 3 {
        return Err("Invalid data format".into());
    }
    
    let id: UserId = parts[0].parse()?;
    let name = parts[1].to_string();
    let age: u32 = parts[2].parse()?;
    
    Ok(UserProfile::new(id, name, age))
}

/// Efficient pattern matching without nesting
pub fn handle_option_elegantly(value: Option<i32>) -> i32 {
    value.unwrap_or(0)
}

/// Clean pattern matching with exhaustive handling
pub fn process_status(status: RequestStatus) -> &'static str {
    match status {
        RequestStatus::Pending => "Request is pending",
        RequestStatus::Approved => "Request approved",
        RequestStatus::Rejected => "Request rejected",
    }
}

/// Proper ownership handling without unnecessary clones
pub fn efficient_string_handling(s: &str) -> usize {
    s.len()
}

/// Using references appropriately
pub fn process_vector(data: &[i32]) -> i32 {
    data.iter().sum()
}

/// Module with proper snake_case naming
pub mod user_management {
    use super::*;
    
    /// Initialize the user management system
    pub fn init() -> Result<(), ProcessError> {
        // Initialization logic here
        Ok(())
    }
}

/// Generic function with proper type parameter naming
pub fn generic_function<T: Clone>(value: &T) -> T {
    value.clone()
}

/// Trait with associated type following conventions
pub trait Container {
    type ItemType;
    
    fn get_item(&self) -> Option<&Self::ItemType>;
}

/// Function demonstrating proper error handling with ?
pub fn read_config_file(path: &str) -> Result<String, Box<dyn Error>> {
    use std::fs;
    let contents = fs::read_to_string(path)?;
    Ok(contents.trim().to_string())
}

/// Clean iteration without cloning in loops
pub fn process_items(items: &[String]) -> Vec<usize> {
    items.iter().map(|item| item.len()).collect()
}

/// Proper lifetime elision
pub fn first_word(s: &str) -> &str {
    let bytes = s.as_bytes();
    
    for (i, &item) in bytes.iter().enumerate() {
        if item == b' ' {
            return &s[0..i];
        }
    }
    
    s
}

/// Safe function with proper bounds checking
pub fn safe_array_access(data: &[i32], index: usize) -> Option<i32> {
    data.get(index).copied()
}

/// Builder pattern with method chaining
pub struct ConfigBuilder {
    timeout: Option<u64>,
    retries: Option<u32>,
}

impl ConfigBuilder {
    pub fn new() -> Self {
        ConfigBuilder {
            timeout: None,
            retries: None,
        }
    }
    
    pub fn timeout(mut self, timeout: u64) -> Self {
        self.timeout = Some(timeout);
        self
    }
    
    pub fn retries(mut self, retries: u32) -> Self {
        self.retries = Some(retries);
        self
    }
    
    pub fn build(self) -> Config {
        Config {
            timeout: self.timeout.unwrap_or(DEFAULT_TIMEOUT),
            retries: self.retries.unwrap_or(MAX_RETRY_COUNT),
        }
    }
}

pub struct Config {
    timeout: u64,
    retries: u32,
}

#[cfg(test)]
mod tests {
    use super::*;
    
    #[test]
    fn test_user_profile_creation() {
        let user = UserProfile::new(1, "Alice".to_string(), 30);
        assert_eq!(user.get_user_age(), 30);
    }
    
    #[test]
    fn test_status_processing() {
        assert_eq!(process_status(RequestStatus::Pending), "Request is pending");
    }
}