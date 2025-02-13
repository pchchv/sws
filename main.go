package main

import "github.com/pchchv/sws/helpers"

func main() {
	helpers.Newline = true
	helpers.SlogIt = true
	helpers.SetupSlog()
}
