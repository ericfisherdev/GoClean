package samples

// BadNaming demonstrates poor naming conventions
var x = 10           // Single letter variable
var tmpVar = 20      // Abbreviation
var isNotReady = true // Double negative

// a is a poorly named function
func a(b, c int) int {
	d := b + c // Single letter variable
	return d
}

// ProcessDataAndStuff has a vague name
func ProcessDataAndStuff(data []byte) {
	// Function does nothing specific
}

// MAGIC_NUMBER demonstrates magic numbers
const MAGIC_NUMBER = 42

func CalculateSomething() int {
	return MAGIC_NUMBER * 100 // Another magic number
}