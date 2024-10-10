package reepak

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/ricochhet/murmurhash3"
	"github.com/ricochhet/readwrite"
	"github.com/ricochhet/simplefs"
	"github.com/ricochhet/simplelog"
)

var (
	errInvalidFileFormat = errors.New("invalid file format")
	errFileTooLarge      = errors.New("file too large")
)

//nolint:funlen,gocognit,gocyclo,cyclop // wontfix
func ProcessDirectory(source, outputFile string, embed bool) error {
	directory, _ := filepath.Abs(source)
	sortedFiles := simplefs.GetFiles(filepath.Join(directory, "natives"))
	writer, err := readwrite.NewWriter(outputFile, false)

	data := []readwrite.DataEntry{}
	list := []readwrite.FileEntry{}

	if err != nil {
		return err
	}

	if err := writer.WriteUInt32(1095454795); err != nil { //nolint:mnd // wontfix
		return err
	}

	if err := writer.WriteUInt32(4); err != nil { //nolint:mnd // wontfix
		return err
	}

	if err := writer.WriteUInt32(uint32(len(sortedFiles))); err != nil { //nolint:gosec // wontfix
		return err
	}

	if err := writer.WriteUInt32(0); err != nil {
		return err
	}

	pos, _ := writer.Position()

	if _, err := writer.Seek(int64(48*len(sortedFiles))+pos, 0); err != nil { //nolint:mnd // wontfix
		return err
	}

	for _, obj := range sortedFiles {
		fileEntry2 := readwrite.FileEntry{} //nolint:exhaustruct // wontfix
		text := strings.ReplaceAll(obj, directory, "")
		text = strings.ReplaceAll(text, "\\", "/")

		if text[0] == '/' {
			text = text[1:]
		}

		hashBytes := murmurhash3.NewX86_32(math.MaxUint32)
		hashBytes.Write(readwrite.Utf8ToUtf16(strings.ToLower(text)))
		hash := binary.LittleEndian.Uint32(hashBytes.Sum(nil))

		hashBytes2 := murmurhash3.NewX86_32(math.MaxUint32)
		hashBytes2.Write(readwrite.Utf8ToUtf16(strings.ToUpper(text)))
		hash2 := binary.LittleEndian.Uint32(hashBytes2.Sum(nil))

		reader, err := readwrite.NewReader(obj)
		if err != nil {
			return err
		}

		size, _ := reader.Size()
		array2 := make([]byte, size)

		if _, err := reader.Read(array2); err != nil {
			return err
		}

		fileEntry2.FileName = text
		pos, _ = writer.Position()
		fileEntry2.Offset = uint64(pos) //nolint:gosec // wontfix
		fileEntry2.UncompSize = uint64(len(array2))
		fileEntry2.FileNameLower = hash
		fileEntry2.FileNameUpper = hash2
		list = append(list, fileEntry2)

		if _, err := writer.Write(array2); err != nil {
			return err
		}

		data = append(data, readwrite.DataEntry{Hash: hash, FileName: text})
		data = append(data, readwrite.DataEntry{Hash: hash2, FileName: text})
	}

	if _, err := writer.SeekFromBeginning(16); err != nil { //nolint:mnd // wontfix
		return err
	}

	for _, item := range list {
		fmt.Printf("%s, %v, %v\n", item.FileName, item.FileNameLower, item.FileNameUpper)

		if err := writer.WriteUInt32(item.FileNameLower); err != nil {
			return err
		}

		if err := writer.WriteUInt32(item.FileNameUpper); err != nil {
			return err
		}

		if err := writer.WriteUInt64(item.Offset); err != nil {
			return err
		}

		if err := writer.WriteUInt64(item.UncompSize); err != nil {
			return err
		}

		if err := writer.WriteUInt64(item.UncompSize); err != nil {
			return err
		}

		if err := writer.WriteUInt64(0); err != nil {
			return err
		}

		if err := writer.WriteUInt32(0); err != nil {
			return err
		}

		if err := writer.WriteUInt32(0); err != nil {
			return err
		}
	}

	//nolint:nestif // wontfix
	if embed {
		if _, err := writer.SeekFromEnd(0); err != nil {
			return err
		}

		if err := WriteData(writer, data); err != nil {
			return err
		}
	} else {
		dataWriter, err := readwrite.NewWriter(outputFile+".data", false)
		if err != nil {
			return err
		}

		if err := WriteData(dataWriter, data); err != nil {
			panic(err)
		}

		if err := dataWriter.Close(); err != nil {
			return err
		}
	}

	if err := writer.Close(); err != nil {
		return err
	}

	return nil
}

//nolint:funlen,gocognit,gocyclo,cyclop // wontfix
func ExtractDirectory(source, outputPath string, embed bool) error {
	reader, err := readwrite.NewReader(source)

	var table []readwrite.DataEntry

	if embed {
		table, err = ReadData(reader)
	} else {
		dataReader, err := readwrite.NewReader(source + ".data")
		if err != nil {
			return err
		}

		table, err = ReadData(dataReader)
		if err != nil {
			return err
		}

		if err := dataReader.Close(); err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	if _, err := reader.Seek(0, 0); err != nil {
		return err
	}

	var unk0 uint32

	var unk1 uint32

	var unk2 uint32

	if unk0, err = reader.ReadUInt32(); err != nil {
		return err
	}

	if unk1, err = reader.ReadUInt32(); err != nil {
		return err
	}

	if unk2, err = reader.ReadUInt32(); err != nil {
		return err
	}

	if _, err := reader.ReadUInt32(); err != nil {
		return err
	}

	if unk0 != 1095454795 || unk1 != 4 {
		return errInvalidFileFormat
	}

	var list []readwrite.FileEntry

	for i := uint32(0); i < unk2; i++ { //nolint:intrange // wontfix
		fileEntry := readwrite.FileEntry{} //nolint:exhaustruct // wontfix

		fileEntry.FileNameLower, _ = reader.ReadUInt32()
		fileEntry.FileNameUpper, _ = reader.ReadUInt32()
		fileEntry.Offset, _ = reader.ReadUInt64()
		fileEntry.UncompSize, _ = reader.ReadUInt64()

		if _, err := reader.SeekFromCurrent(8); err != nil { //nolint:mnd // wontfix
			return err
		}

		if _, err := reader.SeekFromCurrent(8); err != nil { //nolint:mnd // wontfix
			return err
		}

		if _, err := reader.SeekFromCurrent(4); err != nil { //nolint:mnd // wontfix
			return err
		}

		if _, err := reader.SeekFromCurrent(4); err != nil { //nolint:mnd // wontfix
			return err
		}

		list = append(list, fileEntry)
	}

	for _, entry := range list {
		dataEntry := readwrite.FindByHash(table, entry.FileNameLower)
		if dataEntry == nil {
			simplelog.SharedLogger.Errorf("File entry not found")
			break
		}

		filePath := filepath.Join(outputPath, dataEntry.FileName)
		fileData := make([]byte, entry.UncompSize)

		if _, err := reader.Read(fileData); err != nil {
			return err
		}

		if len(fileData) > 1073741824 { //nolint:mnd // wontfix
			return errFileTooLarge
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		writer, err := readwrite.NewWriter(filePath, false)
		if err != nil {
			return err
		}

		if _, err := writer.Write(fileData); err != nil {
			return err
		}

		if err := writer.Close(); err != nil {
			return err
		}
	}

	if err := reader.Close(); err != nil {
		return err
	}

	return nil
}
