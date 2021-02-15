package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"strconv"
	"strings"
)

const DQ = "<<<DQ>>>"

type DataType int

const (
	DataTypeInt DataType = iota
	DataTypeFloat
	DataTypeString
)

type Colmun struct {
	name     string
	dataType DataType
}

var convertForCsvParseReplacer = strings.NewReplacer(
	`"`, DQ,
	`'`, `"`,
	`\`, `"`,
)

func convertForCsvParse(str string) string {
	return convertForCsvParseReplacer.Replace(str)
}

func deconvertForCsvParse(str string) string {
	return strings.Replace(
		strings.Replace(str, "\"", "'", -1),
		DQ, "\"", -1,
	)
}

func getDataType(str string) DataType {
	s := strings.ToLower(str)
	switch {
	case strings.Contains(s, "int"):
		return DataTypeInt
	case strings.Contains(s, "float"):
		return DataTypeFloat
	case strings.Contains(s, "double"):
		return DataTypeFloat
	case strings.Contains(s, "decimal"):
		return DataTypeFloat
	}
	return DataTypeString
}

func printInsertStatementAsJsonl(insertStatement string, columns []Colmun) error {
	stlen := len(insertStatement)
	trimLen := 2 // ");"
	if strings.Contains(insertStatement[stlen-trimLen:], "\r") {
		trimLen++
	}
	if strings.Contains(insertStatement[stlen-trimLen:], "\n") {
		trimLen++
	}

	valuesListStr := convertForCsvParse(insertStatement[strings.Index(insertStatement, "(")+1 : stlen-trimLen])
	valuesCsvStr := strings.Replace(valuesListStr, "),(", "\n", -1)
	cr := csv.NewReader(strings.NewReader(valuesCsvStr))
	for {
		values, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		jsonData := map[string]interface{}{}
		for i, v := range values {
			switch {
			case columns[i].dataType == DataTypeInt:
				num, _ := strconv.Atoi(v)
				jsonData[columns[i].name] = num
			case columns[i].dataType == DataTypeFloat:
				num, _ := strconv.ParseFloat(v, 64)
				jsonData[columns[i].name] = num
			default:
				jsonData[columns[i].name] = deconvertForCsvParse(v)
			}
		}
		json, _ := json.Marshal(jsonData)
		fmt.Printf("%s\n", json)
	}
	return nil
}

func run(args []string) int {
	var rd io.Reader

	if terminal.IsTerminal(0) {
		if len(args) != 2 {
			return 1
		}
		f, err := os.Open(args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		defer f.Close()
		rd = f
	} else {
		rd = os.Stdin
	}

	reader := bufio.NewReader(rd)
	inCreateStatement := false
	columns := []Colmun{}
	for {
		lineBytes, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, err)
				return 1
			} else if err == io.EOF && len(lineBytes) == 0 {
				return 0
			}
		}

		line := string(lineBytes)

		if strings.HasPrefix(line, "CREATE") {
			inCreateStatement = true
			columns = []Colmun{}
			continue
		}

		if inCreateStatement {
			if strings.HasPrefix(line, ")") {
				inCreateStatement = false
				continue
			}

			startQuoteIndex := strings.Index(line, "`")
			if startQuoteIndex < 0 || startQuoteIndex > 3 {
				inCreateStatement = false
				continue
			}

			endQuoteIndex := startQuoteIndex + 1 + strings.Index(line[startQuoteIndex+1:], "`") + 1
			dataTypeStartIndex := endQuoteIndex + 1
			dataTypeEndIndex := dataTypeStartIndex + strings.Index(line[dataTypeStartIndex:], " ")
			columns = append(columns, Colmun{
				name:     line[startQuoteIndex+1 : endQuoteIndex-1],
				dataType: getDataType(line[endQuoteIndex:dataTypeEndIndex]),
			})
		}

		if strings.HasPrefix(line, "INSERT") {
			if err := printInsertStatementAsJsonl(line, columns); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}

	return 0
}

func main() {
	os.Exit(run(os.Args))
}
