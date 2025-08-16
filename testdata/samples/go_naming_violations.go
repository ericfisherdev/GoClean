// Package samples provides test cases demonstrating various clean code violations.
package samples

import "errors"

// Variables with naming violations
var bad_variable_name = "test"      // Underscore violation
var BadVariableName = "exported"    // Should be camelCase if unexported
var goodVariable = "correct"        // Correct camelCase

// Constants with naming violations
const bad_constant = "test"         // Should use camelCase for unexported
const BadConstant = "wrong"         // Should use SCREAMING_SNAKE_CASE for exported
const GOOD_CONSTANT = "correct"     // Correct SCREAMING_SNAKE_CASE
const goodConstant = "correct"      // Correct camelCase for unexported

// Boolean variables without proper prefixes
var ready = true                    // Should have is/has/can prefix
var isReady = true                  // Correct boolean naming
var hasPermission = false           // Correct boolean naming

// Error variables without proper naming
var customError = errors.New("test")    // Should end with Err or be 'err'
var parseErr = errors.New("parse")      // Correct error naming
var err = errors.New("generic")         // Correct error naming

// Functions with naming violations
func get_data() string {            // Underscore violation
	return "data"
}

func GetData() string {             // Inappropriate "get" prefix
	return "data"
}

func data() string {                // Correct - no "get" prefix
	return "data"
}

func bad_function_name() {          // Underscore violation
	// Implementation
}

func GoodFunctionName() {           // Correct exported function
	// Implementation
}

func goodFunctionName() {           // Correct unexported function
	// Implementation
}

// Types with naming violations
type bad_struct struct {            // Underscore violation
	Field string
}

type BadStruct struct {             // This should be unexported based on usage
	field string
}

type GoodStruct struct {            // Correct exported struct
	Field string
}

type goodStruct struct {            // Correct unexported struct
	field string
}

// Interface with naming violations
type DataProcessor interface {      // Should conventionally have -er suffix
	Process(data string) error
}

type Processor interface {          // Correct interface naming
	Process(data string) error
}

// Test various edge cases
func HTTP() {                       // Acronym - should be acceptable
	// Implementation
}

func URL() {                        // Acronym - should be acceptable
	// Implementation
}