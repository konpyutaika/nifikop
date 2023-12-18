package k8sutil

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"io"
	"net/http"

	meta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

var metadataAccessor = meta.NewAccessor()

// SetAnnotation set an annotation.
func SetAnnotation(annotationKey string, obj runtime.Object, value []byte) error {
	if len(value) < 1 {
		return nil
	}

	annots, err := metadataAccessor.Annotations(obj)
	if err != nil {
		return err
	}

	if annots == nil {
		annots = map[string]string{}
	}

	annots[annotationKey], err = zipAndBase64EncodeAnnotation(value)
	if err != nil {
		return err
	}
	return metadataAccessor.SetAnnotations(obj, annots)
}

// GetAnnotation get a annotation.
func GetAnnotation(annotationKey string, obj runtime.Object) ([]byte, error) {
	annots, err := metadataAccessor.Annotations(obj)
	if err != nil {
		return nil, err
	}

	if annots == nil {
		return nil, nil
	}

	original, ok := annots[annotationKey]
	if !ok {
		return nil, nil
	}

	// Try to base64 decode, and fallback to non-base64 encoded content for backwards compatibility.
	if decoded, err := base64.StdEncoding.DecodeString(original); err == nil {
		if http.DetectContentType(decoded) == "application/zip" {
			return unZipAnnotation(decoded)
		}
		return decoded, nil
	}

	return []byte(original), nil
}

func zipAndBase64EncodeAnnotation(original []byte) (string, error) {
	// Create a buffer to write our archive to.
	buf := new(bytes.Buffer)

	// Create a new zip archive.
	w := zip.NewWriter(buf)

	f, err := w.Create("original")
	if err != nil {
		return "", err
	}
	_, err = f.Write(original)
	if err != nil {
		return "", err
	}

	// Make sure to check the error on Close.
	err = w.Close()
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func unZipAnnotation(original []byte) ([]byte, error) {
	annotation, err := io.ReadAll(bytes.NewReader(original))
	if err != nil {
		return nil, err
	}

	zipReader, err := zip.NewReader(bytes.NewReader(annotation), int64(len(annotation)))
	if err != nil {
		return nil, err
	}

	// Read the file from zip archive
	zipFile := zipReader.File[0]
	unzippedFileBytes, err := readZipFile(zipFile)
	if err != nil {
		return nil, err
	}

	return unzippedFileBytes, nil
}

func readZipFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}
