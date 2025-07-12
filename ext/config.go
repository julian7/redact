package ext

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/julian7/redact/repo"
)

const ConfigFilename = "config.json"

type Config struct {
	Exts map[string]Ext
	repo *repo.Repo
}

func Load(r *repo.Repo) (*Config, error) {
	conf := &Config{
		Exts: map[string]Ext{},
		repo: r,
	}
	kxdir := r.ExchangeDir()

	data, err := r.Workdir.Open(r.Workdir.Join(kxdir, ConfigFilename))
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
	kxdir := conf.repo.ExchangeDir()

	fd, err := conf.repo.Workdir.OpenFile(
		conf.repo.Workdir.Join(kxdir, ConfigFilename),
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0644,
	)
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
