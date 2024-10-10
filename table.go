package reepak

import (
	"github.com/ricochhet/readwrite"
)

func WriteData(writer *readwrite.Writer, data []readwrite.DataEntry) error {
	startPos, _ := writer.Position()

	for _, entry := range data {
		if err := writer.WriteUInt32(entry.Hash); err != nil {
			return err
		}

		if _, err := writer.WriteChar(entry.FileName + "\000"); err != nil {
			return err
		}
	}

	endPos, _ := writer.Position()

	if err := writer.WriteUInt64(uint64(endPos - startPos)); err != nil { //nolint:gosec // wontfix
		return err
	}

	return nil
}

func ReadData(reader *readwrite.Reader) ([]readwrite.DataEntry, error) {
	var data []readwrite.DataEntry

	if _, err := reader.SeekFromEnd(-8); err != nil {
		return nil, err
	}

	dataSize, _ := reader.ReadUInt64()

	if _, err := reader.SeekFromEnd(int64(-dataSize - 8)); err != nil { //nolint:gosec,mnd // wontfix
		return nil, err
	}

	pos, _ := reader.Position()
	size, _ := reader.Size()

	for pos < size-8 {
		pos, _ = reader.Position()
		hash, _ := reader.ReadUInt32()

		var fileName string

		for {
			c, _ := reader.ReadChar()
			if c == '\000' {
				break
			}

			fileName += string(c)
		}

		data = append(data, readwrite.DataEntry{Hash: hash, FileName: fileName})
	}

	return data, nil
}
