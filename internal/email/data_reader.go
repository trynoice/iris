package email

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sync"
)

func NewDataReader(defaultFile string, dataFile string) (*DataReader, error) {
	defaultValues := map[string]string{}
	if file, err := os.Open(defaultFile); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to open default csv: %w", err)
	} else if file != nil {
		defer file.Close()
		rows, err := csv.NewReader(file).ReadAll()
		if err != nil {
			return nil, fmt.Errorf("failed to read records from default csv: %w", err)
		}

		if len(rows) > 2 {
			return nil, fmt.Errorf("default csv must not have more than 2 rows")
		}

		if len(rows) == 2 {
			defaultValues = buildMap(rows[0], rows[1])
		}
	}

	dataFh, err := os.Open(dataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open data csv: %w", err)
	}

	reader := csv.NewReader(dataFh)
	reader.ReuseRecord = true
	row, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read headers from data csv: %w", err)
	}

	headers := make([]string, len(row))
	copy(headers, row)
	return &DataReader{
		mutex:         sync.Mutex{},
		defaultValues: defaultValues,
		fileCloser:    dataFh,
		reader:        reader,
		headers:       headers,
	}, nil
}

type DataReader struct {
	mutex         sync.Mutex
	defaultValues map[string]string
	fileCloser    io.Closer
	reader        *csv.Reader
	headers       []string
}

func (r *DataReader) Read() (map[string]string, error) {
	// needs mutex because a shared buffer is used for sequentially reading csv
	// records. although never invoked in parallel, it is still a good practice.
	r.mutex.Lock()
	defer r.mutex.Unlock()

	record, err := r.reader.Read()
	if err == io.EOF {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("failed to read record from data csv: %w", err)
	}

	row := buildMap(r.headers, record)
	for key, defaultValue := range r.defaultValues {
		if value, ok := row[key]; !ok || value == "" {
			row[key] = defaultValue
		}
	}

	return row, nil
}

func (r *DataReader) Close() error {
	return r.fileCloser.Close()
}

func buildMap(keys []string, values []string) map[string]string {
	data := map[string]string{}
	for i, key := range keys {
		data[key] = values[i]
	}
	return data
}
