// Test file for commented code and TODO/FIXME violations

fn active_function() {
    println!("This is active code");
    
    // Violation: Commented out code block
    // let old_variable = 42;
    // if old_variable > 40 {
    //     println!("Old logic");
    // }
    
    let current = 10;
    
    /* Violation: Multi-line commented code
    fn removed_function() {
        println!("This was removed");
        let x = 5;
        return x * 2;
    }
    */
    
    // This is a normal comment explaining the next line (OK)
    let result = current * 2;
    
    // Violation: More commented code
    // let mut vec = Vec::new();
    // vec.push(1);
    // vec.push(2);
    // for item in vec {
    //     println!("{}", item);
    // }
}

// TODO: Implement this function properly
fn incomplete_function() {
    // FIXME: This is a temporary hack
    let hack_value = 999;
    
    // TODO(user): Add proper error handling here
    println!("Value: {}", hack_value);
    
    // HACK: This shouldn't be hardcoded
    let hardcoded = "localhost:8080";
    
    // XXX: This needs urgent attention
    unsafe {
        // Dangerous operation
    }
    
    // NOTE: This is just a note (should be OK)
    let documented = true;
    
    // OPTIMIZE: This could be faster
    for i in 0..1000000 {
        // Slow operation
    }
    
    // REVIEW: Please review this logic
    if hack_value > 1000 {
        panic!("Too large");
    }
}

struct DataProcessor;

impl DataProcessor {
    fn process(&self) {
        // Violation: Commented method
        // fn old_process(&self) {
        //     println!("Old processing logic");
        //     self.validate();
        //     self.transform();
        // }
        
        println!("New processing");
    }
    
    /* Violation: Entire commented impl block
    fn validate(&self) -> bool {
        // Validation logic
        true
    }
    
    fn transform(&self) -> String {
        // Transformation logic
        String::from("transformed")
    }
    */
}

// Violation: Commented struct definition
// struct OldStruct {
//     field1: i32,
//     field2: String,
// }

// Violation: Commented use statements
// use std::collections::HashSet;
// use std::sync::Arc;

fn function_with_mixed_comments() {
    // This explains the algorithm (OK)
    // We use a two-pointer approach to find the target
    let mut left = 0;
    let mut right = 100;
    
    // Violation: Old algorithm commented out
    // while left < right {
    //     let mid = (left + right) / 2;
    //     if mid > target {
    //         right = mid;
    //     } else {
    //         left = mid + 1;
    //     }
    // }
    
    // Current implementation
    let result = (left + right) / 2;
    
    /* This is documentation about the result (OK)
     * The result represents the midpoint
     * between left and right boundaries
     */
    println!("Result: {}", result);
}

// Violation: Commented test code
// #[test]
// fn old_test() {
//     assert_eq!(2 + 2, 4);
//     let processor = DataProcessor;
//     processor.process();
// }

mod feature {
    // TODO: Complete feature implementation
    pub fn feature_stub() {
        unimplemented!("Feature not ready"); // FIXME: Remove before release
    }
    
    // Violation: Commented module code
    // pub fn removed_feature() {
    //     println!("This feature was removed");
    //     let config = load_config();
    //     apply_config(config);
    // }
    
    // Violation: Block of commented code with syntax errors (definitely code)
    /*
    fn broken_function() {
        let x = 10
        let y = 20  // Missing semicolons
        return x + y
    }
    */
}

// Violation: Commented macro definition
// macro_rules! old_macro {
//     ($x:expr) => {
//         println!("Old macro: {}", $x);
//     };
// }

// DEPRECATED: This function will be removed in v2.0
fn deprecated_function() {
    // TODO: Migrate callers to new_function()
    println!("Deprecated");
}

// Violation: Large block of commented code at end of file
/*
impl OldImplementation {
    fn method1(&self) -> i32 {
        42
    }
    
    fn method2(&self, input: String) -> String {
        format!("Processed: {}", input)
    }
    
    fn method3(&mut self) {
        self.state = State::Active;
        self.counter += 1;
    }
}

fn old_helper_function(x: i32, y: i32) -> i32 {
    if x > y {
        x - y
    } else {
        y - x
    }
}
*/