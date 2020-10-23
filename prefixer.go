/// 2>/dev/null ; exec gorun "$0" "$@"
// Some code forked from fzf. See https://github.com/junegunn/fzf/blob/master/LICENSE

package main

import (
	. "fmt"
	"github.com/acarl005/stripansi"
	"github.com/docopt/docopt-go"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	//"bufio"
)

///
const rangeEllipsis = 0

// Range represents nth-expression
type Range struct {
	begin int
	end   int
}

func errorExit(msg string) {
	os.Stderr.WriteString(msg + "\n")
	os.Exit(2)
}
func newRange(begin int, end int) Range {
	if begin == 1 {
		begin = rangeEllipsis
	}
	if end == -1 {
		end = rangeEllipsis
	}
	return Range{begin, end}
}

// ParseRange parses nth-expression and returns the corresponding Range object
func ParseRange(str *string) (Range, bool) {
	if (*str) == ".." {
		return newRange(rangeEllipsis, rangeEllipsis), true
	} else if strings.HasPrefix(*str, "..") {
		end, err := strconv.Atoi((*str)[2:])
		if err != nil || end == 0 {
			return Range{}, false
		}
		return newRange(rangeEllipsis, end), true
	} else if strings.HasSuffix(*str, "..") {
		begin, err := strconv.Atoi((*str)[:len(*str)-2])
		if err != nil || begin == 0 {
			return Range{}, false
		}
		return newRange(begin, rangeEllipsis), true
	} else if strings.Contains(*str, "..") {
		ns := strings.Split(*str, "..")
		if len(ns) != 2 {
			return Range{}, false
		}
		begin, err1 := strconv.Atoi(ns[0])
		end, err2 := strconv.Atoi(ns[1])
		if err1 != nil || err2 != nil || begin == 0 || end == 0 {
			return Range{}, false
		}
		return newRange(begin, end), true
	}

	n, err := strconv.Atoi(*str)
	if err != nil || n == 0 {
		return Range{}, false
	}
	return newRange(n, n), true
}
func splitNth(str string) []Range {
	if match, _ := regexp.MatchString("^[0-9,-.]+$", str); !match {
		errorExit("invalid format: " + str)
	}

	tokens := strings.Split(str, ",")
	ranges := make([]Range, len(tokens))
	for idx, s := range tokens {
		r, ok := ParseRange(&s)
		if !ok {
			errorExit("invalid format: " + str)
		}
		ranges[idx] = r
	}
	return ranges
}

///

// @todo Merge tokens according to the ranges given
//func mergeTokens(tokens []string, withNth []Range) []string {
//	transTokens := make([]string, len(withNth))
//	numTokens := len(tokens)
//	for idx, r := range withNth {
//
//	}
//}

func rangesIn(ranges []Range, totalLen int, target int) bool {
	target += 1 // one-based indexing
	for _, r := range ranges {
		beg := r.begin
		if beg == 0 {
			beg = 1
		}
		if beg < 0 {
			beg += totalLen + 1
		}
		end := r.end
		if end == 0 {
			end = -1
		}
		if end < 0 {
			end += totalLen + 1
		}
		if (target <= end) && (target >= beg) {
			return true
		}
	}
	return false
}

func main() {
	usage := `Prefixer is a general tool that allows you to manipulate records stored in a string format.

The string '\x00' will be converted to the null character in <add-prefix>, <add-postfix>, <rm-prefix>, <isep>, and <osep>. (No escape mechanism has been implemented for them yet.)
This is because there does not seem to be a way to pass this character as an argument on the OS level.

The separators are by default the newline character '\n'.

When tracking is enabled, the magic string 'PREFIXER_LINENUMBER' in <add-prefix>, <add-postfix>, and <replace> will be replaced with the line number of the current record.

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
  --process-include=<process-include>  Ranges of the input records to process. This uses fzf's range syntax. Unprocessed records will be output as they are.
  --rm-ansi  Strip the ANSI color codes from input records when testing for equality in rm or replace.
  --rm-x  Enable \x00 to NUL conversion for <record>.
  --from-x  Enable \x00 to NUL conversion for <from>.
  --to-x  Enable \x00 to NUL conversion for <to>.
  --replace=<replace>  Replace all records without matches in <from> records with <replace>. '$1' will be expanded to the original record.
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

	var processInclude []Range
	if arguments["--process-include"] != nil {
		processInclude = splitNth(arguments["--process-include"].(string))
	} else {
		//processInclude = []Range{newRange(0, 0)}
	}

	var rep string
	if arguments["--replace"] != nil {
		rep = arguments["--replace"].(string)
	} else {
		rep = ""
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
		processThis := len(processInclude) == 0 || rangesIn(processInclude, len(records), i)
		recLineCount := strings.Count(rec, "\n") // We need to save this before changing <rec>
		currAddPrefix := addPrefix
		currAddPostfix := addPostfix
		if processThis {
			if trimMode {
				rec = strings.Trim(rec, " \t\n\r")
			}
			rec = strings.TrimPrefix(rec, rmPrefix)
		} else {
			currAddPrefix = ""
			currAddPostfix = ""
		}
		//if locationMode { // redundant check
		if i != 0 {
			lastLocation += linesInIsep
		}
		//}
		if skipEmpty && rec == "" {
			continue
		}
		if processThis && rmMode {
			rmRec := rec
			if rmAnsi {
				rmRec = stripansi.Strip(rmRec)
			}
			_, exists := recordsArgsSet[rmRec]
			if exists {
				continue
			}
		}
		lastLocationStr := "PREFIXER_LINENUMBER"
		if locationMode {
			lastLocationStr = strconv.Itoa(lastLocation)
			locationData.WriteString(lastLocationStr + "\n")
			currAddPrefix = strings.ReplaceAll(currAddPrefix, "PREFIXER_LINENUMBER", lastLocationStr)
			currAddPostfix = strings.ReplaceAll(currAddPostfix, "PREFIXER_LINENUMBER", lastLocationStr)
			lastLocation += recLineCount
		}
		if processThis && replaceMode {
			rmRec := rec
			if rmAnsi {
				rmRec = stripansi.Strip(rmRec)
			}
			to, exists := fromTo[rmRec]
			if exists {
				rec = to
			} else if rep != "" {
				repTmp := strings.ReplaceAll(rep, "PREFIXER_LINENUMBER", lastLocationStr)
				rec = strings.ReplaceAll(repTmp, `$1`, rec)
			}
			if skipEmpty && rec == "" {
				continue
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
