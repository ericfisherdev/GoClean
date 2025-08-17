// Test file for Rust naming convention violations

// Violation: Function should be snake_case, not camelCase
fn getUserName() -> String {
    String::from("John Doe")
}

// Violation: Function should be snake_case, not PascalCase
fn ProcessData(data: &str) -> bool {
    !data.is_empty()
}

// Violation: Struct should be PascalCase, not snake_case
struct user_profile {
    name: String,
    age: u32,
}

// Violation: Struct should be PascalCase, not camelCase
struct userAccount {
    id: u64,
    balance: f64,
}

// Violation: Enum should be PascalCase, not snake_case
enum request_status {
    Pending,
    Approved,
    Rejected,
}

// Violation: Trait should be PascalCase, not snake_case
trait data_processor {
    fn process(&self);
}

// Violation: Constant should be SCREAMING_SNAKE_CASE, not camelCase
const maxRetryCount: u32 = 3;

// Violation: Constant should be SCREAMING_SNAKE_CASE, not snake_case
const default_timeout: u64 = 30;

// Violation: Static should be SCREAMING_SNAKE_CASE, not PascalCase
static GlobalCounter: u64 = 0;

// Violation: Variable should be snake_case, not camelCase
fn main() {
    let userName = "Alice";
    let UserAge = 25;
    let TEMP_VALUE = 42; // Should be snake_case for local variable
    
    // Violation: Function parameter should be snake_case
    fn helper(InputData: &str, RetryCount: u32) {
        println!("{} {}", InputData, RetryCount);
    }
}

// Violation: Module should be snake_case, not PascalCase
mod UserManagement {
    pub fn init() {}
}

// Violation: Method should be snake_case, not camelCase
impl user_profile {
    fn getUserAge(&self) -> u32 {
        self.age
    }
    
    // Violation: Method should be snake_case, not PascalCase
    fn UpdateName(&mut self, name: String) {
        self.name = name;
    }
}

// Violation: Type alias should be PascalCase, not snake_case
type user_id = u64;

// Violation: Associated type should be PascalCase
trait Container {
    type item_type; // Should be ItemType
}

// Violation: Generic type parameter should be PascalCase (conventionally single letter or descriptive)
fn generic_function<t>(value: t) -> t {
    value
}