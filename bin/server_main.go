package main

import "flag"

var ip string

func main() {
	flag.StringVar(&ip, "h", "localhost", "connect to")
}
