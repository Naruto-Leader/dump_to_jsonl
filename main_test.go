package main

import (
	"os"
	"syscall"
	"testing"
)

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
