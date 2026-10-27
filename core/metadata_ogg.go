package core

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

// createOggCoverMetadataFile writes a base64-encoded FLAC picture block as a
// Vorbis comment. A file avoids the OS argument-length limit for large artwork.
func createOggCoverMetadataFile(data []byte, mime string) (string, error) {
	var picture bytes.Buffer
	writeUint32 := func(v uint32) { _ = binary.Write(&picture, binary.BigEndian, v) }
	writeUint32(3) // front cover
	writeUint32(uint32(len(mime)))
	picture.WriteString(mime)
	writeUint32(0) // empty description
	var width, height uint32
	if config, _, err := image.DecodeConfig(bytes.NewReader(data)); err == nil {
		width, height = uint32(config.Width), uint32(config.Height)
	}
	writeUint32(width)
	writeUint32(height)
	writeUint32(0) // unspecified color depth
	writeUint32(0) // unspecified palette size
	writeUint32(uint32(len(data)))
	picture.Write(data)

	file, err := os.CreateTemp("", "gomusicdl-ogg-cover-*.ffmetadata")
	if err != nil {
		return "", err
	}
	path := file.Name()
	_, writeErr := file.WriteString(";FFMETADATA1\nMETADATA_BLOCK_PICTURE=" + base64.StdEncoding.EncodeToString(picture.Bytes()) + "\n")
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		_ = os.Remove(path)
		if writeErr != nil {
			return "", writeErr
		}
		return "", closeErr
	}
	return path, nil
}
