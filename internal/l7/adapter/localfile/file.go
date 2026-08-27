// Package localfile provides bounded, link-resistant local persistence.
package localfile

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
)

var errLimit = errors.New("file exceeds size limit")

func Read(path string, limit int64) ([]byte, error) {
	if limit < 1 {
		return nil, errors.New("size limit must be positive")
	}
	before, err := regularFileInfo(path)
	if err != nil {
		return nil, err
	}
	if before.Size() > limit {
		return nil, errLimit
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open regular file: %w", err)
	}
	defer file.Close()
	after, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("inspect open file: %w", err)
	}
	if !after.Mode().IsRegular() || !os.SameFile(before, after) {
		return nil, errors.New("file identity changed while opening")
	}
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read regular file: %w", err)
	}
	if int64(len(data)) > limit {
		return nil, errLimit
	}
	return data, nil
}

func AtomicCreate(path string, data []byte, mode os.FileMode) error {
	return atomicWrite(path, data, mode, false)
}

func AtomicReplace(path string, data []byte, mode os.FileMode) error {
	return atomicWrite(path, data, mode, true)
}

func atomicWrite(path string, data []byte, mode os.FileMode, replace bool) error {
	if !filepath.IsAbs(path) {
		return errors.New("atomic path must be absolute")
	}
	directory := filepath.Dir(filepath.Clean(path))
	root, directoryInfo, err := openAnchoredDirectory(directory)
	if err != nil {
		return err
	}
	defer root.Close()
	name := filepath.Base(path)
	if info, err := root.Lstat(name); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return errors.New("destination is not a regular file")
		}
		if !replace {
			return os.ErrExist
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect destination: %w", err)
	}

	temporary, temporaryName, err := createAnchoredTemp(root)
	if err != nil {
		return fmt.Errorf("create atomic temporary file: %w", err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = temporary.Close()
		}
		if temporaryName != "" {
			_ = root.Remove(temporaryName)
		}
	}()
	if err := temporary.Chmod(mode.Perm()); err != nil {
		return fmt.Errorf("set atomic file mode: %w", err)
	}
	if err := writeAll(temporary, data); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("sync atomic file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close atomic file: %w", err)
	}
	closed = true
	if err := revalidateAnchoredDirectory(root, directory, directoryInfo); err != nil {
		return err
	}

	if replace {
		if err := root.Rename(temporaryName, name); err != nil {
			return fmt.Errorf("replace atomic file: %w", err)
		}
	} else {
		if err := root.Link(temporaryName, name); err != nil {
			return fmt.Errorf("create atomic file: %w", err)
		}
		if err := root.Remove(temporaryName); err != nil {
			return fmt.Errorf("unlink atomic temporary file: %w", err)
		}
	}
	temporaryName = ""
	if err := syncAnchoredDirectory(root); err != nil {
		return err
	}
	return revalidateAnchoredDirectory(root, directory, directoryInfo)
}

func EnsureDirectory(path string, mode os.FileMode) error {
	if !filepath.IsAbs(path) {
		return errors.New("directory path must be absolute")
	}
	clean := filepath.Clean(path)
	var missing []string
	cursor := clean
	for {
		info, err := os.Lstat(cursor)
		if err == nil {
			if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
				return fmt.Errorf("unsafe directory component: %s", cursor)
			}
			break
		}
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("inspect directory: %w", err)
		}
		missing = append(missing, cursor)
		parent := filepath.Dir(cursor)
		if parent == cursor {
			return errors.New("no existing directory ancestor")
		}
		cursor = parent
	}
	for index := len(missing) - 1; index >= 0; index-- {
		if err := os.Mkdir(missing[index], mode.Perm()); err != nil && !errors.Is(err, os.ErrExist) {
			return fmt.Errorf("create directory: %w", err)
		}
		if err := ValidateDirectory(missing[index]); err != nil {
			return err
		}
	}
	return nil
}

func ValidateDirectory(path string) error {
	if !filepath.IsAbs(path) {
		return errors.New("directory path must be absolute")
	}
	clean := filepath.Clean(path)
	volume := filepath.VolumeName(clean)
	remainder := clean[len(volume):]
	cursor := volume + string(filepath.Separator)
	for _, component := range splitPath(remainder) {
		cursor = filepath.Join(cursor, component)
		info, err := os.Lstat(cursor)
		if err != nil {
			return fmt.Errorf("inspect directory component: %w", err)
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("unsafe directory component: %s", cursor)
		}
	}
	return nil
}

