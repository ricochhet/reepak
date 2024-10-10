package reepak

import (
	"github.com/ricochhet/readwrite"
	"github.com/ricochhet/simplefs"
	"github.com/ricochhet/simplelog"
)

func CompressPakData(path string) error {
	writer, err := readwrite.NewWriter(path, true)
	if err != nil {
		return err
	}

	reader, err := readwrite.NewReader(path + ".data")
	if err != nil {
		simplelog.SharedLogger.Errorf("error reading file: %s", err)
		return err
	}

	data, err := ReadData(reader)
	if err != nil {
		return err
	}

	if _, err := writer.SeekFromEnd(0); err != nil {
		return err
	}

	if err := WriteData(writer, data); err != nil {
		return err
	}

	if err := reader.Close(); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	if err := simplefs.DeleteDirectory(simplefs.GetRelativePath(path + ".data")); err != nil {
		return err
	}

	return nil
}

//nolint:cyclop // wontfix
func DecompressPakData(path string) error {
	reader, err := readwrite.NewReader(path)
	if err != nil {
		return err
	}

	writer, err := readwrite.NewWriter(path+".data", false)
	if err != nil {
		return err
	}

	data, err := ReadData(reader)
	if err != nil {
		return err
	}

	if err := WriteData(writer, data); err != nil {
		return err
	}

	writer.Close()

	if _, err := reader.SeekFromEnd(-8); err != nil {
		return err
	}

	dataSize, _ := reader.ReadUInt64()

	if _, err := reader.Seek(0, 0); err != nil {
		return err
	}

	size, _ := reader.Size()
	decompSize := size - (int64(dataSize) - 8) //nolint:gosec,mnd // wontfix
	buffer := make([]byte, decompSize)

	if _, err := reader.Read(buffer); err != nil {
		return err
	}

	if err := reader.Close(); err != nil {
		return err
	}

	decomp, err := readwrite.NewWriter(path, false)
	if err != nil {
		return err
	}

	if _, err := decomp.Write(buffer); err != nil {
		return err
	}

	if err := decomp.Close(); err != nil {
		return err
	}

	return err
}
