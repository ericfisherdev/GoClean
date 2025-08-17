fn main() {
    let result = Some(42);
    let value = result.unwrap(); // This should be flagged
    let arr = [1, 2, 3];
    let item = arr[10]; // This should panic
    println!("{} {}", value, item);
}
