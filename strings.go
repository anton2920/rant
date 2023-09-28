package main

func FindRune(haystack string, needle rune) int {
	for i, r := range haystack {
		if r == needle {
			return i
		}
	}

	return -1
}

func FindSubstring(haystack, needle string) int {
	for i := 0; i < len(haystack)-len(needle); i++ {
		toSearch := haystack[i : i+len(needle)]
		if (toSearch[0] == needle[0]) && (toSearch[len(needle)-1] == needle[len(needle)-1]) {
			if toSearch == needle {
				return i
			}
		}
	}
	return -1
}
