/// 2>/dev/null ; exec gorun "$0" "$@"

package main

import (
	. "fmt"
	"github.com/docopt/docopt-go"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	usage := `Prefixer is a general tool that allows you to manipulate records stored in a string format.

Usage:
  prefixer [--add-prefix=<add-prefix> --remove-prefix=<rm-prefix> --input-sep=<isep> --output-sep=<osep> --skip-empty] 
  prefixer -h | --help

Options:
  -a --add-prefix=<add-prefix>  Adds this prefix to the beginning of each record.
  -r --remove-prefix=<rm-prefix>  Removes this prefix from the beginning of each record.
  -s --skip-empty  Skip empty records.
  -i --input-sep=<isep>  Input separator.
  -o --output-sep=<osep>  Input separator.
  -h --help     Show this screen.`

	debug := os.Getenv("DEBUGME") != ""
	arguments, _ := docopt.ParseDoc(usage)
	if debug {
		log.Println(arguments)
	}

	var skipEmpty bool = false
	if arguments["--skip-empty"] != nil {
		skipEmpty = arguments["--skip-empty"].(bool)
	}

	var isep string
	if arguments["--input-sep"] != nil {
		isep = arguments["--input-sep"].(string)
	} else {
		isep = "\n"
	}

	var osep string
	if arguments["--output-sep"] != nil {
		osep = arguments["--output-sep"].(string)
	} else {
		osep = "\n"
	}

	var addPrefix string
	if arguments["--add-prefix"] != nil {
		addPrefix = arguments["--add-prefix"].(string)
	} else {
		addPrefix = ""
	}
	if debug {
		log.Println("a:" + addPrefix)
	}
	
	var rmPrefix string
	if arguments["--remove-prefix"] != nil {
		rmPrefix = arguments["--remove-prefix"].(string)
	} else {
		rmPrefix = ""
	}

	inBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalln(err.Error())
	}
	input := string(inBytes)
	records := strings.Split(input, isep)
	for _, rec := range records {
		rec = strings.TrimPrefix(rec, rmPrefix)
		if skipEmpty && rec == "" {
			continue
		}
		Print(addPrefix + rec + osep)
	}
}