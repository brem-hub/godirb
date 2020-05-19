package main

import (
	"flag"
	"fmt"
	"os"
)

/*TODO:
create output file#
change colors#
change dicts#
create http, https changes (default https)
create recursion?
create bot
create POST
create log
create soft exit#
*/
var (
	url         = flag.String("u", "", "specify url to run. Usage: -u <url>")
	custom_dict = flag.String("cd", "", "use custom dictionary. Usage: -cd <path>")
	verbose     = flag.Bool("v", false, "show more information(each request)")
	method      = flag.String("m", "get", "specify method to use [get, post](post is not supported for now)")
	file        = flag.String("f", "", "specify file to output into. Usage: -f <path>")
	power       = flag.Int("p", 1, "amount of goroutines. Usage -p [1...5]")
	extensions  StringSlice
)

func main() {
	dict := "data/dicc.txt"
	flag.Var(&extensions, "e", "extensions to pass. Usage: -e=php,txt,rcc")
	flag.Parse()
	if *url == "" {
		fmt.Println("specify url to run, usage: -u <url>")
		os.Exit(1)
	}
	if *custom_dict != "" {
		dict = *custom_dict
	}
	if *file == "" {
		bruteWebSite(*url, dict, extensions, *method, *power, *verbose, os.Stdout)
	} else {
		file, err := os.Create(*file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		bruteWebSite(*url, dict, extensions, *method, *power, *verbose, file)

	}
}