func DecodeJSON(data []byte, target any) error {
	if target == nil || reflect.TypeOf(target).Kind() != reflect.Pointer || reflect.ValueOf(target).IsNil() {
		return errors.New("JSON target is nil")
	}
	validator := json.NewDecoder(bytes.NewReader(data))
	if err := validateJSONShape(validator, reflect.TypeOf(target).Elem()); err != nil {
		return err
	}
	if _, err := validator.Token(); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("JSON contains trailing data")
		}
		return fmt.Errorf("read JSON trailer: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode strict JSON: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("JSON contains trailing value")
		}
		return fmt.Errorf("decode JSON trailer: %w", err)
	}
	return nil
}

func validateJSONShape(decoder *json.Decoder, expected reflect.Type) error {
	for expected.Kind() == reflect.Pointer {
		expected = expected.Elem()
	}
	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("read JSON value: %w", err)
	}
	if token == nil {
		return errors.New("JSON null is unsupported in strict records")
	}
	switch expected.Kind() {
	case reflect.Struct:
		if token != json.Delim('{') {
			return errors.New("JSON value is not the required object")
		}
		fields := exactJSONFields(expected)
		seen := make(map[string]bool)
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return fmt.Errorf("read JSON object key: %w", err)
			}
			key, ok := keyToken.(string)
			if !ok {
				return errors.New("JSON object key is not a string")
			}
			fieldType, ok := fields[key]
			if !ok {
				return fmt.Errorf("unknown or non-canonical JSON field %q", key)
			}
			if seen[key] {
				return fmt.Errorf("duplicate JSON field %q", key)
			}
			seen[key] = true
			if err := validateJSONShape(decoder, fieldType); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim('}') {
			return errors.New("unterminated JSON object")
		}
		return nil
	case reflect.Slice, reflect.Array:
		if token != json.Delim('[') {
			return errors.New("JSON value is not the required array")
		}
		for decoder.More() {
			if err := validateJSONShape(decoder, expected.Elem()); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim(']') {
			return errors.New("unterminated JSON array")
		}
		return nil
	case reflect.Map, reflect.Interface:
		return validateDynamicJSONToken(decoder, token, expected)
	default:
		if _, ok := token.(json.Delim); ok {
			return errors.New("JSON container does not match the required scalar")
		}
		return nil
	}
}

func validateDynamicJSONToken(decoder *json.Decoder, token json.Token, expected reflect.Type) error {
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := make(map[string]bool)
		valueType := reflect.TypeOf((*any)(nil)).Elem()
		if expected.Kind() == reflect.Map {
			valueType = expected.Elem()
		}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok || seen[key] {
				return errors.New("JSON object key is invalid or duplicated")
			}
			seen[key] = true
			if err := validateJSONShape(decoder, valueType); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim('}') {
			return errors.New("unterminated JSON object")
		}
	case '[':
		valueType := reflect.TypeOf((*any)(nil)).Elem()
		for decoder.More() {
			if err := validateJSONShape(decoder, valueType); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim(']') {
			return errors.New("unterminated JSON array")
		}
	default:
		return errors.New("unexpected JSON delimiter")
	}
	return nil
}

func exactJSONFields(structType reflect.Type) map[string]reflect.Type {
	fields := make(map[string]reflect.Type)
	for index := 0; index < structType.NumField(); index++ {
		field := structType.Field(index)
		if field.PkgPath != "" {
			continue
		}
		name := field.Name
		if tag := field.Tag.Get("json"); tag != "" {
			if comma := strings.IndexByte(tag, ','); comma >= 0 {
				tag = tag[:comma]
			}
			if tag == "-" {
				continue
			}
			if tag != "" {
				name = tag
			}
		}
		fields[name] = field.Type
	}
	return fields
}

func EncodeJSON(value any) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode JSON: %w", err)
	}
	return append(data, '\n'), nil
}

type Lock struct {
	file *os.File
	root *os.Root
}

