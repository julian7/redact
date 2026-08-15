package ext

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/julian7/redact/repo"
)

const (
	ConfigFilename = "config.json"
	DotconfDir     = ".config"
	DotconfFile    = "redact.json"
)

type Config struct {
	Exts map[string]Ext
	repo *repo.Repo
}

func configFile(r *repo.Repo) (string, bool) {
	filename := ""

	for _, candidate := range []string{
		r.Workdir.Join(repo.DefaultKeyExchangeDir, ConfigFilename),
		r.Workdir.Join(DotconfDir, DotconfFile),
	} {
		stat, err := r.Workdir.Stat(candidate)
		if err == nil && stat.Mode().Perm().IsRegular() {
			filename = candidate

			break
		}
	}

	if filename == "" {
		return "", false
	}

	return filename, true
}

func Load(r *repo.Repo) (*Config, error) {
	conf := &Config{
		Exts: map[string]Ext{},
		repo: r,
	}

	configFilename, ok := configFile(r)
	if !ok {
		return conf, nil
	}

	data, err := r.Workdir.Open(configFilename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return conf, nil
		}

		return nil, err
	}

	decoder := json.NewDecoder(data)

	exts := map[string]Ext{}
	if err = decoder.Decode(&exts); err != nil {
		return conf, err
	}

	_ = data.Close()

	for name, item := range exts {
		if err := conf.AddExt(name, item); err != nil {
			return nil, err
		}
	}

	return conf, nil
}

func (conf *Config) AddExt(name string, ext Ext) error {
	if _, ok := conf.Exts[name]; ok {
		return ErrExtAlreadyExists
	}

	ext.name = name
	ext.repo = conf.repo
	conf.Exts[name] = ext

	return nil
}

func (conf *Config) UpdateExt(name string, ext Ext) error {
	if _, ok := conf.Exts[name]; !ok {
		return ErrExtNotFound
	}

	ext.name = name
	ext.repo = conf.repo
	conf.Exts[name] = ext

	return nil
}

func (conf *Config) DelExt(name string) {
	delete(conf.Exts, name)
}

func (conf *Config) Ext(name string) (Ext, bool) {
	item, ok := conf.Exts[name]

	return item, ok
}

func (conf *Config) Save() error {
	configFilename, ok := configFile(conf.repo)
	if !ok {
		configFilename = conf.repo.Workdir.Join(repo.DefaultKeyExchangeDir, ConfigFilename)
	}

	fd, err := conf.repo.Workdir.OpenFile(configFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	encoder := json.NewEncoder(fd)
	if err = encoder.Encode(&conf.Exts); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	_ = fd.Close()

	return nil
}

func (conf *Config) SaveKey(data []byte) error {
	var err error
	for _, ext := range conf.Exts {
		err = ext.SaveKey(data)
		if err != nil {
			break
		}
	}

	return err
}

func (conf *Config) List() {
	for _, ext := range conf.Exts {
		_ = ext.List()
	}
}
