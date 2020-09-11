/// 2>/dev/null ; exec gorun "$0" "$@"

package main

import (
	. "fmt"
	"github.com/docopt/docopt-go"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	//"bufio"
)

func main() {
	usage := `Prefixer is a general tool that allows you to manipulate records stored in a string format.

The string '\x00' will be converted to the null character in all input strings. (No escape mechanism has been implemented for this yet.)
This is because there does not seem to be a way to pass this character as an argument on the OS level.

The separators are by default the newline character '\n'.

When tracking is enabled, the magic string 'PREFIXER_LINENUMBER' in --add-prefix will be replaced with the line number of the current record.

Usage:
  prefixer [options] 
  prefixer rm [options --] [<record>...]
  prefixer -h | --help

  rm: will skip the records supplied (i.e., remove those records from the output). This happens after potentially trimming the record.

Options:
  -a --add-prefix=<add-prefix>  Adds this prefix to the beginning of each record.
  -r --remove-prefix=<rm-prefix>  Removes this prefix from the beginning of each record.
  -s --skip-empty  Skip empty records (after the removal of --remove-prefix).
  -i --input-sep=<isep>  Input record separator.
  -o --output-sep=<osep>  Output record separator.
  -t --trim  Trims whitespace from around each record before other transformations have been done.
  -l --location=<loc-file>  Enables tracking the starting line number of each record, and prints those numbers to the supplied file (separated by newlines). Use /dev/null to just enable the tracking.
  -h --help  Show this screen.`

	debug := os.Getenv("DEBUGME") != ""
	arguments, _ := docopt.ParseDoc(usage)
	if debug {
		log.Println(os.Args)
		log.Println(arguments)
	}

	rmMode := arguments["rm"].(bool)
	trimMode := arguments["--trim"].(bool)
	recordsArgs := arguments["<record>"].([]string)
	recordsArgsSet := make(map[string]struct{}, len(recordsArgs))
	for _, s := range recordsArgs {
		recordsArgsSet[s] = struct{}{}
	}

	var skipEmpty bool = false
	if arguments["--skip-empty"] != nil {
		skipEmpty = arguments["--skip-empty"].(bool)
	}

	var isep string
	if arguments["--input-sep"] != nil {
		isep = arguments["--input-sep"].(string)
		isep = strings.ReplaceAll(isep, `\x00`, "\x00")
	} else {
		isep = "\n"
	}

	var osep string
	if arguments["--output-sep"] != nil {
		osep = arguments["--output-sep"].(string)
		osep = strings.ReplaceAll(osep, `\x00`, "\x00")
	} else {
		osep = "\n"
	}

	var addPrefix string
	if arguments["--add-prefix"] != nil {
		addPrefix = arguments["--add-prefix"].(string)
		addPrefix = strings.ReplaceAll(addPrefix, `\x00`, "\x00")
	} else {
		addPrefix = ""
	}
	if debug {
		log.Println("a:" + addPrefix)
	}

	var rmPrefix string
	if arguments["--remove-prefix"] != nil {
		rmPrefix = arguments["--remove-prefix"].(string)
		rmPrefix = strings.ReplaceAll(rmPrefix, `\x00`, "\x00")
	} else {
		rmPrefix = ""
	}

	var locationPath string = ""
	var locationData strings.Builder
	locationMode := false
	if arguments["--location"] != nil {
		locationPath = arguments["--location"].(string)
		locationMode = true
	}
	linesInIsep := strings.Count(isep, "\n")
	lastLocation := 1

	inBytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalln(err.Error())
	}
	input := string(inBytes)
	records := strings.Split(input, isep)
	//recordsLenr := len(records) - 1
	isFirst := true
	for i, rec := range records {
		currAddPrefix := addPrefix
		if trimMode {
			rec = strings.Trim(rec, " \t\n\r")
		}
		rec = strings.TrimPrefix(rec, rmPrefix)
		//if locationMode { // redundant check
		if i != 0 {
			lastLocation += linesInIsep
		}
		//}
		if skipEmpty && rec == "" {
			continue
		}
		if rmMode {
			_, exists := recordsArgsSet[rec]
			if exists {
				continue
			}
		}
		//if i == recordsLenr {
		//	osep = ""
		//}
		if locationMode {
			lastLocationStr := strconv.Itoa(lastLocation)
			locationData.WriteString(lastLocationStr + "\n")
			currAddPrefix = strings.ReplaceAll(addPrefix, "PREFIXER_LINENUMBER", lastLocationStr)
			lastLocation += strings.Count(rec, "\n")
		}
		if isFirst {
			Print(currAddPrefix + rec)
		} else {
			Print(osep + currAddPrefix + rec)
		}
		isFirst = false
	}

	if locationMode && locationPath != "/dev/null" {
		err = ioutil.WriteFile(locationPath, []byte(locationData.String()), 0644)
		check(err)
		//f, err := os.Create(locationPath)
		//check(err)
		//defer f.Close()
		//w := bufio.NewWriter(f)
		//_, err = Print(w, locationData.String())
		//check(err)
		//err = w.Flush()
		//check(err)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
