// Test file for magic number violations

fn calculate_price(quantity: i32) -> f64 {
    // Violation: Magic number 19.99
    let unit_price = 19.99;
    
    // Violation: Magic number 0.15 (should be TAX_RATE)
    let tax = quantity as f64 * unit_price * 0.15;
    
    // Violation: Magic number 5.0 (should be SHIPPING_FEE)
    quantity as f64 * unit_price + tax + 5.0
}

fn check_temperature(temp: f32) -> &'static str {
    // Violation: Magic numbers 32 and 212 (freezing and boiling points)
    if temp <= 32.0 {
        "freezing"
    } else if temp >= 212.0 {
        "boiling"
    } else {
        "normal"
    }
}

fn calculate_circle_area(radius: f64) -> f64 {
    // Violation: Magic number 3.14159 (should be PI constant)
    3.14159 * radius * radius
}

fn get_status_code(code: i32) -> &'static str {
    match code {
        200 => "OK",           // Violation: Magic number 200
        404 => "Not Found",    // Violation: Magic number 404
        500 => "Server Error", // Violation: Magic number 500
        _ => "Unknown",
    }
}

fn calculate_days(years: i32) -> i32 {
    // Violation: Magic number 365 (should be DAYS_PER_YEAR)
    years * 365
}

fn convert_to_seconds(hours: i32) -> i32 {
    // Violations: Magic numbers 60 and 60 (should be MINUTES_PER_HOUR and SECONDS_PER_MINUTE)
    hours * 60 * 60
}

fn calculate_discount(price: f64, customer_type: i32) -> f64 {
    match customer_type {
        1 => price * 0.9,  // Violation: Magic number 0.9 (10% discount)
        2 => price * 0.8,  // Violation: Magic number 0.8 (20% discount)
        3 => price * 0.7,  // Violation: Magic number 0.7 (30% discount)
        _ => price,
    }
}

fn check_array_bounds(index: usize) -> bool {
    // Violation: Magic number 100 (should be MAX_ARRAY_SIZE)
    index < 100
}

fn calculate_fibonacci(n: i32) -> i32 {
    if n <= 1 {
        n
    } else {
        // The 1 here is arguably OK (mathematical definition)
        calculate_fibonacci(n - 1) + calculate_fibonacci(n - 2)
    }
}

struct Rectangle {
    width: f64,
    height: f64,
}

impl Rectangle {
    fn is_square(&self) -> bool {
        // This comparison is OK (comparing for equality)
        (self.width - self.height).abs() < 0.0001 // Violation: Magic number 0.0001 (epsilon)
    }
    
    fn scale(&mut self, factor: f64) {
        self.width *= factor;
        self.height *= factor;
    }
}

fn process_buffer(buffer: &[u8]) -> Vec<u8> {
    let mut result = Vec::new();
    
    for &byte in buffer {
        // Violation: Magic number 128 (should be BUFFER_THRESHOLD)
        if byte > 128 {
            // Violation: Magic number 255 (should be MAX_BYTE_VALUE)
            result.push(255 - byte);
        } else {
            result.push(byte);
        }
    }
    
    // Violation: Magic number 1024 (should be MAX_BUFFER_SIZE)
    if result.len() > 1024 {
        result.truncate(1024);
    }
    
    result
}

fn calculate_score(hits: i32, misses: i32) -> f64 {
    // Violations: Magic numbers 10 and 5 (point values)
    (hits * 10 - misses * 5) as f64
}

// Constants properly defined (good examples for contrast)
const MAX_RETRIES: i32 = 3;
const DEFAULT_TIMEOUT: u64 = 30;
const PI: f64 = 3.14159265359;

fn good_example_with_constants() -> i32 {
    // These are OK - using named constants
    MAX_RETRIES * DEFAULT_TIMEOUT as i32
}

fn array_indexing_examples() {
    let arr = [1, 2, 3, 4, 5];
    
    // These are generally OK (array indices)
    let first = arr[0];
    let second = arr[1];
    
    // But this might be flagged (depends on context)
    let specific = arr[4]; // Could be arr[arr.len() - 1] for last element
}

fn loop_examples() {
    // Loop bounds that are magic numbers
    for i in 0..50 {  // Violation: Magic number 50
        println!("{}", i);
    }
    
    // This is OK (iterating from 0)
    for i in 0..MAX_RETRIES {
        println!("{}", i);
    }
}