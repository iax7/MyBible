// Package mtp copies Modules from a phone into the Module directory.
//
// It reaches the phone over MTP by driving KIO's kioclient, the same stack
// Dolphin uses, so a phone in file transfer mode is the only prerequisite.
package mtp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// appDirs are the folders the MyBible app keeps its data in, relative to a
// storage: the shared folder first, then the app's own folder Android 11+
// falls back to.
var appDirs = []string{
	"MyBible",
	path.Join("Android", "data", "ua.mybible", "files", "MyBible"),
}

// CopyModules copies every Module and the app's user folder from the phone into
// dest, which is created when missing. Existing files are overwritten.
func CopyModules(dest string, out io.Writer) error {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}

	devices, err := entries("mtp:/")
	if err != nil {
		return fmt.Errorf("list MTP devices: %w", err)
	}
	if len(devices) == 0 {
		return errors.New("no MTP device: connect the phone and choose file transfer")
	}
	if len(devices) > 1 {
		return fmt.Errorf("several MTP devices (%s): leave only the phone connected", strings.Join(devices, ", "))
	}
	phone := devices[0]

	storages, err := entries("mtp:/" + phone)
	if err != nil {
		return fmt.Errorf("list the storages of %s: %w", phone, err)
	}
	var sources []string
	for _, storage := range storages {
		for _, appDir := range appDirs {
			url := "mtp:/" + phone + "/" + storage + "/" + appDir
			// A phone that hides Android/data over MTP simply has no such
			// folder; the shared one still works.
			if _, err := entries(url); err == nil {
				sources = append(sources, url)
			}
		}
	}
	if len(sources) == 0 {
		return fmt.Errorf("no MyBible folder on %s: is the app installed and its Modules downloaded?", phone)
	}

	copied := 0
	var failed []string
	for _, source := range sources {
		names, err := entries(source)
		if err != nil {
			return fmt.Errorf("read %s: %w", source, err)
		}
		for _, name := range names {
			if skipped(name) {
				continue
			}
			if name == "user" {
				n, errs := copyTree(source+"/"+name, filepath.Join(dest, name), name, out)
				copied += n
				failed = append(failed, errs...)
				continue
			}
			fmt.Fprintln(out, name)
			if err := copyItem(source+"/"+name, filepath.Join(dest, name)); err != nil {
				failed = append(failed, name+": "+err.Error())
				continue
			}
			copied++
		}
	}
	fmt.Fprintf(out, "%d files copied to %s\n", copied, dest)
	if len(failed) > 0 {
		return fmt.Errorf("could not copy some files:\n  %s", strings.Join(failed, "\n  "))
	}
	return nil
}

// copyTree copies a folder entry by entry: KIO nests a folder inside an
// existing destination folder instead of merging into it.
func copyTree(src, dest, display string, out io.Writer) (int, []string) {
	names, err := entries(src)
	if err != nil {
		return 0, []string{display + ": " + err.Error()}
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return 0, []string{display + ": " + err.Error()}
	}
	copied := 0
	var failed []string
	for _, name := range names {
		if skipped(name) {
			continue
		}
		srcChild := src + "/" + name
		destChild := filepath.Join(dest, name)
		displayChild := display + "/" + name
		if _, err := entries(srcChild); err == nil {
			n, errs := copyTree(srcChild, destChild, displayChild, out)
			copied += n
			failed = append(failed, errs...)
			continue
		}
		fmt.Fprintln(out, displayChild)
		if err := copyItem(srcChild, destChild); err != nil {
			failed = append(failed, displayChild+": "+err.Error())
			continue
		}
		copied++
	}
	return copied, failed
}

// skipped reports the files that are not worth copying: the app's caches,
// search indexes, journals, an unfinished download, and the reference data the
// app ships for itself, which the reader would mistake for a Bible.
func skipped(name string) bool {
	switch {
	case strings.HasPrefix(name, "-"):
		return true
	case strings.HasSuffix(name, ".search"),
		strings.HasSuffix(name, "-journal"),
		strings.HasSuffix(name, ".new"):
		return true
	case name == "ReferenceData.SQLite3", name == "persisted_registry.json":
		return true
	}
	return false
}

// entries lists one MTP folder. kioclient prints one name per line, blank lines
// between them and a final "." for the folder itself.
func entries(url string) ([]string, error) {
	out, err := kioclient("ls", url)
	if err != nil {
		return nil, err
	}
	return parseListing(out), nil
}

func parseListing(out string) []string {
	var names []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "." {
			continue
		}
		names = append(names, line)
	}
	return names
}

// copyItem copies through a ".part" name so a failed transfer leaves the
// previous backup intact instead of truncating it.
func copyItem(src, dest string) error {
	part := dest + ".part"
	if _, err := kioclient("--overwrite", "cp", src, part); err != nil {
		_ = os.Remove(part)
		return err
	}
	if err := os.Rename(part, dest); err != nil {
		_ = os.Remove(part)
		return err
	}
	return nil
}

func kioclient(args ...string) (string, error) {
	bin, err := exec.LookPath("kioclient5")
	if err != nil {
		bin, err = exec.LookPath("kioclient")
	}
	if err != nil {
		return "", errors.New("kioclient5 not found: install KDE's kio, which provides MTP access")
	}
	cmd := exec.Command(bin, append([]string{"--noninteractive"}, args...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return "", errors.New(msg)
		}
		return "", err
	}
	return string(out), nil
}
