package collection

import (
	"bufio"
	"github.com/adamcolton/err"
	"io"
	"os"
	"strings"
)

func LoadConfig(pathStr string) map[string]string {
	settings := make(map[string]string)
	if configFile, e := os.Open(pathStr); err.Check(e) {
		defer configFile.Close()
		reader := bufio.NewReader(configFile)
		for {
			lineBytes, e := reader.ReadBytes('\n')
			if e != io.EOF && e != nil {
				err.Panic(e)
			}
			line := trimWs(string(lineBytes))
			setting := strings.SplitN(line, ":", 2)
			if len(line) > 0 && line[0] != '#' && len(setting) == 2 {
				settings[trimWs(strings.ToLower(setting[0]))] = trimWs(setting[1])
			}
			if e == io.EOF {
				break
			}
		}
	}
	return settings
}

func trimWs(str string) string {
	return strings.Trim(str, " \n\t")
}

// StringList converts a string into a list of strings
func StringList(strLst string) []string {
	out := make([]string, 0)
	s := 0
	skip := false
	str := ""
	for i, r := range strLst {
		if skip {
			str += strLst[s : i-1]
			s = i
			skip = false
		} else if r == '\\' {
			skip = true
		} else if r == ',' {
			out = append(out, trimWs(str+strLst[s:i]))
			s = i + 1
			str = ""
		}
	}
	out = append(out, trimWs(strLst[s:]))
	return out
}

var toLower = strings.ToLower
