package lib

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var ErrNoEditorFound = errors.New("no editor found")

func CreateTempFile(fileName string) (*os.File, error) {
	date := time.Now().Format("01-02-2006_0204")
	path, err := os.MkdirTemp("", date)
	if err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("failed to create tempdir: %w", err)
	}
	f, err := os.Create(filepath.Join(path, fileName))
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	return f, nil
}

func Hash(fname string) ([]byte, error) {
	h := sha256.New()
	f, err := os.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("erorr creating file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(h, f); err != nil {
		return nil, fmt.Errorf("error computing file hash: %w", err)
	}
	return h.Sum(make([]byte, 0)), nil
}

func OpenEditor(fpath string) error {
	editor, err := getEditor()
	if err != nil {
		return err
	}
	cmd := exec.Command(editor, fpath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error opening editor: %w", err)
	}
	return nil
}

func getEditor() (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		for _, ed := range []string{"vim", "vi", "nano"} {
			editor, err := exec.LookPath(ed)
			if err == nil {
				return editor, nil
			}
		}
	}
	if editor == "" {
		return "", ErrNoEditorFound
	}
	return editor, nil
}
