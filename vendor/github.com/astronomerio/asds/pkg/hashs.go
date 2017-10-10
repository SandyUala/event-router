package pkg

// Hash retuns the hashcode of the passed in string.  Uses the same algorithem as Java
func Hash(s string) int64 {
	h := int64(0)
	for i := 0; i < len(s); i++ {
		h = 31*h + int64(s[i])
	}
	return h
}
