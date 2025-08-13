package samples

// ComplexFunction has high cyclomatic complexity for testing
func ComplexFunction(x, y, z int) int {
	result := 0
	
	if x > 0 {
		if y > 0 {
			if z > 0 {
				for i := 0; i < x; i++ {
					if i%2 == 0 {
						result += i
					} else {
						result -= i
					}
				}
			} else if z < 0 {
				result = x + y
			} else {
				result = x - y
			}
		} else if y < 0 {
			result = x * 2
		} else {
			result = x / 2
		}
	} else if x < 0 {
		result = -1
	} else {
		result = 0
	}
	
	switch result {
	case 0:
		return 1
	case 1:
		return 2
	default:
		return result
	}
}