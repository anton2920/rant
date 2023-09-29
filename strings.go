package main

func FindChar(haystack string, needle byte) int {
	var i int

	for i = 0; i < len(haystack); i++ {
		if haystack[i] == needle {
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

func StrToPositiveInt(xs string) (int, bool) {
	var ret int

	for _, x := range xs {
		if (x < '0') || (x > '9') {
			return 0, false
		}
		ret = (ret * 10) + int(x-'0')
	}

	return ret, true
}
