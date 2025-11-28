package fileworker

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testData struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestNewFileWriter_EmptyFileName(t *testing.T) {
	_, err := NewFileWriter("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fileName is empty")
}

func TestNewFileWriter_Success(t *testing.T) {
	tmpFile := t.TempDir() + "/test_output.json"
	writer, err := NewFileWriter(tmpFile)
	require.NoError(t, err)
	require.NotNil(t, writer)
	require.NoError(t, writer.Close())

	_, err = os.Stat(tmpFile)
	require.NoError(t, err)
}

func TestFileWriter_Write_Success(t *testing.T) {
	tmpFile := t.TempDir() + "/test_write.json"
	writer, err := NewFileWriter(tmpFile)
	require.NoError(t, err)

	data := testData{Name: "metric1", Value: 42}
	err = writer.Write(data)
	require.NoError(t, err)

	require.NoError(t, writer.Close())

	content, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	var readData testData
	err = json.Unmarshal(content[:len(content)-1], &readData)
	require.NoError(t, err)

	assert.Equal(t, data, readData)
}

func TestFileWriter_Write_MultipleRecords(t *testing.T) {
	tmpFile := t.TempDir() + "/test_multiple.json"
	writer, err := NewFileWriter(tmpFile)
	require.NoError(t, err)

	data1 := testData{Name: "first", Value: 1}
	data2 := testData{Name: "second", Value: 2}

	err = writer.Write(data1)
	require.NoError(t, err)

	err = writer.Write(data2)
	require.NoError(t, err)

	require.NoError(t, writer.Close())

	content, err := os.ReadFile(tmpFile)
	require.NoError(t, err)

	lines := strings.Split(string(content), "\n")

	var read1, read2 testData
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &read1))
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &read2))

	assert.Equal(t, data1, read1)
	assert.Equal(t, data2, read2)
}

func TestFileWriter_Close_Success(t *testing.T) {
	tmpFile := t.TempDir() + "/test_close.json"
	writer, err := NewFileWriter(tmpFile)
	require.NoError(t, err)

	err = writer.Write(testData{Name: "close-test", Value: 99})
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)

}
