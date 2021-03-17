package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"path/filepath"
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
	`\"`, DQ,
	`"`, DQ,
	`'`, `"`,
	`\`, `"`,
)

var (
	dumpFilePath string
	outputDir    string
)

func flagInit() *flag.FlagSet {
	flg := flag.NewFlagSet("dump_to_jsonl", flag.ExitOnError)
	flg.StringVar(&dumpFilePath, "file", "", "dump file")
	flg.StringVar(&outputDir, "outdir", "", "output directory")
	return flg
}

func convertForCsvParse(str string) string {
	return convertForCsvParseReplacer.Replace(str)
}

func deconvertForCsvParse(str string) string {
	return strings.Replace(
		strings.Replace(str, `"`, `'`, -1),
		DQ, `"`, -1,
	)
}

func getDataType(str string) DataType {
	s := strings.ToLower(str)
	switch {
	case strings.Contains(s, "int("):
		// case strings.Contains(s, " tinyint"),
		// strings.Contains(s, " smallint"),
		// strings.Contains(s, " mediumint"),
		// strings.Contains(s, " int"),
		// strings.Contains(s, " bigint"):
		return DataTypeInt
	case strings.Contains(s, " float"):
		return DataTypeFloat
	case strings.Contains(s, " double"):
		return DataTypeFloat
	case strings.Contains(s, " decimal"):
		return DataTypeFloat
	}
	return DataTypeString
}

func getTableName(st string) string {
	startQuoteIndex := strings.Index(st, "`")
	if startQuoteIndex != 13 { // 'CREATE TABLE' `<table_name>`
		return ""
	}
	endQuoteIndex := startQuoteIndex + 1 + strings.Index(st[startQuoteIndex+1:], "`") + 1
	return st[startQuoteIndex+1 : endQuoteIndex-1]
}

func getColumn(st string) *Colmun {
	startQuoteIndex := strings.Index(st, "`")
	if startQuoteIndex < 0 || startQuoteIndex > 3 {
		var c *Colmun
		return c
	}

	endQuoteIndex := startQuoteIndex + 1 + strings.Index(st[startQuoteIndex+1:], "`") + 1
	dataTypeStartIndex := endQuoteIndex + 1
	dataTypeEndIndex := dataTypeStartIndex + strings.Index(st[dataTypeStartIndex:], " ")
	return &Colmun{
		name:     st[startQuoteIndex+1 : endQuoteIndex-1],
		dataType: getDataType(st[endQuoteIndex:dataTypeEndIndex]),
	}
}

func printInsertStatementAsJsonl(w io.Writer, insertStatement string, columns []*Colmun) error {
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
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		jsonData := map[string]interface{}{}
		for i, v := range values {
			switch columns[i].dataType {
			case DataTypeInt:
				num, _ := strconv.Atoi(v)
				jsonData[columns[i].name] = num
			case DataTypeFloat:
				num, _ := strconv.ParseFloat(v, 64)
				jsonData[columns[i].name] = num
			default:
				jsonData[columns[i].name] = deconvertForCsvParse(v)
			}
		}
		json, _ := json.Marshal(jsonData)
		fmt.Fprintf(w, "%s\n", json)
	}
	return nil
}

func run(args []string) int {
	flg := flagInit()
	if err := flg.Parse(args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		flg.Usage()
		return 1
	}

	var rd io.Reader

	if dumpFilePath != "" {
		f, err := os.Open(dumpFilePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		defer f.Close()
		rd = f
	} else {
		if terminal.IsTerminal(0) {
			flg.Usage()
			return 1
		}
		rd = os.Stdin
	}

	if outputDir != "" {
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 1
			}
		}
	}

	reader := bufio.NewReader(rd)
	inCreateStatement := false
	columns := []*Colmun{}
	tableName := ""
	w := os.Stdout
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
			columns = []*Colmun{}
			tableName = getTableName(line)
			if tableName == "" {
				fmt.Fprintln(os.Stderr, "Failed to get table name.")
			}
			if outputDir != "" {
				var err error
				if w, err = os.Create(filepath.Join(outputDir, tableName+".jsonl")); err != nil {
					fmt.Fprintln(os.Stderr, err)
					return 1
				}
			}
			continue
		}

		if inCreateStatement {
			if strings.HasPrefix(line, ")") {
				inCreateStatement = false
				continue
			}

			column := getColumn(line)
			if column == nil {
				inCreateStatement = false
				continue
			}
			columns = append(columns, column)
		}

		if strings.HasPrefix(line, "INSERT") {
			if err := printInsertStatementAsJsonl(w, line, columns); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}

	return 0
}

func main() {
	os.Exit(run(os.Args))
}
