package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const DQ = "<<<DQ>>>"

func convertForCsvParse(str string) string {
	str = strings.Replace(str, "\"", DQ, -1)
	str = strings.Replace(str, "'", "\"", -1)
	return strings.Replace(str, "\\", "\"", -1)
}

func deconvertForCsvParse(str string) string {
	str = strings.Replace(str, "\"", "'", -1)
	return strings.Replace(str, DQ, "\"", -1)
}

func run(args []string) int {
	if len(args) != 2 {
		return 1
	}

	f, err := os.Open(args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inCreateStatement := false
	columns := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "CREATE") {
			inCreateStatement = true
			continue
		}

		if inCreateStatement {
			startQuoteIndex := strings.Index(line, "`")
			if startQuoteIndex < 0 || startQuoteIndex > 3 {
				inCreateStatement = false
				continue
			}

			endQuoteIndex := startQuoteIndex + 1 + strings.Index(line[startQuoteIndex+1:], "`") + 1
			columns = append(columns, line[startQuoteIndex+1:endQuoteIndex-1])

			if strings.HasPrefix(line, ")") {
				inCreateStatement = false
				continue
			}
		}

		if strings.HasPrefix(line, "INSERT") {
			valuesListStr := line[strings.Index(line, "(")+1 : len(line)-2]
			for _, valuesCsv := range strings.Split(valuesListStr, "),(") {
				valuesCsv = convertForCsvParse(valuesCsv)
				cr := csv.NewReader(strings.NewReader(valuesCsv))
				values, err := cr.Read()
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					return 1
				}

				jsonData := map[string]interface{}{}
				for i, v := range values {
					if num, err := strconv.Atoi(v); err == nil {
						jsonData[columns[i]] = num
					} else if num, err := strconv.ParseFloat(v, 64); err == nil {
						jsonData[columns[i]] = num
					} else {
						jsonData[columns[i]] = deconvertForCsvParse(v)
					}
				}
				json, _ := json.Marshal(jsonData)
				fmt.Printf("%s\n", json)
			}

			columns = []string{}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}

func main() {
	os.Exit(run(os.Args))
}
