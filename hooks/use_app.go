package hooks

import "os"

// UseApp returns the application instance (as interface{})
func UseApp(hc *HooksContext) any {
	return hc.App
}

// UseStdin returns the standard input file handle
func UseStdin(hc *HooksContext) *os.File {
	return hc.Stdin
}

// UseStdout returns the standard output file handle
func UseStdout(hc *HooksContext) *os.File {
	return hc.Stdout
}

// UseStderr returns the standard error file handle
func UseStderr(hc *HooksContext) *os.File {
	return hc.Stderr
}
