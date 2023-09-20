package main

func Fatal(msg string) {
	Print(msg)
	Exit(1)
}

func Print(msg string) {
	Write(2, msg)
	Write(2, "\n")
}
