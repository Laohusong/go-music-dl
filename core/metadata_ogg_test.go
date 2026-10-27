package core

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dhowden/tag"
	"github.com/guohuiyuan/music-lib/model"
)

func TestEmbedOggMetadataPreservesAudioAndArtwork(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not available")
	}
	// A large picture exercises the metadata-file path instead of a potentially
	// oversized command-line argument (Linux limits a single argument to 128 KiB).
	img := image.NewNRGBA(image.Rect(0, 0, 256, 256))
	rng := rand.New(rand.NewSource(1))
	_, _ = rng.Read(img.Pix)
	for i := 3; i < len(img.Pix); i += 4 {
		img.Pix[i] = 255
	}
	var cover bytes.Buffer
	if err := png.Encode(&cover, img); err != nil {
		t.Fatal(err)
	}
	for _, codec := range []string{"libvorbis", "libopus"} {
		t.Run(codec, func(t *testing.T) {
			dir := t.TempDir()
			input := filepath.Join(dir, "input.ogg")
			cmd := exec.Command("ffmpeg", "-y", "-loglevel", "error", "-f", "lavfi", "-i", "sine=frequency=440:sample_rate=48000", "-t", "0.1", "-c:a", codec, "-metadata", "title=Old title", "-metadata", "genre=Jazz", input)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("generate Ogg: %v: %s", err, out)
			}
			audio, err := os.ReadFile(input)
			if err != nil {
				t.Fatal(err)
			}
			const raw = "[00:01.00]你[00:01.50]好[00:02.00]"
			const want = "[00:01.00]你好"
			embedded, err := EmbedSongMetadata(audio, &model.Song{Name: "新歌", Artist: "歌手", Album: "专辑", Ext: "ogg"}, raw, cover.Bytes(), "image/png")
			if err != nil {
				t.Fatal(err)
			}
			checkMetadata := func(data []byte) {
				t.Helper()
				metadata, err := tag.ReadFrom(bytes.NewReader(data))
				if err != nil {
					t.Fatal(err)
				}
				if metadata.Title() != "新歌" || metadata.Artist() != "歌手" || metadata.Album() != "专辑" || metadata.Genre() != "Jazz" || metadata.Lyrics() != want {
					t.Fatalf("unexpected Ogg metadata: %q/%q/%q/%q/%q", metadata.Title(), metadata.Artist(), metadata.Album(), metadata.Genre(), metadata.Lyrics())
				}
				pic := metadata.Picture()
				if pic == nil || pic.MIMEType != "image/png" || !bytes.Equal(pic.Data, cover.Bytes()) {
					t.Fatal("Ogg artwork missing or changed")
				}
			}
			checkMetadata(embedded)
			// Updating another field must preserve existing artwork and lyrics.
			embedded, err = EmbedSongMetadata(embedded, &model.Song{Ext: "ogg"}, "", nil, "")
			if err != nil {
				t.Fatal(err)
			}
			checkMetadata(embedded)
			output := filepath.Join(dir, "output.ogg")
			if err := os.WriteFile(output, embedded, 0644); err != nil {
				t.Fatal(err)
			}
			packetHashes := func(path string) []byte {
				t.Helper()
				out, err := exec.Command("ffprobe", "-v", "error", "-select_streams", "a", "-show_packets", "-show_data_hash", "sha256", "-show_entries", "packet=data_hash", "-of", "json", path).CombinedOutput()
				if err != nil {
					t.Fatalf("probe audio packets: %v: %s", err, out)
				}
				var result struct {
					Packets []struct {
						Hash string `json:"data_hash"`
					} `json:"packets"`
				}
				if err := json.Unmarshal(out, &result); err != nil {
					t.Fatal(err)
				}
				var hashes []string
				for _, packet := range result.Packets {
					hashes = append(hashes, packet.Hash)
				}
				if len(hashes) == 0 {
					t.Fatal("no compressed audio packets found")
				}
				return []byte(strings.Join(hashes, "\n"))
			}
			if before, after := packetHashes(input), packetHashes(output); !bytes.Equal(before, after) {
				t.Fatalf("compressed audio packets changed while embedding metadata:\nbefore=%s\nafter=%s", before, after)
			}
		})
	}
}
