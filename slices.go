package main

func SlicePutPositiveInt(buf []byte, x int) int {
	var ndigits int
	var rx int

	if x == 0 {
		buf[0] = '0'
		return 1
	}

	for x > 0 {
		rx = (10 * rx) + (x % 10)
		x /= 10
		ndigits++
	}

	var i int
	for i = 0; ndigits > 0; i++ {
		buf[i] = byte((rx % 10) + '0')
		rx /= 10
		ndigits--
	}
	return i
}

func SlicePutTm(buf []byte, tm Tm) int {
	var n, ndigits int

	if tm.Mday+1 < 10 {
		buf[n] = '0'
		n++
	}
	ndigits = SlicePutPositiveInt(buf[n:], tm.Mday)
	n += ndigits
	buf[n] = '.'
	n++

	if tm.Mon+1 < 10 {
		buf[n] = '0'
		n++
	}
	ndigits = SlicePutPositiveInt(buf[n:], tm.Mon+1)
	n += ndigits
	buf[n] = '.'
	n++

	ndigits = SlicePutPositiveInt(buf[n:], tm.Year+1900)
	n += ndigits
	buf[n] = ' '
	n++

	if tm.Hour < 10 {
		buf[n] = '0'
		n++
	}
	ndigits = SlicePutPositiveInt(buf[n:], tm.Hour)
	n += ndigits
	buf[n] = ':'
	n++

	if tm.Min < 10 {
		buf[n] = '0'
		n++
	}
	ndigits = SlicePutPositiveInt(buf[n:], tm.Min)
	n += ndigits
	buf[n] = ':'
	n++

	if tm.Sec < 10 {
		buf[n] = '0'
		n++
	}
	ndigits = SlicePutPositiveInt(buf[n:], tm.Sec)
	n += ndigits
	buf[n] = ' '
	n++

	buf[n] = 'M'
	buf[n+1] = 'S'
	buf[n+2] = 'K'

	return n + 3
}
