package main

import (
	"bytes"
	"io"
	"os"
	"syscall"
	"testing"
)

func captureOutput(f func() error) (string, error) {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	stdout := os.Stdout
	stderr := os.Stderr
	os.Stdout = w
	os.Stderr = w

	ferr := f()

	os.Stdout = stdout
	os.Stderr = stderr
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String(), ferr
}

func TestConvertForCsvParse(t *testing.T) {
	actual := convertForCsvParse(`'"\\"`)
	expected := `"<<<DQ>>>"<<<DQ>>>`
	if actual != expected {
		t.Fatalf("%v not match %v", actual, expected)
	}
}

func TestDeconvertForCsvParse(t *testing.T) {
	actual := deconvertForCsvParse(`"<<<DQ>>>`)
	expected := `'"`
	if actual != expected {
		t.Fatalf("%v not match %v", actual, expected)
	}
}

func TestGetDataType(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		tests := []string{
			"`id` int(11) NOT NULL",
			"`id` tinyint(4) NOT NULL",
			"`id` INT(11) NOT NULL",
			"`id` TINYINT(4) NOT NULL",

			"`id` smallint(6) NOT NULL",
			"`id` mediumint(9) NOT NULL",
			"`id` bigint(20) NOT NULL",
		}
		for _, input := range tests {
			actual := getDataType(input)
			expected := DataTypeInt
			if actual != expected {
				t.Fatalf("input:%s\n%v not match %v", input, actual, expected)
			}
		}
	})

	t.Run("float", func(t *testing.T) {
		tests := []string{
			"`id` double NOT NULL",
			"`id` float NOT NULL",
			"`id` decimal(10, 2) NOT NULL",
			"`id` DOUBLE NOT NULL",
			"`id` FLOAT NOT NULL",
			"`id` DECIMAL(10, 2) NOT NULL",
		}
		for _, input := range tests {
			actual := getDataType(input)
			expected := DataTypeFloat
			if actual != expected {
				t.Fatalf("input: %s\n%v not match %v", input, actual, expected)
			}
		}
	})

	t.Run("string", func(t *testing.T) {
		tests := []string{
			"`id` varchar(255)",
			"`id` text",
			"`id` VARCHAR(255)",
			"`id` TEXT",

			"`introduction` varchar(255)",        // Contains 'int'.
			"`floating_roof` varchar(255)",       // Contains 'float'.
			"`double_angle` varchar(255)",        // Contains 'double'.
			"`circulating_decimal` varchar(255)", // Contains 'decimal'.
		}
		for _, input := range tests {
			actual := getDataType(input)
			expected := DataTypeString
			if actual != expected {
				t.Fatalf("input: %s\n%v not match %v", input, actual, expected)
			}
		}
	})
}

func TestPrintInsertStatementAsJsonl(t *testing.T) {
	t.Run("test_table", func(t *testing.T) {
		columns := []Colmun{
			{"id", DataTypeInt},
			{"name", DataTypeString},
			{"description", DataTypeString},
			{"category_id", DataTypeInt},
			{"rate", DataTypeFloat},
			{"created_at", DataTypeString},
		}
		insertStatement := `INSERT INTO test_table VALUES (1,'name1','description1,\'A\':"A"',1,1.1,'2020-09-09 10:02:35'),(2,'name2','description2,\'B\':"B"',2,2.2,'2020-09-09 10:02:46');`
		actual, err := captureOutput(func() error {
			return printInsertStatementAsJsonl(insertStatement, columns)
		})

		if err != nil {
			t.Fatal("err != nil")
		}
		expected := `{"category_id":1,"created_at":"2020-09-09 10:02:35","description":"description1,'A':\"A\"","id":1,"name":"name1","rate":1.1}
{"category_id":2,"created_at":"2020-09-09 10:02:46","description":"description2,'B':\"B\"","id":2,"name":"name2","rate":2.2}
`
		if actual != expected {
			t.Fatalf("%v not match %v", actual, expected)
		}
	})

	t.Run("json_table", func(t *testing.T) {
		columns := []Colmun{
			{"id", DataTypeInt},
			{"json", DataTypeString},
		}
		insertStatement := `INSERT INTO json_table VALUES (1,'{\"key\": \"value\"}'),(2,'{\"no\": 1}');
`
		actual, err := captureOutput(func() error {
			return printInsertStatementAsJsonl(insertStatement, columns)
		})

		if err != nil {
			t.Fatal("err != nil")
		}
		expected := `{"id":1,"json":"{\"key\": \"value\"}"}
{"id":2,"json":"{\"no\": 1}"}
`
		if actual != expected {
			t.Fatalf("%v not match %v", actual, expected)
		}
	})
}

func BenchmarkPrintInsertStatementAsJsonl(b *testing.B) {
	defer func(stdout *os.File) {
		os.Stdout = stdout
	}(os.Stdout)
	os.Stdout = os.NewFile(uintptr(syscall.Stdin), os.DevNull)

	b.Run("test_table", func(b *testing.B) {
		columns := []Colmun{
			{"id", DataTypeInt},
			{"name", DataTypeString},
			{"description", DataTypeString},
			{"category_id", DataTypeInt},
			{"rate", DataTypeFloat},
			{"created_at", DataTypeString},
		}
		insertStatement := `INSERT INTO test_table VALUES (1,'name1','description1,\'A\':"A"',1,1.1,'2020-09-09 10:02:35'),(2,'name2','description2,\'B\':"B"',2,2.2,'2020-09-09 10:02:46');`

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			printInsertStatementAsJsonl(insertStatement, columns)
		}
	})

	b.Run("json_table", func(b *testing.B) {
		columns := []Colmun{
			{"id", DataTypeInt},
			{"json", DataTypeString},
		}
		insertStatement := `INSERT INTO json_table VALUES (1,'{\"key\": \"value\"}'),(2,'{\"no\": 1}');
`
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			printInsertStatementAsJsonl(insertStatement, columns)
		}
	})
}
