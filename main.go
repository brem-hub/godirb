package main

import (
	"flag"
	"fmt"
	"os"
)

/*TODO: create output file
change colors#
change dicts#
create recursion
create bot
create POST
create log
create soft exit#
*/
var (
	url         = flag.String("u", "http://127.0.0.1:8000/", "url to bruteforce")
	custom_dict = flag.Bool("cd", false, "use custom dictionary")
	dict_path   = flag.String("d", "", "custom dictionary path")
	verbose     = flag.Bool("v", false, "more output")
	method      = flag.String("m", "get", "specify method to use [get, post]")
	power       = flag.Int("p", 1, "Amount of Goroutines X10. Normal usage: [1 ... 5]")
	extensions  StringSlice
)

func main() {
	flag.Var(&extensions, "e", "extensions to pass. Usage: -e=php,txt,rcc")
	flag.Parse()
	if *custom_dict {
		if *dict_path == "" {
			fmt.Println("specify custom dictionary path")
			os.Exit(1)
		}
		bruteWebSite(*url, *dict_path, extensions, *method, *power, *verbose)
	} else {
		bruteWebSite(*url, "data/dicc.txt", extensions, *method, *power, *verbose)
	}
}
