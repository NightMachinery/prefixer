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
	"github.com/acarl005/stripansi"
	//"bufio"
)

func main() {
	usage := `Prefixer is a general tool that allows you to manipulate records stored in a string format.

The string '\x00' will be converted to the null character in <add-prefix>, <add-postfix>, <rm-prefix>, <isep>, and <osep>. (No escape mechanism has been implemented for them yet.)
This is because there does not seem to be a way to pass this character as an argument on the OS level.

The separators are by default the newline character '\n'.

When tracking is enabled, the magic string 'PREFIXER_LINENUMBER' in <add-prefix> and <add-postfix> will be replaced with the line number of the current record.

Usage:
  prefixer [options] 
  prefixer rm [options --] [<record>...]
  prefixer replace [options --] [(<from> <to>)...]
  prefixer -h | --help

  rm: will skip the records supplied (i.e., remove those records from the output). This happens after potentially trimming the record.

Options:
  -a --add-prefix=<add-prefix>  Adds this prefix to the beginning of each record.
  -p --add-postfix=<add-postfix>  Adds this to the end of each record.
  -r --remove-prefix=<rm-prefix>  Removes this prefix from the beginning of each record.
  -s --skip-empty  Skip empty records (after the removal of --remove-prefix).
  -i --input-sep=<isep>  Input record separator.
  -o --output-sep=<osep>  Output record separator.
  -t --trim  Trims whitespace from around each record before other transformations have been done.
  --rm-ansi  Strip the ANSI color codes from input records when testing for equality in rm or replace.
  --rm-x  Enable \x00 to NUL conversion for <record>.
  --from-x  Enable \x00 to NUL conversion for <from>.
  --to-x  Enable \x00 to NUL conversion for <to>.
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
	rmAnsi := arguments["--rm-ansi"].(bool)
	rmX := arguments["--rm-x"].(bool)
	replaceMode := arguments["replace"].(bool)
	fromX := arguments["--from-x"].(bool)
	toX := arguments["--to-x"].(bool)

	recordsArgs := arguments["<record>"].([]string)
	recordsArgsSet := make(map[string]struct{}, len(recordsArgs))
	for _, s := range recordsArgs {
		if rmX {
			s = strings.ReplaceAll(s, `\x00`, "\x00")
		}
		recordsArgsSet[s] = struct{}{}
	}
	fromArgs := arguments["<from>"].([]string)
	toArgs := arguments["<to>"].([]string)
	fromTo := make(map[string]string, len(fromArgs))
	for i, from := range fromArgs {
		if fromX {
			from = strings.ReplaceAll(from, `\x00`, "\x00")
		}
		to := toArgs[i]
		if toX {
			to = strings.ReplaceAll(to, `\x00`, "\x00")
		}
		fromTo[from] = to
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

	var addPostfix string
	if arguments["--add-postfix"] != nil {
		addPostfix = arguments["--add-postfix"].(string)
		addPostfix = strings.ReplaceAll(addPostfix, `\x00`, "\x00")
	} else {
		addPostfix = ""
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
	isFirst := true
	for i, rec := range records {
		recLineCount := strings.Count(rec, "\n") // We need to save this before changing <rec>
		currAddPrefix := addPrefix
		currAddPostfix := addPostfix
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
			rmRec := rec
			if rmAnsi {
				rmRec = stripansi.Strip(rmRec)
			}
			_, exists := recordsArgsSet[rmRec]
			if exists {
				continue
			}
		}
		if locationMode {
			lastLocationStr := strconv.Itoa(lastLocation)
			locationData.WriteString(lastLocationStr + "\n")
			currAddPrefix = strings.ReplaceAll(currAddPrefix, "PREFIXER_LINENUMBER", lastLocationStr)
			currAddPostfix = strings.ReplaceAll(currAddPostfix, "PREFIXER_LINENUMBER", lastLocationStr)
			lastLocation += recLineCount
		}
		if replaceMode {
			rmRec := rec
			if rmAnsi {
				rmRec = stripansi.Strip(rmRec)
			}
			to, exists := fromTo[rmRec]
			if exists {
				rec = to
				if skipEmpty && rec == "" {
					continue
				}
			}
		}
		if isFirst {
			Print(currAddPrefix + rec + currAddPostfix)
		} else {
			Print(osep + currAddPrefix + rec + currAddPostfix)
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
