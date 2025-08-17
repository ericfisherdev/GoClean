// Test file for Rust error handling violations

use std::fs::File;
use std::io::Read;

// Violation: Overuse of unwrap()
fn unwrap_overuse() {
    let file = File::open("config.txt").unwrap(); // Could panic
    let mut contents = String::new();
    file.read_to_string(&mut contents).unwrap(); // Another unwrap
    let number: i32 = contents.trim().parse().unwrap(); // And another
    println!("Number: {}", number);
}

// Violation: Using expect() excessively
fn expect_overuse() {
    let data = "not a number";
    let num: i32 = data.parse().expect("Failed to parse"); // Will panic
    let vec = vec![1, 2, 3];
    let item = vec.get(10).expect("Index out of bounds"); // Will panic
}

// Violation: Missing error propagation - swallowing errors
fn swallows_errors() -> Option<String> {
    let file = File::open("data.txt");
    if file.is_ok() {
        Some(String::from("Success"))
    } else {
        None // Lost error information
    }
}

// Violation: Inconsistent error handling - mix of Result and Option
fn inconsistent_error_types(flag: bool) -> Option<i32> {
    if flag {
        Some(42)
    } else {
        // Should return Result<i32, Error> for consistency
        None
    }
}

// Violation: Panic-prone code patterns
fn panic_prone() {
    let vec: Vec<i32> = vec![];
    let first = vec[0]; // Will panic on empty vector
    
    let slice = &[1, 2, 3];
    let item = slice[10]; // Will panic - index out of bounds
    
    assert!(false, "This will always panic"); // Intentional panic
}

// Violation: Not using ? operator where appropriate
fn manual_error_propagation() -> Result<String, std::io::Error> {
    let file = File::open("test.txt");
    let mut file = match file {
        Ok(f) => f,
        Err(e) => return Err(e), // Should use ?
    };
    
    let mut contents = String::new();
    match file.read_to_string(&mut contents) {
        Ok(_) => Ok(contents),
        Err(e) => Err(e), // Should use ?
    }
}

// Violation: Ignoring Result values
fn ignores_results() {
    let _ = File::create("output.txt"); // Ignoring potential error
    std::fs::remove_file("temp.txt"); // Not checking if removal succeeded
}

// Violation: Using panic! in library code
pub fn library_function(value: i32) -> i32 {
    if value < 0 {
        panic!("Negative values not allowed"); // Should return Result instead
    }
    value * 2
}

// Violation: todo! and unimplemented! in production code
fn incomplete_function() -> i32 {
    todo!("Need to implement this function") // Should not be in production
}

fn another_incomplete() -> String {
    unimplemented!("This feature is not ready") // Should not be in production
}

// Violation: Nested Result/Option unwrapping
fn nested_unwrapping() {
    let data: Option<Result<i32, String>> = Some(Ok(42));
    let value = data.unwrap().unwrap(); // Double unwrap - dangerous
}

// Violation: Using unreachable! inappropriately
fn uses_unreachable(x: i32) {
    match x {
        1 => println!("One"),
        2 => println!("Two"),
        _ => unreachable!("This could actually be reached"), // Dangerous assumption
    }
}

// Violation: Error type inconsistency
fn mixed_error_types(flag: bool) -> Result<i32, String> {
    if flag {
        Err(String::from("String error")) // Using String as error type
    } else {
        Err(format!("Formatted error")) // Inconsistent error creation
    }
}

// Violation: Not handling all error cases
fn partial_error_handling() -> Result<String, std::io::Error> {
    let mut file = File::open("data.txt")?;
    let mut contents = String::new();
    file.read_to_string(&mut contents)?;
    
    // Not handling potential UTF-8 errors in string processing
    Ok(contents.to_uppercase()) // Could have non-ASCII issues
}

// Violation: Using assert! for non-test validation
fn uses_assert_for_validation(value: i32) {
    assert!(value > 0); // Should use proper error handling
    println!("Value is positive: {}", value);
}