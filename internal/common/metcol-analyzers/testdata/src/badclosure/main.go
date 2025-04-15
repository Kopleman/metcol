package main

import "os"

func main() {
	func() {
		os.Exit(1) // want "os.Exit call forbidden in main.main"
	}()
}
