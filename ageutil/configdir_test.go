package ageutil

import (
	"testing"
)

func TestXDGConfigDir(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		home string
		want string
	}{
		{
			name: "XDG_CONFIG_HOME names the directory",
			env:  map[string]string{"XDG_CONFIG_HOME": "/xdg/config"},
			home: "/home/user",
			want: "/xdg/config",
		},
		{
			name: "an absolute XDG_CONFIG_HOME is taken as it stands",
			env:  map[string]string{"XDG_CONFIG_HOME": "/home/user/.config"},
			home: "/home/user",
			want: "/home/user/.config",
		},
		{
			name: "home holds it when XDG_CONFIG_HOME is unset",
			env:  map[string]string{},
			home: "/home/user",
			want: "/home/user/.config",
		},
		{
			// os.Getenv cannot tell a variable that is set but empty from
			// one that is unset, and the spec reads both as unset.
			name: "home holds it when XDG_CONFIG_HOME is set but empty",
			env:  map[string]string{"XDG_CONFIG_HOME": ""},
			home: "/home/user",
			want: "/home/user/.config",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := func(key string) string { return test.env[key] }

			got, err := xdgConfigDir(env, test.home)
			if err != nil {
				t.Fatalf("configuration directory of home %q: %v", test.home, err)
			}

			if got != test.want {
				t.Errorf("configuration directory = %q, want %q", got, test.want)
			}
		})
	}
}

func TestXDGConfigDirRefused(t *testing.T) {
	for _, test := range []struct {
		name string
		env  map[string]string
		home string
	}{
		{name: "no home", env: map[string]string{}, home: ""},
		{name: "no home and an empty XDG_CONFIG_HOME", env: map[string]string{"XDG_CONFIG_HOME": ""}, home: ""},
		{name: "a relative XDG_CONFIG_HOME", env: map[string]string{"XDG_CONFIG_HOME": ".config"}, home: "/home/user"},
		{name: "a relative XDG_CONFIG_HOME without a home", env: map[string]string{"XDG_CONFIG_HOME": ".config"}, home: ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			env := func(key string) string { return test.env[key] }

			dir, err := xdgConfigDir(env, test.home)
			if err == nil {
				t.Errorf("configuration directory returned %q and no error, want it refused", dir)
			}
		})
	}
}
