package main

func FindRune(haystack string, needle rune) int {
	for i, r := range haystack {
		if r == needle {
			return i
		}
	}

	return -1
}

func FindSubstring(haystack, needle string) int
