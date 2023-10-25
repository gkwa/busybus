package busybus

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	mymazda "github.com/taylormonacelli/forestfish/mymazda"
)

func Main() int {
	slog.Debug("busybus", "test", true)

	return 0
}

type CacheConfig struct {
	CachePath     string
	CacheLifetime time.Duration
}

func NewConfig(cachePath string, cacheLifetime time.Duration) (*CacheConfig, error) {
	dir := filepath.Dir(cachePath)
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return nil, err
	}

	c := CacheConfig{
		CachePath:     cachePath,
		CacheLifetime: cacheLifetime,
	}

	return &c, nil
}

func DecodeFromCache(cachePath string, target interface{}) error {
	if !mymazda.FileExists(cachePath) {
		return fmt.Errorf("cache file does not exist")
	}

	byteSlice, err := os.ReadFile(cachePath)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	buffer.Write(byteSlice)

	dec := gob.NewDecoder(&buffer)
	err = dec.Decode(target)

	if err != nil {
		return err
	}

	return nil
}

func SaveToCache(cachePath string, data interface{}) error {
	var buffer bytes.Buffer
	gob.Register(data)

	enc := gob.NewEncoder(&buffer)
	err := enc.Encode(data)
	if err != nil {
		return err
	}

	file, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(buffer.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (c *CacheConfig) RemoveExpiredCache() error {
	if !mymazda.FileExists(c.CachePath) {
		return nil
	}

	fileInfo, err := os.Stat(c.CachePath)
	if err != nil {
		return err
	}

	age := time.Since(fileInfo.ModTime()).Truncate(time.Second)
	expires := time.Until(fileInfo.ModTime().Add(c.CacheLifetime)).Truncate(time.Second)

	if age > c.CacheLifetime {
		slog.Debug("cache is old", "age", age, "path", c.CachePath)
		defer os.Remove(c.CachePath)
	} else {
		slog.Debug("cache stats", "age", age, "expires", expires, "path", c.CachePath)
	}

	return nil
}
