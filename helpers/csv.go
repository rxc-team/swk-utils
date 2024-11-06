package helpers

import (
	"bytes"
	"encoding/csv"
	"io"
	"os"

	"github.com/saintfish/chardet"

	"github.com/dimchansky/utfbom"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

var (
	// FileEncodings file reader encodings
	FileEncodings = map[string]string{
		"UTF-8":     "UTF-8",
		"ShiftJIS":  "ShiftJIS",
		"Shift_JIS": "Shift_JIS",
		"shift_jis": "shift_jis",
	}

	// UTF8Encoding default encoding
	UTF8Encoding = "UTF-8"
)

// NewCSVFileReader new csv file reader
func NewCSVFileReader(encoding string, file io.Reader) (reader io.Reader) {
	switch encoding {
	case FileEncodings["UTF-8"]:
		// remove BOM if nessesary(use utfbom)
		reader = utfbom.SkipOnly(file)
	case FileEncodings["ShiftJIS"]:
		reader = transform.NewReader(file, japanese.ShiftJIS.NewDecoder())
	case FileEncodings["Shift_JIS"]:
		reader = transform.NewReader(file, japanese.ShiftJIS.NewDecoder())
	case FileEncodings["shift_jis"]:
		reader = transform.NewReader(file, japanese.ShiftJIS.NewDecoder())
	default:
		reader = file
	}

	return reader
}

// ReadCSVLines read number of rows from a csv file reader
func ReadCSVLines(csvReader *csv.Reader, numLines int, returnErr bool) (lines [][]string, err error) {
	for i := 0; i < numLines; i++ {
		var line []string
		line, err = csvReader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			if returnErr {
				return
			}
			if err != csv.ErrFieldCount {
				if len(line) < csvReader.FieldsPerRecord {
					for len(line) < csvReader.FieldsPerRecord {
						line = append(line, "")
					}
				}
			}
			err = nil
		}
		lines = append(lines, line)
	}
	return
}

// DetectFileEncoding simple detect encoding
func DetectFileEncoding(file io.ReadSeeker) string {
	decidedEncoding := UTF8Encoding
	var buff bytes.Buffer
	_, err := buff.ReadFrom(file)
	if err != nil {
		return decidedEncoding
	}

	textDetector := chardet.NewTextDetector()
	result, err := textDetector.DetectBest(buff.Bytes())
	if err != nil {
		return decidedEncoding
	}

	switch result.Charset {
	case "Shift_JIS":
		decidedEncoding = result.Charset
	default:
	}

	SeekOrigin(file)
	return decidedEncoding
}

// SeekOrigin position cursor to the very first point
func SeekOrigin(r io.Seeker) {
	r.Seek(0, os.SEEK_SET)
}
