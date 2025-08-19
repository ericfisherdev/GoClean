// Test file for Rust unsafe code violations

use std::ptr;

// Violation: Unnecessary unsafe block
fn unnecessary_unsafe() -> i32 {
    unsafe {
        // This operation doesn't require unsafe
        let x = 5;
        let y = 10;
        x + y
    }
}

// Violation: Unsafe without documentation comment
unsafe fn undocumented_unsafe_function(ptr: *const i32) -> i32 {
    *ptr
}

// Violation: Transmute abuse
fn transmute_abuse() {
    unsafe {
        let x: i32 = 42;
        let y: f32 = std::mem::transmute(x); // Dangerous transmute
        println!("Transmuted: {}", y);
    }
}

// Violation: Raw pointer manipulation without bounds checking
fn raw_pointer_arithmetic(data: &[i32]) {
    unsafe {
        let ptr = data.as_ptr();
        let offset_ptr = ptr.add(10); // No bounds check
        let value = *offset_ptr; // Could be out of bounds
        println!("Value: {}", value);
    }
}

// Violation: Multiple unsafe operations in single block
fn multiple_unsafe_operations() {
    unsafe {
        // Too many unsafe operations in one block
        let mut x = 5;
        let raw_ptr = &mut x as *mut i32;
        *raw_ptr = 10;
        
        let y: u32 = std::mem::transmute(raw_ptr as usize);
        
        let uninit: i32 = std::mem::uninitialized(); // Deprecated and dangerous
        
        println!("Results: {} {}", x, y);
    }
}

// Violation: Global mutable static without proper synchronization
static mut GLOBAL_COUNTER: i32 = 0;

fn unsafe_global_access() {
    unsafe {
        GLOBAL_COUNTER += 1; // Race condition potential
        println!("Counter: {}", GLOBAL_COUNTER);
    }
}

// Violation: Unsafe trait implementation without safety documentation
unsafe trait UnsafeTrait {
    fn dangerous_operation(&self);
}

// Violation: Implementing unsafe trait without documenting safety
unsafe impl UnsafeTrait for i32 {
    fn dangerous_operation(&self) {
        println!("Dangerous: {}", self);
    }
}

// Violation: Creating uninitialized memory unsafely
fn uninitialized_memory() {
    unsafe {
        let mut uninit: [i32; 100] = std::mem::MaybeUninit::uninit().assume_init();
        uninit[0] = 42; // Using uninitialized memory
    }
}

// Violation: Unsafe FFI without null checks
extern "C" {
    fn external_function(ptr: *const i32) -> i32;
}

fn unsafe_ffi_call(value: *const i32) -> i32 {
    unsafe {
        external_function(value) // No null check
    }
}

// Violation: Mutable aliasing through unsafe
fn mutable_aliasing() {
    let mut x = 5;
    let ptr1 = &mut x as *mut i32;
    let ptr2 = &mut x as *mut i32;
    
    unsafe {
        *ptr1 = 10;
        *ptr2 = 20; // Mutable aliasing violation
    }
}

// Violation: Unsafe implementation of Send/Sync without justification
struct NotThreadSafe {
    data: *const i32,
}

unsafe impl Send for NotThreadSafe {} // Should document why this is safe
unsafe impl Sync for NotThreadSafe {} // Should document why this is safe

// Violation: Inline assembly without safety documentation
#[cfg(target_arch = "x86_64")]
fn inline_assembly() {
    unsafe {
        std::arch::asm!("nop"); // Should document what this does and why it's safe
    }
}

// Violation: Dereferencing raw pointers from safe code
fn creates_dangling_pointer() -> *const i32 {
    let x = 42;
    &x as *const i32 // Returns pointer to local variable
}

// Violation: Union access without safety checks
union IntOrFloat {
    i: i32,
    f: f32,
}

fn unsafe_union_access() {
    let mut u = IntOrFloat { i: 42 };
    unsafe {
        u.f = 3.14;
        println!("Int value: {}", u.i); // Reading wrong variant
    }
}

// Violation: Casting between incompatible function pointers
fn function_pointer_cast() {
    unsafe {
        let f: fn(i32) -> i32 = |x| x * 2;
        let g: fn(f32) -> f32 = std::mem::transmute(f); // Dangerous cast
        let result = g(3.14);
        println!("Result: {}", result);
    }
}