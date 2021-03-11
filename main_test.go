package main

import (
	"bytes"
	"io"
	"testing"
)

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
		expected := DataTypeInt
		for _, input := range tests {
			actual := getDataType(input)
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
		expected := DataTypeFloat
		for _, input := range tests {
			actual := getDataType(input)
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
		expected := DataTypeString
		for _, input := range tests {
			actual := getDataType(input)
			if actual != expected {
				t.Fatalf("input: %s\n%v not match %v", input, actual, expected)
			}
		}
	})
}

func TestGetColumn(t *testing.T) {
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
		expected := &Colmun{"id", DataTypeInt}
		for _, input := range tests {
			actual := getColumn(input)
			if actual.name != expected.name {
				t.Fatalf("input: %s\n%v not match %v", input, actual.name, expected.name)
			}
			if actual.dataType != expected.dataType {
				t.Fatalf("input: %s\n%v not match %v", input, actual.dataType, expected.dataType)
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
		expected := &Colmun{"id", DataTypeFloat}
		for _, input := range tests {
			actual := getColumn(input)
			if actual.name != expected.name {
				t.Fatalf("input: %s\n%v not match %v", input, actual.name, expected.name)
			}
			if actual.dataType != expected.dataType {
				t.Fatalf("input: %s\n%v not match %v", input, actual.dataType, expected.dataType)
			}
		}
	})

	t.Run("string", func(t *testing.T) {
		tests := []string{
			"`id` varchar(255)",
			"`id` text",
			"`id` VARCHAR(255)",
			"`id` TEXT",
		}
		expected := &Colmun{"id", DataTypeString}
		for _, input := range tests {
			actual := getColumn(input)
			if actual.name != expected.name {
				t.Fatalf("input: %s\n%v not match %v", input, actual.name, expected.name)
			}
			if actual.dataType != expected.dataType {
				t.Fatalf("input: %s\n%v not match %v", input, actual.dataType, expected.dataType)
			}
		}
	})
}

func TestPrintInsertStatementAsJsonl(t *testing.T) {
	t.Run("test_table", func(t *testing.T) {
		columns := []*Colmun{
			{"id", DataTypeInt},
			{"name", DataTypeString},
			{"description", DataTypeString},
			{"category_id", DataTypeInt},
			{"rate", DataTypeFloat},
			{"created_at", DataTypeString},
		}
		insertStatement := `INSERT INTO test_table VALUES (1,'name1','description1,\'A\':"A"',1,1.1,'2020-09-09 10:02:35'),(2,'name2','description2,\'B\':"B"',2,2.2,'2020-09-09 10:02:46');`
		w := new(bytes.Buffer)
		err := printInsertStatementAsJsonl(w, insertStatement, columns)

		if err != nil {
			t.Fatal("err != nil")
		}
		expected := `{"category_id":1,"created_at":"2020-09-09 10:02:35","description":"description1,'A':\"A\"","id":1,"name":"name1","rate":1.1}
{"category_id":2,"created_at":"2020-09-09 10:02:46","description":"description2,'B':\"B\"","id":2,"name":"name2","rate":2.2}
`
		if w.String() != expected {
			t.Fatalf("%v not match %v", w.String(), expected)
		}
	})

	t.Run("json_table", func(t *testing.T) {
		columns := []*Colmun{
			{"id", DataTypeInt},
			{"json", DataTypeString},
		}
		insertStatement := `INSERT INTO json_table VALUES (1,'{\"key\": \"value\"}'),(2,'{\"no\": 1}');
`
		w := new(bytes.Buffer)
		err := printInsertStatementAsJsonl(w, insertStatement, columns)

		if err != nil {
			t.Fatal("err != nil")
		}
		expected := `{"id":1,"json":"{\"key\": \"value\"}"}
{"id":2,"json":"{\"no\": 1}"}
`
		if w.String() != expected {
			t.Fatalf("%v not match %v", w.String(), expected)
		}
	})
}

func BenchmarkPrintInsertStatementAsJsonl(b *testing.B) {
	b.Run("test_table", func(b *testing.B) {
		columns := []*Colmun{
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
			printInsertStatementAsJsonl(io.Discard, insertStatement, columns)
		}
	})

	b.Run("json_table", func(b *testing.B) {
		columns := []*Colmun{
			{"id", DataTypeInt},
			{"json", DataTypeString},
		}
		insertStatement := `INSERT INTO json_table VALUES (1,'{\"key\": \"value\"}'),(2,'{\"no\": 1}');
`
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			printInsertStatementAsJsonl(io.Discard, insertStatement, columns)
		}
	})
}
