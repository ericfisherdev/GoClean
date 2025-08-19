// Test file for code duplication violations

// Violation: Duplicated function logic
fn calculate_area_rectangle(width: f64, height: f64) -> f64 {
    if width <= 0.0 || height <= 0.0 {
        return 0.0;
    }
    let area = width * height;
    println!("Rectangle area: {}", area);
    area
}

// Violation: Nearly identical function (duplicated logic)
fn calculate_area_square(side: f64) -> f64 {
    if side <= 0.0 {
        return 0.0;
    }
    let area = side * side;
    println!("Square area: {}", area);
    area
}

// Violation: Duplicated error handling pattern
fn read_config_file(path: &str) -> Result<String, String> {
    use std::fs;
    match fs::read_to_string(path) {
        Ok(content) => {
            if content.is_empty() {
                Err("File is empty".to_string())
            } else {
                Ok(content)
            }
        }
        Err(e) => Err(format!("Failed to read file: {}", e))
    }
}

// Violation: Same error handling pattern duplicated
fn read_data_file(path: &str) -> Result<String, String> {
    use std::fs;
    match fs::read_to_string(path) {
        Ok(content) => {
            if content.is_empty() {
                Err("File is empty".to_string())
            } else {
                Ok(content)
            }
        }
        Err(e) => Err(format!("Failed to read file: {}", e))
    }
}

// Violation: Duplicated validation logic
struct User {
    name: String,
    email: String,
    age: u32,
}

impl User {
    fn validate_name(&self) -> bool {
        !self.name.is_empty() && 
        self.name.len() >= 2 && 
        self.name.len() <= 50 &&
        self.name.chars().all(|c| c.is_alphabetic() || c.is_whitespace())
    }
    
    fn validate_email(&self) -> bool {
        !self.email.is_empty() && 
        self.email.contains('@') &&
        self.email.len() >= 5 &&
        self.email.len() <= 100
    }
}

// Violation: Duplicated validation logic in another struct
struct Employee {
    name: String,
    email: String,
    department: String,
}

impl Employee {
    // Duplicate of User::validate_name
    fn validate_name(&self) -> bool {
        !self.name.is_empty() && 
        self.name.len() >= 2 && 
        self.name.len() <= 50 &&
        self.name.chars().all(|c| c.is_alphabetic() || c.is_whitespace())
    }
    
    // Duplicate of User::validate_email
    fn validate_email(&self) -> bool {
        !self.email.is_empty() && 
        self.email.contains('@') &&
        self.email.len() >= 5 &&
        self.email.len() <= 100
    }
}

// Violation: Duplicated match arms
fn process_status_code(code: i32) -> String {
    match code {
        200 => {
            println!("Success");
            "OK".to_string()
        }
        201 => {
            println!("Success");
            "Created".to_string()
        }
        204 => {
            println!("Success");
            "No Content".to_string()
        }
        400 => {
            println!("Client Error");
            "Bad Request".to_string()
        }
        401 => {
            println!("Client Error");
            "Unauthorized".to_string()
        }
        403 => {
            println!("Client Error");
            "Forbidden".to_string()
        }
        404 => {
            println!("Client Error");
            "Not Found".to_string()
        }
        500 => {
            println!("Server Error");
            "Internal Server Error".to_string()
        }
        502 => {
            println!("Server Error");
            "Bad Gateway".to_string()
        }
        503 => {
            println!("Server Error");
            "Service Unavailable".to_string()
        }
        _ => {
            println!("Unknown");
            "Unknown Status".to_string()
        }
    }
}

// Violation: Duplicated loop logic
fn sum_positive_numbers(numbers: &[i32]) -> i32 {
    let mut sum = 0;
    for &num in numbers {
        if num > 0 {
            sum += num;
        }
    }
    sum
}

fn sum_negative_numbers(numbers: &[i32]) -> i32 {
    let mut sum = 0;
    for &num in numbers {
        if num < 0 {
            sum += num;
        }
    }
    sum
}

fn sum_even_numbers(numbers: &[i32]) -> i32 {
    let mut sum = 0;
    for &num in numbers {
        if num % 2 == 0 {
            sum += num;
        }
    }
    sum
}

// Violation: Duplicated trait implementations
trait Printable {
    fn print(&self);
}

struct Document {
    content: String,
}

impl Printable for Document {
    fn print(&self) {
        println!("=== Document ===");
        println!("{}", self.content);
        println!("================");
    }
}

struct Report {
    content: String,
}

impl Printable for Report {
    fn print(&self) {
        println!("=== Report ===");
        println!("{}", self.content);
        println!("==============");
    }
}

// Violation: Duplicated constant definitions across modules
mod module_a {
    pub const MAX_SIZE: usize = 1024;
    pub const MIN_SIZE: usize = 10;
    pub const DEFAULT_TIMEOUT: u64 = 30;
}

mod module_b {
    pub const MAX_SIZE: usize = 1024;  // Duplicate
    pub const MIN_SIZE: usize = 10;    // Duplicate
    pub const DEFAULT_TIMEOUT: u64 = 30; // Duplicate
}