package slice

func ContainsString(slice []string, s string) bool {
	for _, elem := range slice {
		if elem == s {
			return true
		}
	}
	return false
}

func RemoveString(slice []string, s string) []string {
	newSlice := []string{}
	for _, elem := range slice {
		if elem != s {
			newSlice = append(newSlice, elem)
		}
	}
	return newSlice
}
