// Test Rust file for parser validation
use std::collections::HashMap;

pub struct TestStruct {
    name: String,
    value: i32,
}

pub trait TestTrait {
    fn do_something(&self) -> bool;
}

impl TestTrait for TestStruct {
    fn do_something(&self) -> bool {
        true
    }
}

pub fn test_function(param1: &str, param2: i32) -> String {
    if param2 > 0 {
        format!("{}: {}", param1, param2)
    } else {
        String::new()
    }
}

pub const MAX_SIZE: usize = 100;

pub mod test_module {
    pub fn internal_function() {
        println!("Hello from module");
    }
}

macro_rules! test_macro {
    ($name:expr) => {
        println!("Hello, {}!", $name);
    };
}

pub enum Color {
    Red,
    Green,
    Blue,
}