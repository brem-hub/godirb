package main

import (
	"flag"
	"fmt"
	"os"
)

//TODO: add output file
var (
	url         = flag.String("u", "http://127.0.0.1:8000/", "url to bruteforce")
	custom_dict = flag.Bool("cd", false, "use custom dictionary")
	dict_path   = flag.String("d", "", "custom dictionary path")
	visual      = flag.Bool("v", false, "more output")
	//change
	depth      = flag.Int("s", 10, "choose size of default dict [10, 100, 1000, 10000]")
	power      = flag.Int("p", 1, "Amount of Goroutines X10. Normal usage: [1 ... 5]")
	extensions StringSlice
)

func main() {
	flag.Var(&extensions, "e", "extensions to pass. Usage: -e=php,txt,rcc")
	flag.Parse()
	if *custom_dict {
		if *dict_path == "" {
			fmt.Println("specify custom dictionary path")
			os.Exit(1)
		}
		bruteWebSite(*url, *dict_path, extensions, *power, *visual)
	} else {
		switch *depth {
		case 10:
			bruteWebSite(*url, "data/brute10.txt", extensions, *power, *visual)
		case 100:
			bruteWebSite(*url, "data/dicc.txt", extensions, *power, *visual)
		}
	}
}
