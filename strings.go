package main

//go:noescape
//go:nosplit
func FindChar(haystack string, needle byte) int

//go:noescape
//go:nosplit
func FindSubstring(haystack, needle string) int

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
