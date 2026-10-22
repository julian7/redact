package ext

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"slices"
	"sort"

	"github.com/julian7/redact/repo"
)

const (
	orderKey       = "_order_"
	ConfigFilename = "config.json"
	DotconfDir     = ".config"
	DotconfFile    = "redact.json"
)

type Config struct {
	Exts  map[string]Ext
	Order []string
	repo  *repo.Repo
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

	defer data.Close()

	decoder := json.NewDecoder(data)

	exts := map[string]json.RawMessage{}
	if err = decoder.Decode(&exts); err != nil {
		return conf, err
	}

	for name, raw := range exts {
		if name == orderKey {
			if len(raw) == 0 || raw[0] != '[' {
				return nil, fmt.Errorf("extension _order_ must be an array of names")
			}

			if err := json.Unmarshal(raw, &conf.Order); err != nil {
				return nil, fmt.Errorf("extension _order_: %w", err)
			}

			continue
		}

		var item Ext
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, fmt.Errorf("extension %s: %w", name, err)
		}

		if err := conf.AddExt(name, item); err != nil {
			return nil, err
		}
	}

	return conf, nil
}

func (conf *Config) AddExt(name string, ext Ext) error {
	if name == orderKey {
		return ErrReservedName
	}

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

	if err := conf.repo.Workdir.MkdirAll(conf.repo.Workdir.Join(configFilename, ".."), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	fd, err := conf.repo.Workdir.OpenFile(configFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	defer fd.Close()

	entries := make(map[string]any, len(conf.Exts)+1)
	for name, item := range conf.Exts {
		entries[name] = item
	}

	if conf.Order != nil {
		entries[orderKey] = conf.Order
	}

	encoder := json.NewEncoder(fd)
	if err = encoder.Encode(entries); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	return fd.Close()
}

// Names returns configured extensions in preference order, followed by unlisted
// extensions alphabetically. Stale names and duplicates in Order are ignored.
func (conf *Config) Names() []string {
	names := make([]string, 0, len(conf.Exts))
	for _, name := range conf.Order {
		if _, ok := conf.Exts[name]; ok && !slices.Contains(names, name) {
			names = append(names, name)
		}
	}

	remaining := make([]string, 0, len(conf.Exts))
	for name := range conf.Exts {
		if !slices.Contains(names, name) {
			remaining = append(remaining, name)
		}
	}

	sort.Strings(remaining)

	return append(names, remaining...)
}

// Resolve creates an invocation without modifying configured defaults. Explicit
// selection can resolve redact-ext-NAME even when NAME is not configured.
func (conf *Config) Resolve(name, command string, options map[string]string) (Ext, error) {
	if name == "" || name == orderKey {
		return Ext{}, ErrExtNotFound
	}

	item := conf.Exts[name]
	item.name, item.repo = name, conf.repo

	item.options = maps.Clone(options)

	if command != "" {
		item.Command = command
	}

	return item, nil
}

// Select applies the same invocation settings to all configured destinations
// unless one extension is explicitly selected.
func (conf *Config) Select(name, command string, options map[string]string) ([]Ext, error) {
	if name == "" && command != "" {
		return nil, ErrCmdRequiresExt
	}

	names := conf.Names()
	if name != "" {
		names = []string{name}
	}

	items := make([]Ext, 0, len(names))
	for _, name := range names {
		item, err := conf.Resolve(name, command, options)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

// RunAll continues after success or not-applicable outcomes. A missing-data or
// failure outcome stops dispatch immediately, retaining earlier changes.
func RunAll(items []Ext, operation string, data []byte) error {
	var skipped []error

	successful := false

	for _, item := range items {
		err := item.Run(operation, data)
		switch ExitCode(err) {
		case Success:
			successful = true
		case NotApplicable:
			skipped = append(skipped, err)
		default:
			return err
		}
	}

	if !successful {
		return errors.Join(skipped...)
	}

	return nil
}

func (conf *Config) SaveKey(data []byte) error {
	items, err := conf.Select("", "", nil)
	if err != nil {
		return err
	}

	return RunAll(items, "put", data)
}

func (conf *Config) List() error {
	var skipped []error

	successful := false

	for _, name := range conf.Names() {
		item := conf.Exts[name]

		err := item.List()
		switch ExitCode(err) {
		case Success:
			successful = true
		case NotApplicable:
			skipped = append(skipped, err)
		default:
			return err
		}
	}

	if !successful {
		return errors.Join(skipped...)
	}

	return nil
}