func AcquireLock(path string) (*Lock, error) {
	if !filepath.IsAbs(path) {
		return nil, errors.New("lock path must be absolute")
	}
	directory := filepath.Dir(path)
	root, directoryInfo, err := openAnchoredDirectory(directory)
	if err != nil {
		return nil, err
	}
	file, err := root.OpenFile(filepath.Base(path), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		_ = root.Close()
		return nil, fmt.Errorf("open repository lock: %w", err)
	}
	fail := func(cause error) (*Lock, error) {
		_ = file.Close()
		_ = root.Close()
		return nil, cause
	}
	opened, err := file.Stat()
	if err != nil || !opened.Mode().IsRegular() {
		return fail(errors.New("repository lock is not a regular file"))
	}
	linked, err := root.Lstat(filepath.Base(path))
	if err != nil || linked.Mode()&os.ModeSymlink != 0 || !os.SameFile(opened, linked) {
		return fail(errors.New("repository lock identity is unsafe"))
	}
	if err := revalidateAnchoredDirectory(root, directory, directoryInfo); err != nil {
		return fail(err)
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return fail(errors.New("another Level 7 mutation is active"))
	}
	return &Lock{file: file, root: root}, nil
}

func (lock *Lock) Close() error {
	if lock == nil || lock.file == nil {
		return nil
	}
	file := lock.file
	root := lock.root
	lock.file = nil
	lock.root = nil
	unlockErr := syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	closeErr := file.Close()
	if unlockErr != nil {
		return fmt.Errorf("unlock repository: %w", unlockErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close repository lock: %w", closeErr)
	}
	if root != nil {
		if err := root.Close(); err != nil {
			return fmt.Errorf("close repository lock root: %w", err)
		}
	}
	return nil
}

func regularFileInfo(path string) (os.FileInfo, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("inspect regular file: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, errors.New("path is not a regular file")
	}
	return info, nil
}

func writeAll(file *os.File, data []byte) error {
	for len(data) > 0 {
		written, err := file.Write(data)
		if err != nil {
			return fmt.Errorf("write atomic file: %w", err)
		}
		if written == 0 {
			return io.ErrShortWrite
		}
		data = data[written:]
	}
	return nil
}

func syncAnchoredDirectory(root *os.Root) error {
	directory, err := root.Open(".")
	if err != nil {
		return fmt.Errorf("open directory for sync: %w", err)
	}
	defer directory.Close()
	if err := directory.Sync(); err != nil {
		return fmt.Errorf("sync directory: %w", err)
	}
	return nil
}

func openAnchoredDirectory(path string) (*os.Root, os.FileInfo, error) {
	if err := ValidateDirectory(path); err != nil {
		return nil, nil, err
	}
	before, err := os.Lstat(path)
	if err != nil || before.Mode()&os.ModeSymlink != 0 || !before.IsDir() {
		return nil, nil, errors.New("atomic parent is not a safe directory")
	}
	root, err := os.OpenRoot(path)
	if err != nil {
		return nil, nil, fmt.Errorf("anchor atomic parent: %w", err)
	}
	opened, err := root.Stat(".")
	if err != nil || !opened.IsDir() || !os.SameFile(before, opened) {
		_ = root.Close()
		return nil, nil, errors.New("atomic parent identity changed while opening")
	}
	if err := revalidateAnchoredDirectory(root, path, opened); err != nil {
		_ = root.Close()
		return nil, nil, err
	}
	return root, opened, nil
}

func revalidateAnchoredDirectory(root *os.Root, path string, opened os.FileInfo) error {
	linked, err := os.Lstat(path)
	if err != nil || linked.Mode()&os.ModeSymlink != 0 || !linked.IsDir() || !os.SameFile(opened, linked) {
		return errors.New("atomic parent identity changed")
	}
	current, err := root.Stat(".")
	if err != nil || !current.IsDir() || !os.SameFile(opened, current) {
		return errors.New("anchored atomic parent changed")
	}
	return nil
}

func createAnchoredTemp(root *os.Root) (*os.File, string, error) {
	for attempt := 0; attempt < 100; attempt++ {
		var random [16]byte
		if _, err := rand.Read(random[:]); err != nil {
			return nil, "", err
		}
		name := ".l7-write-" + hex.EncodeToString(random[:])
		file, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
		if err == nil {
			return file, name, nil
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, "", err
		}
	}
	return nil, "", errors.New("cannot allocate a unique atomic temporary file")
}

func splitPath(value string) []string {
	var components []string
	start := 0
	for index := 0; index <= len(value); index++ {
		if index == len(value) || os.IsPathSeparator(value[index]) {
			if index > start {
				components = append(components, value[start:index])
			}
			start = index + 1
		}
	}
	return components
}
