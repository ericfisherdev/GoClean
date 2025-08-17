// Test file for Rust pattern matching violations

#[derive(Debug)]
enum Status {
    Success,
    Warning,
    Error,
    Critical,
}

// Violation: Non-exhaustive pattern matching (would fail compilation, but shows bad practice)
fn non_exhaustive_match(status: Status) {
    match status {
        Status::Success => println!("OK"),
        Status::Error => println!("Error"),
        _ => println!("Other"), // Using catch-all instead of explicit patterns
    }
}

// Violation: Deeply nested pattern matching
fn nested_pattern_matching(opt1: Option<Option<Option<i32>>>) -> i32 {
    match opt1 {
        Some(opt2) => {
            match opt2 {
                Some(opt3) => {
                    match opt3 {
                        Some(value) => value,
                        None => 0,
                    }
                }
                None => 0,
            }
        }
        None => 0,
    }
}

// Violation: Inefficient pattern destructuring
fn inefficient_destructuring(tuple: (i32, i32, i32, i32, i32)) -> i32 {
    match tuple {
        (a, _, _, _, _) => a, // Could use tuple.0 directly
    }
}

// Violation: Redundant pattern matching
fn redundant_match(opt: Option<i32>) -> i32 {
    match opt {
        Some(x) => x,
        None => 0,
    } // Could use opt.unwrap_or(0)
}

// Violation: Complex pattern with too many conditions
fn complex_pattern(x: i32, y: i32, z: i32) -> &'static str {
    match (x, y, z) {
        (0, 0, 0) => "all zero",
        (1, 0, 0) | (0, 1, 0) | (0, 0, 1) => "one is one",
        (1, 1, 0) | (1, 0, 1) | (0, 1, 1) => "two are one",
        (1, 1, 1) => "all one",
        (2..=10, 2..=10, 2..=10) => "all in range",
        _ => "other",
    }
}

// Violation: Using if-let when match would be clearer
fn multiple_if_lets(opt1: Option<i32>, opt2: Option<i32>) -> i32 {
    if let Some(x) = opt1 {
        if let Some(y) = opt2 {
            x + y
        } else {
            x
        }
    } else if let Some(y) = opt2 {
        y
    } else {
        0
    }
}

// Violation: Pattern matching on boolean (unnecessary)
fn match_on_bool(flag: bool) -> &'static str {
    match flag {
        true => "yes",
        false => "no",
    } // Should use if-else
}

// Violation: Overly complex guard conditions
fn complex_guards(x: i32) -> &'static str {
    match x {
        n if n > 0 && n < 10 && n % 2 == 0 => "even single digit",
        n if n >= 10 && n < 100 && n % 2 == 0 => "even double digit",
        n if n > 0 && n < 10 && n % 2 != 0 => "odd single digit",
        n if n >= 10 && n < 100 && n % 2 != 0 => "odd double digit",
        _ => "other",
    }
}

// Violation: Not using pattern matching where appropriate
fn without_pattern_matching(result: Result<i32, String>) -> i32 {
    if result.is_ok() {
        result.unwrap()
    } else {
        -1
    } // Should use match or unwrap_or
}

// Violation: Unnecessary Box pattern matching
fn unnecessary_box_pattern(boxed: Box<i32>) -> i32 {
    match *boxed {
        x => x, // Overly complex for simple dereference
    }
}

// Violation: Pattern matching with side effects
fn pattern_with_side_effects(opt: Option<i32>) -> i32 {
    let mut counter = 0;
    match opt {
        Some(x) => {
            counter += 1; // Side effect in pattern
            println!("Got value: {}", x);
            x + counter
        }
        None => {
            counter += 2; // Side effect in pattern
            println!("Got none");
            counter
        }
    }
}

// Violation: Overlapping patterns (would cause warning)
fn overlapping_patterns(x: i32) -> &'static str {
    match x {
        0..=10 => "low",
        5..=15 => "mid", // Overlaps with previous
        10..=20 => "high", // Overlaps with previous
        _ => "other",
    }
}

// Violation: Using match for simple value extraction
fn simple_extraction(tuple: (i32, String)) -> i32 {
    match tuple {
        (x, _) => x, // Could use tuple.0
    }
}

// Violation: Inconsistent pattern style
fn inconsistent_patterns(opt: Option<(i32, i32)>) -> i32 {
    match opt {
        Some((x, y)) => x + y, // Destructuring
        None => {
            0 // Inconsistent block usage
        }
    }
}