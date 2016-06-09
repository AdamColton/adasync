package adasync

import (
	"bufio"
	"github.com/adamcolton/err"
	"io"
	"strings"
)

func LoadConfig(pathStr string) (map[string]string, error) {
	settings := make(map[string]string)
	eOut := error(nil)
	if configFile, e := filesystem.Open(pathStr); err.Check(e) {
		defer configFile.Close()
		ParseConfig(configFile, settings)
	} else {
		eOut = e
	}
	return settings, eOut
}

func ParseConfig(config io.Reader, settings map[string]string) {
	reader := bufio.NewReader(config)
	for {
		lineBytes, e := reader.ReadBytes('\n')
		if e != io.EOF && e != nil {
			err.Panic(e)
		}
		line := trimWs(string(lineBytes))
		setting := strings.SplitN(line, ":", 2)
		if len(line) > 0 && line[0] != '#' {
			if len(setting) == 1 {
				setting = append(setting, "true") // anything
			}
			key := trimWs(strings.ToLower(setting[0]))
			if key != "" {
				val := trimWs(setting[1])
				settings[key] = val
			}
		}
		if e == io.EOF {
			break
		}
	}
}

func trimWs(str string) string {
	return strings.Trim(str, " \n\t")
}

// StringList converts a string into a list of strings splitting on comma but
// allowing \, to escape the comma
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
