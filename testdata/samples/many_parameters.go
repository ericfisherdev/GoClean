package samples

// FunctionWithManyParameters has too many parameters
func FunctionWithManyParameters(a, b, c, d, e, f, g, h int) int {
	return a + b + c + d + e + f + g + h
}

// AnotherComplexFunction also has too many parameters
func AnotherComplexFunction(
	firstName string,
	lastName string,
	age int,
	email string,
	phone string,
	address string,
	city string,
	state string,
	zipCode string,
) {
	// Implementation
}

// BetterDesign shows a better approach using a struct
type UserInfo struct {
	FirstName string
	LastName  string
	Age       int
	Email     string
	Phone     string
	Address   string
	City      string
	State     string
	ZipCode   string
}

func BetterFunction(user UserInfo) {
	// Better design with fewer parameters
}