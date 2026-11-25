package snapshots

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/flanksource/commons/properties"
	"github.com/klauspost/compress/zstd"
)

const (
	compressionThresholdProperty = "gavel.compress.threshold"
	compressionDefaultProperty   = "gavel.compress.default"
	gzipSuffix                   = ".gz"
	zstdSuffix                   = ".zst"
)

type snapshotCompression struct {
	threshold int64
	suffix    string
}

func compressionConfig() (snapshotCompression, error) {
	value := properties.String("64KB", compressionThresholdProperty)
	threshold, err := properties.ParseBytes(value)
	if err != nil || threshold < 0 {
		return snapshotCompression{}, fmt.Errorf("%s=%q: expected a non-negative byte size such as 64KB", compressionThresholdProperty, value)
	}
	config := snapshotCompression{threshold: threshold}
	format := properties.String("zstd", compressionDefaultProperty)
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "zstd", "zst":
		config.suffix = zstdSuffix
	case "gzip", "gz":
		config.suffix = gzipSuffix
	case "none":
	default:
		return snapshotCompression{}, fmt.Errorf("%s=%q: expected zstd, gzip, or none", compressionDefaultProperty, format)
	}
	return config, nil
}

// SnapshotStem removes the JSON and compression suffixes from a snapshot name.
// It returns an empty string for other kinds of files.
func SnapshotStem(name string) string {
	stem, ok := strings.CutSuffix(snapshotJSONName(name), ".json")
	if !ok {
		return ""
	}
	return stem
}

func snapshotJSONName(name string) string {
	for _, suffix := range [...]string{gzipSuffix, zstdSuffix} {
		if strings.HasSuffix(name, ".json"+suffix) {
			return strings.TrimSuffix(name, suffix)
		}
	}
	return name
}

func snapshotKeys(name string) [3]string {
	base := snapshotJSONName(name)
	return [3]string{base, base + zstdSuffix, base + gzipSuffix}
}

func openSnapshot(name string, open func(string) (io.ReadCloser, error), missing error) (string, io.ReadCloser, error) {
	reader, err := open(name)
	if !errors.Is(err, missing) || SnapshotStem(name) == "" {
		return name, reader, err
	}
	for _, candidate := range snapshotKeys(name) {
		if candidate == name {
			continue
		}
		if reader, candidateErr := open(candidate); !errors.Is(candidateErr, missing) {
			return candidate, reader, candidateErr
		}
	}
	return name, nil, err
}

// OpenFile opens the decoded JSON stream of a plain, gzip or zstd snapshot.
func OpenFile(path string) (io.ReadCloser, error) {
	name, reader, err := openSnapshot(path, func(path string) (io.ReadCloser, error) {
		return os.Open(path)
	}, os.ErrNotExist)
	if err != nil {
		return nil, err
	}
	return decodeReader(name, reader)
}

// ReadFile reads a snapshot as JSON, decoding compression when present.
func ReadFile(path string) ([]byte, error) {
	reader, err := OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

type snapshotReader struct {
	io.Reader
	close func() error
}

func (r snapshotReader) Close() error { return r.close() }

func decodeReader(name string, source io.ReadCloser) (io.ReadCloser, error) {
	var err error
	switch {
	case strings.HasSuffix(name, gzipSuffix):
		var reader *gzip.Reader
		reader, err = gzip.NewReader(source)
		if err == nil {
			return snapshotReader{reader, func() error { return errors.Join(reader.Close(), source.Close()) }}, nil
		}
	case strings.HasSuffix(name, zstdSuffix):
		var reader *zstd.Decoder
		reader, err = zstd.NewReader(source, zstd.WithDecoderConcurrency(1))
		if err == nil {
			return snapshotReader{reader, func() error { reader.Close(); return source.Close() }}, nil
		}
	default:
		return source, nil
	}
	return nil, errors.Join(fmt.Errorf("open compressed snapshot %s: %w", name, err), source.Close())
}

func compressSnapshot(dst io.Writer, source io.Reader, suffix string) error {
	var writer io.WriteCloser
	switch suffix {
	case gzipSuffix:
		writer = gzip.NewWriter(dst)
	case zstdSuffix:
		encoder, err := zstd.NewWriter(dst, zstd.WithEncoderConcurrency(1))
		if err != nil {
			return err
		}
		writer = encoder
	default:
		return fmt.Errorf("unsupported snapshot compression suffix %q", suffix)
	}
	_, err := io.Copy(writer, source)
	return errors.Join(err, writer.Close())
}
