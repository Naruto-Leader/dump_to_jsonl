package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"runtime"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertForCsvParse(t *testing.T) {
	actual := convertForCsvParse(`'"\\"`)
	expected := `"<<<DQ>>>"<<<DQ>>>`
	assert.Equal(t, expected, actual)
}

func TestDeconvertForCsvParse(t *testing.T) {
	actual := deconvertForCsvParse(`"<<<DQ>>>`)
	expected := `'"`
	assert.Equal(t, expected, actual)
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
			assert.Equal(t, expected, actual, input)
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
			assert.Equal(t, expected, actual, input)
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
			assert.Equal(t, expected, actual, input)
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
			assert.Equal(t, expected.name, actual.name, input)
			assert.Equal(t, expected.dataType, actual.dataType, input)
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
			assert.Equal(t, expected.name, actual.name, input)
			assert.Equal(t, expected.dataType, actual.dataType, input)
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
			assert.Equal(t, expected.name, actual.name, input)
			assert.Equal(t, expected.dataType, actual.dataType, input)
		}
	})

	t.Run("Invalid value", func(t *testing.T) {
		tests := []string{
			"id varchar(255)",
			"111",
			"hoge",
		}
		var expected *Colmun
		for _, input := range tests {
			actual := getColumn(input)
			assert.Equal(t, expected, actual, input)
		}
	})
}

func TestGetTableName(t *testing.T) {
	t.Run("get table name", func(t *testing.T) {
		actual := getTableName("CREATE TABLE `table_name` (")
		expected := "table_name"
		assert.Equal(t, expected, actual)
	})

	t.Run("err", func(t *testing.T) {
		actual := getTableName("DROP TABLE `table_name`;")
		expected := ""
		assert.Equal(t, expected, actual)
	})
}

func TestPrintInsertStatementAsJsonl(t *testing.T) {
	t.Run("test_table", func(t *testing.T) {
		t.Run("No line break", func(t *testing.T) {
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
			assert.Nil(t, err)

			expected := `{"category_id":1,"created_at":"2020-09-09 10:02:35","description":"description1,'A':\"A\"","id":1,"name":"name1","rate":1.1}
{"category_id":2,"created_at":"2020-09-09 10:02:46","description":"description2,'B':\"B\"","id":2,"name":"name2","rate":2.2}
`
			assert.Equal(t, expected, w.String())
		})

		t.Run(`ends with "\n"`, func(t *testing.T) {
			columns := []*Colmun{
				{"id", DataTypeInt},
				{"name", DataTypeString},
				{"description", DataTypeString},
				{"category_id", DataTypeInt},
				{"rate", DataTypeFloat},
				{"created_at", DataTypeString},
			}
			insertStatement := `INSERT INTO test_table VALUES (1,'name1','description1,\'A\':"A"',1,1.1,'2020-09-09 10:02:35'),(2,'name2','description2,\'B\':"B"',2,2.2,'2020-09-09 10:02:46');` + "\n"
			w := new(bytes.Buffer)
			err := printInsertStatementAsJsonl(w, insertStatement, columns)
			assert.Nil(t, err)

			expected := `{"category_id":1,"created_at":"2020-09-09 10:02:35","description":"description1,'A':\"A\"","id":1,"name":"name1","rate":1.1}
{"category_id":2,"created_at":"2020-09-09 10:02:46","description":"description2,'B':\"B\"","id":2,"name":"name2","rate":2.2}
`
			assert.Equal(t, expected, w.String())
		})

		t.Run(`ends with "\r\n"`, func(t *testing.T) {
			columns := []*Colmun{
				{"id", DataTypeInt},
				{"name", DataTypeString},
				{"description", DataTypeString},
				{"category_id", DataTypeInt},
				{"rate", DataTypeFloat},
				{"created_at", DataTypeString},
			}
			insertStatement := `INSERT INTO test_table VALUES (1,'name1','description1,\'A\':"A"',1,1.1,'2020-09-09 10:02:35'),(2,'name2','description2,\'B\':"B"',2,2.2,'2020-09-09 10:02:46');` + "\r\n"
			w := new(bytes.Buffer)
			err := printInsertStatementAsJsonl(w, insertStatement, columns)
			assert.Nil(t, err)

			expected := `{"category_id":1,"created_at":"2020-09-09 10:02:35","description":"description1,'A':\"A\"","id":1,"name":"name1","rate":1.1}
{"category_id":2,"created_at":"2020-09-09 10:02:46","description":"description2,'B':\"B\"","id":2,"name":"name2","rate":2.2}
`
			assert.Equal(t, expected, w.String())
		})
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
		assert.Nil(t, err)

		expected := `{"id":1,"json":"{\"key\": \"value\"}"}
{"id":2,"json":"{\"no\": 1}"}
`
		assert.Equal(t, expected, w.String())
	})
}

func TestPrintInsertStatementAsJsonlError(t *testing.T) {
	columns := []*Colmun{
		{"id", DataTypeInt},
		{"name", DataTypeString},
	}
	// I don't think it is possible for the data in the dump file to look like this.
	insertStatement := `INSERT INTO test_table VALUES (1,'name1'),(2);`
	w := new(bytes.Buffer)
	err := printInsertStatementAsJsonl(w, insertStatement, columns)
	assert.NotNil(t, err)
}

func captureRunOutput(f func() int) (string, int) {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	stdout := os.Stdout
	stderr := os.Stderr
	os.Stdout = w
	os.Stderr = w

	ret := f()

	os.Stdout = stdout
	os.Stderr = stderr
	w.Close()

	var buf bytes.Buffer
	io.Copy(&buf, r)

	return buf.String(), ret
}

