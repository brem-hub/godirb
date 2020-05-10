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
	depth       = flag.Int("s", 10, "choose size of default dict [10, 100, 1000, 10000]")
)

func main() {
	// resp := Response{url: "124.com", code: 404}
	// resp2 := Response{url: "123.com", code: 200}
	// resp3 := Response{url: "122.com", code: 303}
	// cwR := make(chan Response, 3)
	// cw := CommonWriter{responses: cwR, w: os.Stdout}
	// cwR <- resp
	// cwR <- resp2
	// cwR <- resp3
	// close(cwR)
	// fmt.Fprint(&cw)

	flag.Parse()
	if *custom_dict {
		if *dict_path == "" {
			fmt.Println("specify custom dictionary path")
			os.Exit(1)
		}
		bruteWebSite(*url, *dict_path, *visual)
	} else {
		switch *depth {
		case 10:
			bruteWebSite(*url, "data/dicc.txt", *visual)
		case 100:
			bruteWebSite(*url, "data/brute100.txt", *visual)
		case 1000:
			bruteWebSite(*url, "data/brute1000.txt", *visual)
		case 10000:
			bruteWebSite(*url, "data/brute10000.txt", *visual)

		}
	}
}
