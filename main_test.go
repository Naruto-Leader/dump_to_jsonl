package main

import (
	"os"
	"syscall"
	"testing"
)

func TestConvertForCsvParse(t *testing.T) {
	actual := convertForCsvParse(`'"\`)
	expected := `"<<<DQ>>>"`
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

func BenchmarkPrintInsertStatementAsJsonl(b *testing.B) {
	defer func(stdout *os.File) {
		os.Stdout = stdout
	}(os.Stdout)
	os.Stdout = os.NewFile(uintptr(syscall.Stdin), os.DevNull)

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
}