func TestRun(t *testing.T) {
	t.Run("one_create", func(t *testing.T) {
		args := []string{
			"",
			"-file",
			"./test/fixtures/one_create.sql",
		}
		output, ret := captureRunOutput(func() int { return run(args) })
		assert.Equal(t, 0, ret)

		expected := `{"category_id":1,"created_at":"2020-09-09 10:02:35","description":"description1,'A':\"A\"","id":1,"name":"name1","rate":1.1}
{"category_id":2,"created_at":"2020-09-09 10:02:46","description":"description2,'B':\"B\"","id":2,"name":"name2","rate":2.2}
`
		assert.Equal(t, expected, output)
	})

	t.Run("one_create_does_not_have_a_key", func(t *testing.T) {
		args := []string{
			"",
			"-file",
			"./test/fixtures/one_create_does_not_have_a_key.sql",
		}
		output, ret := captureRunOutput(func() int { return run(args) })
		assert.Equal(t, 0, ret)

		expected := `{"category_id":1,"created_at":"2020-09-09 10:02:35","description":"description1,'A':\"A\"","id":1,"name":"name1","rate":1.1}
{"category_id":2,"created_at":"2020-09-09 10:02:46","description":"description2,'B':\"B\"","id":2,"name":"name2","rate":2.2}
`
		assert.Equal(t, expected, output)
	})

	t.Run("two_create", func(t *testing.T) {
		args := []string{
			"",
			"-file",
			"./test/fixtures/two_create.sql",
		}
		output, ret := captureRunOutput(func() int { return run(args) })
		assert.Equal(t, 0, ret)

		expected := `{"id":1,"json":"{\"key\": \"value\"}"}
{"id":2,"json":"{\"no\": 1}"}
{"category_id":1,"created_at":"2020-09-09 10:02:35","description":"description1,'A':\"A\"","id":1,"name":"name1","rate":1.1}
{"category_id":2,"created_at":"2020-09-09 10:02:46","description":"description2,'B':\"B\"","id":2,"name":"name2","rate":2.2}
`
		assert.Equal(t, expected, output)
	})
}

func TestRunError(t *testing.T) {
	type TestValues struct {
		PathForErrorTest     string
		ExpectedErrorMessage string
	}

	t.Run("Failed to get table name.", func(t *testing.T) {
		args := []string{
			"",
			"-file",
			"./test/fixtures/failed_to_get_table_name.sql",
		}
		output, ret := captureRunOutput(func() int { return run(args) })
		assert.Equal(t, 1, ret)

		expected := "Failed to get table name.\n"
		assert.Equal(t, expected, output)
	})

	t.Run("Failed to load dumpfile.", func(t *testing.T) {
		testValues := &TestValues{
			PathForErrorTest:     "",
			ExpectedErrorMessage: "no such file or directory",
		}
		if runtime.GOOS == "windows" {
			testValues.ExpectedErrorMessage = "The system cannot find the file specified."
		}

		args := []string{
			"",
			"-file",
			"HOGE",
		}
		output, ret := captureRunOutput(func() int { return run(args) })
		assert.Equal(t, 1, ret)

		expected := fmt.Sprintf("open HOGE: %s\n", testValues.ExpectedErrorMessage)
		assert.Equal(t, expected, output)
	})

	t.Run("Failed: mkdir OUTPUT_DIR", func(t *testing.T) {
		testValues := &TestValues{
			PathForErrorTest:     "/AAAAAAAAAAAAAAAAA",
			ExpectedErrorMessage: "permission denied",
		}
		switch runtime.GOOS {
		case "windows":
			testValues = &TestValues{
				PathForErrorTest:     "M:/",
				ExpectedErrorMessage: "The system cannot find the path specified.",
			}
		case "darwin":
			testValues.ExpectedErrorMessage = "read-only file system"
		}

		args := []string{
			"",
			"-file",
			"./test/fixtures/two_create.sql",
			"-outdir",
			testValues.PathForErrorTest,
		}
		output, ret := captureRunOutput(func() int { return run(args) })
		assert.Equal(t, 1, ret)

		expected := fmt.Sprintf(
			"mkdir %s: %s\n",
			testValues.PathForErrorTest,
			testValues.ExpectedErrorMessage,
		)
		assert.Equal(t, expected, output)
	})

	t.Run("Failed: open OUTPUT_FILE", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skip this test on windows.")
			return
		}

		testValues := &TestValues{
			PathForErrorTest:     "/",
			ExpectedErrorMessage: "permission denied",
		}
		if runtime.GOOS == "darwin" {
			testValues.ExpectedErrorMessage = "read-only file system"
		}

		args := []string{
			"",
			"-file",
			"./test/fixtures/two_create.sql",
			"-outdir",
			testValues.PathForErrorTest,
		}
		output, ret := captureRunOutput(func() int { return run(args) })
		assert.Equal(t, 1, ret)

		expected := fmt.Sprintf(
			"open %sjson_table.jsonl: %s\n",
			testValues.PathForErrorTest,
			testValues.ExpectedErrorMessage,
		)
		assert.Equal(t, expected, output)
	})
}

// Benchmark
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

func BenchmarkRun(b *testing.B) {
	defer func(stdout *os.File) {
		os.Stdout = stdout
	}(os.Stdout)
	os.Stdout = os.NewFile(uintptr(syscall.Stdin), os.DevNull)

	b.Run("one_create", func(b *testing.B) {
		args := []string{
			"",
			"-file",
			"./test/fixtures/one_create.sql",
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			run(args)
		}
	})

	b.Run("two_create", func(b *testing.B) {
		args := []string{
			"",
			"-file",
			"./test/fixtures/two_create.sql",
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			run(args)
		}
	})
}
