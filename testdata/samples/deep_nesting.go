// Package samples provides test cases demonstrating various clean code violations.
package samples

// DeeplyNested demonstrates deep nesting levels
func DeeplyNested(data [][]int) int {
	result := 0
	
	for i := 0; i < len(data); i++ {
		for j := 0; j < len(data[i]); j++ {
			if data[i][j] > 0 {
				for k := 0; k < data[i][j]; k++ {
					if k%2 == 0 {
						for l := 0; l < 5; l++ {
							if l == k {
								result += l
							}
						}
					}
				}
			}
		}
	}
	
	return result
}

// AnotherNestedFunction also has deep nesting
func AnotherNestedFunction(x int) {
	if x > 0 {
		if x < 100 {
			if x%2 == 0 {
				if x%5 == 0 {
					// Too deeply nested
				}
			}
		}
	}
}