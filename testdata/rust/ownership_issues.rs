// Test file for Rust ownership and borrowing violations

use std::collections::HashMap;

// Violation: Unnecessary clone
fn unnecessary_clone_example(s: String) -> String {
    let result = s.clone(); // Unnecessary - could just return s
    result
}

// Violation: Multiple unnecessary clones
fn excessive_cloning(data: Vec<i32>) -> (Vec<i32>, Vec<i32>) {
    let copy1 = data.clone();
    let copy2 = data.clone(); // Could use references instead
    (copy1, copy2)
}

// Violation: Inefficient borrowing - taking ownership when reference would suffice
fn takes_ownership_unnecessarily(s: String) {
    println!("Length: {}", s.len()); // Only needs to read, should take &str or &String
}

// Violation: Complex lifetime parameters
fn complex_lifetime<'a, 'b, 'c, 'd>(
    x: &'a str,
    y: &'b str,
    z: &'c str,
    w: &'d str,
) -> &'a str
where
    'b: 'a,
    'c: 'b,
    'd: 'c,
{
    x
}

// Violation: Move semantics violation - attempting to use moved value
fn move_violation() {
    let data = vec![1, 2, 3];
    let moved_data = data;
    // Uncommenting would cause compile error, but pattern shows poor ownership design
    // println!("{:?}", data); // data has been moved
}

// Violation: Unnecessary Arc/Rc usage for single-threaded context
use std::rc::Rc;
fn unnecessary_rc() {
    let data = Rc::new(vec![1, 2, 3]); // Overkill for simple single-threaded usage
    println!("{:?}", data);
}

// Violation: Cloning in a loop
fn clone_in_loop(items: Vec<String>) {
    let mut results = Vec::new();
    for item in items {
        results.push(item.clone()); // Cloning every iteration
        results.push(item.clone()); // Double clone!
    }
}

// Violation: Inefficient string concatenation with ownership issues
fn inefficient_string_building(parts: Vec<String>) -> String {
    let mut result = String::new();
    for part in parts {
        result = result + &part; // Creates new String each time, inefficient
    }
    result
}

// Violation: Passing mutable reference when immutable would suffice
fn takes_mut_unnecessarily(data: &mut Vec<i32>) -> usize {
    data.len() // Only reading, doesn't need &mut
}

// Violation: Creating temporary values that immediately go out of scope
fn temporary_allocation_waste() {
    for i in 0..1000 {
        let temp = vec![i; 100]; // Allocated and immediately dropped
        println!("{}", temp[0]);
    }
}

// Violation: Unnecessary Box allocation
fn unnecessary_box() -> Box<i32> {
    Box::new(42) // Simple value doesn't need heap allocation
}

// Violation: Multiple mutable borrows pattern (would fail compilation but shows bad design)
struct Container {
    data: Vec<i32>,
}

impl Container {
    fn bad_borrow_pattern(&mut self) {
        let first = &mut self.data;
        // let second = &mut self.data; // Would fail - multiple mutable borrows
        first.push(1);
    }
}

// Violation: Returning reference to local data (lifetime issue)
// fn returns_local_reference() -> &'static str {
//     let local = String::from("local");
//     &local // Would fail - returning reference to local
// }

// Violation: Unnecessary lifetime annotations
fn simple_function_with_lifetimes<'a>(x: &'a str) -> &'a str {
    x // Lifetime could be elided
}

// Violation: Clone to satisfy borrow checker instead of restructuring
fn clone_to_avoid_borrow_checker(map: &mut HashMap<String, Vec<i32>>, key: &str) {
    if let Some(vec) = map.get(key) {
        let cloned = vec.clone(); // Cloning to avoid borrow checker
        map.insert(key.to_string(), cloned);
    }
}