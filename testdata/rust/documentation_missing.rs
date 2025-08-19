// Test file for missing documentation violations

// Violation: Missing documentation for public struct
pub struct UndocumentedStruct {
    pub field1: i32,
    pub field2: String,
}

// Violation: Missing documentation for public enum
pub enum UndocumentedEnum {
    Variant1,
    Variant2(i32),
    Variant3 { x: i32, y: i32 },
}

// Violation: Missing documentation for public trait
pub trait UndocumentedTrait {
    fn method(&self) -> i32;
}

// Violation: Missing documentation for public function
pub fn undocumented_function(x: i32) -> i32 {
    x * 2
}

// Violation: Missing documentation for public type alias
pub type UndocumentedType = Vec<String>;

// Violation: Missing documentation for public constant
pub const UNDOCUMENTED_CONST: i32 = 42;

// Violation: Missing documentation for public static
pub static UNDOCUMENTED_STATIC: &str = "value";

// Violation: Missing documentation for public module
pub mod undocumented_module {
    // Violation: Missing documentation for nested public function
    pub fn nested_undocumented() -> i32 {
        42
    }
}

impl UndocumentedStruct {
    // Violation: Missing documentation for public method
    pub fn public_method(&self) -> i32 {
        self.field1
    }
    
    // Private method - no violation expected
    fn private_method(&self) -> &str {
        &self.field2
    }
}

// Violation: Missing documentation for public struct with public fields
pub struct StructWithPublicFields {
    // Violation: Missing documentation for public field
    pub important_field: i32,
    
    // Private field - no violation expected
    private_field: String,
}

// Violation: Missing documentation for public impl block
impl UndocumentedTrait for UndocumentedStruct {
    fn method(&self) -> i32 {
        self.field1
    }
}

// Violation: Missing documentation for public macro
#[macro_export]
macro_rules! undocumented_macro {
    () => {
        println!("This macro lacks documentation");
    };
}

// Violation: Missing documentation for pub use
pub use std::collections::HashMap as UndocumentedReexport;

// Violation: Generic public function without docs
pub fn generic_undocumented<T>(value: T) -> T {
    value
}

// Violation: Public async function without docs
pub async fn async_undocumented() -> Result<(), Box<dyn std::error::Error>> {
    Ok(())
}

// Violation: Public unsafe function without safety documentation
pub unsafe fn unsafe_undocumented(ptr: *const i32) -> i32 {
    *ptr
}

// Well-documented items for comparison
/// This struct is properly documented
pub struct DocumentedStruct {
    /// This field has documentation
    pub documented_field: i32,
}

/// This function is properly documented
/// 
/// # Arguments
/// * `x` - The input value
/// 
/// # Returns
/// The doubled value
pub fn documented_function(x: i32) -> i32 {
    x * 2
}