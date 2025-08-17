// Test file for Rust function complexity and length violations

// Violation: Function too long (>25 lines)
fn very_long_function(data: Vec<i32>) -> i32 {
    println!("Starting processing");
    let mut result = 0;
    
    // Line 1
    for item in &data {
        result += item;
    }
    
    // Line 5
    if result > 100 {
        println!("Result is large");
    }
    
    // Line 9
    let average = result / data.len() as i32;
    println!("Average: {}", average);
    
    // Line 12
    let mut max = 0;
    for item in &data {
        if *item > max {
            max = *item;
        }
    }
    
    // Line 19
    let mut min = i32::MAX;
    for item in &data {
        if *item < min {
            min = *item;
        }
    }
    
    // Line 26
    println!("Max: {}, Min: {}", max, min);
    
    // Line 28
    let variance = data.iter()
        .map(|x| (*x - average).pow(2))
        .sum::<i32>() / data.len() as i32;
    
    // Line 32
    println!("Variance: {}", variance);
    
    // Line 34
    if variance > 50 {
        println!("High variance detected");
    }
    
    // Line 38
    result
}

// Violation: Too many parameters (>4)
fn too_many_parameters(
    param1: i32,
    param2: String,
    param3: Vec<u8>,
    param4: bool,
    param5: f64,
    param6: Option<i32>,
) -> i32 {
    if param4 {
        param1 + param6.unwrap_or(0)
    } else {
        param1
    }
}

// Violation: High cyclomatic complexity
fn high_complexity(x: i32, y: i32, z: i32, flag: bool) -> i32 {
    let mut result = 0;
    
    if x > 0 {
        if y > 0 {
            if z > 0 {
                result = x + y + z;
            } else if z < 0 {
                result = x + y - z;
            } else {
                result = x + y;
            }
        } else if y < 0 {
            if z > 0 {
                result = x - y + z;
            } else if z < 0 {
                result = x - y - z;
            } else {
                result = x - y;
            }
        } else {
            result = x;
        }
    } else if x < 0 {
        if flag {
            if y > 10 {
                result = -x + y;
            } else if y > 5 {
                result = -x + y / 2;
            } else {
                result = -x;
            }
        } else {
            result = x;
        }
    } else {
        if flag && y > 0 && z > 0 {
            result = y + z;
        } else if !flag && y < 0 && z < 0 {
            result = -(y + z);
        }
    }
    
    result
}

// Violation: Deep nesting (>3 levels)
fn deeply_nested_function(data: Vec<Vec<Vec<i32>>>) -> i32 {
    let mut sum = 0;
    
    for outer in data {
        if !outer.is_empty() {
            for middle in outer {
                if !middle.is_empty() {
                    for inner in middle {
                        if inner > 0 {
                            if inner % 2 == 0 {
                                if inner < 100 {
                                    sum += inner;
                                }
                            }
                        }
                    }
                }
            }
        }
    }
    
    sum
}

// Violation: Multiple violations in one function
fn multiple_violations(
    a: i32,
    b: i32,
    c: i32,
    d: i32,
    e: i32, // Too many params
) -> i32 {
    // Long function with high complexity
    let mut result = 0;
    
    if a > 0 {
        if b > 0 {
            if c > 0 {
                if d > 0 { // Deep nesting
                    result = a + b + c + d;
                }
            }
        }
    }
    
    // More complexity
    match e {
        0 => result += 1,
        1 => result += 2,
        2 => result += 3,
        3 => result += 4,
        4 => result += 5,
        5 => result += 6,
        6 => result += 7,
        7 => result += 8,
        8 => result += 9,
        9 => result += 10,
        _ => result += 0,
    }
    
    // Even more code to make it longer
    for i in 0..10 {
        if i % 2 == 0 {
            result += i;
        } else {
            result -= i;
        }
    }
    
    result
}

// Violation: Function with too many local variables (cognitive complexity)
fn too_many_locals() -> i32 {
    let var1 = 1;
    let var2 = 2;
    let var3 = 3;
    let var4 = 4;
    let var5 = 5;
    let var6 = 6;
    let var7 = 7;
    let var8 = 8;
    let var9 = 9;
    let var10 = 10;
    let var11 = 11;
    let var12 = 12;
    
    var1 + var2 + var3 + var4 + var5 + var6 + 
    var7 + var8 + var9 + var10 + var11 + var12
}