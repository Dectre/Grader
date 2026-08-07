package manager

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func SavePDFToDir(src io.Reader, pdfDir, name string) error {
	os.MkdirAll(pdfDir, 0755)
	dst, err := os.Create(filepath.Join(pdfDir, name))
	if err != nil {
		return err
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

func ImportZipToPDFs(zipPath, pdfDir string) (int, error) {
	tmp, err := os.MkdirTemp("", "grader-zip-")
	if err != nil {
		return 0, err
	}
	defer os.RemoveAll(tmp)
	if err := extractZip(zipPath, tmp); err != nil {
		return 0, err
	}
	count := 0
	err = filepath.Walk(tmp, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".pdf") {
			return nil
		}
		rel, _ := filepath.Rel(tmp, path)
		parent := filepath.Dir(rel)
		name := info.Name()
		if parent != "." {
			name = filepath.Base(parent) + ".pdf"
		}
		if err := copyFile(path, filepath.Join(pdfDir, name)); err != nil {
			return err
		}
		count++
		return nil
	})
	return count, err
}

func extractZip(src, dst string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		target := filepath.Join(dst, f.Name)
		if !strings.HasPrefix(target, filepath.Clean(dst)+string(os.PathSeparator)) {
			continue
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(target, 0755)
			continue
		}
		os.MkdirAll(filepath.Dir(target), 0755)
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(target)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	os.MkdirAll(filepath.Dir(dst), 0755)
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}