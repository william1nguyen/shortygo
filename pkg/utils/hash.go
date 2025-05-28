package utils

func HashString(s string) int {
	hash := 0
	for _, char := range s {
		hash += int(char)
	}
	return hash
}

func ConsistentHashing(key string, nodeCount int) int {
	if nodeCount <= 0 {
		return 0
	}
	hash := HashString(key)
	return hash % nodeCount
}
