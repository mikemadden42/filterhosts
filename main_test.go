package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// names reduces groups to device names so tests can compare them directly.
func names[T any](groups map[string][]T, name func(T) string) map[string][]string {
	out := make(map[string][]string)
	for key, devices := range groups {
		for _, d := range devices {
			out[key] = append(out[key], name(d))
		}
	}
	return out
}

func TestParseHosts(t *testing.T) {
	const header = "Status\tDevice Name\tIP Address\tMAC Address\tConnection Type\n"

	tests := []struct {
		name        string
		input       string
		want        map[string][]string
		wantSkipped []int
	}{
		{
			name:  "empty",
			input: "",
			want:  map[string][]string{},
		},
		{
			name:  "header only",
			input: header,
			want:  map[string][]string{},
		},
		{
			name: "groups by connection type",
			input: header +
				"On\tlaptop\t10.0.0.2\taa:bb\tWiFi\n" +
				"On\tdesktop\t10.0.0.3\tcc:dd\tEthernet\n" +
				"Off\tphone\t10.0.0.4\tee:ff\tWiFi\n",
			want: map[string][]string{
				"WiFi":     {"laptop", "phone"},
				"Ethernet": {"desktop"},
			},
		},
		{
			name: "skips short rows",
			input: header +
				"On\tlaptop\t10.0.0.2\taa:bb\tWiFi\n" +
				"On\tbroken\t10.0.0.3\n",
			want: map[string][]string{
				"WiFi": {"laptop"},
			},
			wantSkipped: []int{3},
		},
		{
			name: "ignores blank lines",
			input: header +
				"\n" +
				"On\tlaptop\t10.0.0.2\taa:bb\tWiFi\n" +
				"  \n",
			want: map[string][]string{
				"WiFi": {"laptop"},
			},
		},
		{
			name: "handles CRLF line endings",
			input: strings.ReplaceAll(header, "\n", "\r\n") +
				"On\tlaptop\t10.0.0.2\taa:bb\tEthernet\r\n" +
				"On\ttv\t10.0.0.5\tgg:hh\tEthernet\r\n",
			want: map[string][]string{
				"Ethernet": {"laptop", "tv"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups, skipped, err := parseHosts(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("parseHosts: %v", err)
			}
			got := names(groups, func(d hostDevice) string { return d.Name })
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(skipped, tt.wantSkipped) {
				t.Errorf("skipped lines: got %v, want %v", skipped, tt.wantSkipped)
			}
		})
	}
}

func TestParseLanHosts(t *testing.T) {
	const header = "Name\tStatus\tConnection\tSignal\n"

	tests := []struct {
		name           string
		input          string
		wantStatus     map[string][]string
		wantConnection map[string][]string
		wantSkipped    []int
	}{
		{
			name:           "empty",
			input:          "",
			wantStatus:     map[string][]string{},
			wantConnection: map[string][]string{},
		},
		{
			name:           "header only",
			input:          header,
			wantStatus:     map[string][]string{},
			wantConnection: map[string][]string{},
		},
		{
			name: "groups by status and connection",
			input: header +
				"phone\tOnline\tWi-Fi\tWi-Fi   Wi-Fi 3 bars\n" +
				"tv\tOffline\tEthernet\t-\n" +
				"laptop\tOnline\tWi-Fi\tWi-Fi   Wi-Fi 2 bars\n",
			wantStatus: map[string][]string{
				"Online":  {"phone", "laptop"},
				"Offline": {"tv"},
			},
			wantConnection: map[string][]string{
				"Wi-Fi":    {"phone", "laptop"},
				"Ethernet": {"tv"},
			},
		},
		{
			// Regression: full-strength Wi-Fi devices were once diverted out
			// of the connection groups.
			name:  "groups 5-bar Wi-Fi devices",
			input: header + "phone\tOnline\tWi-Fi\tWi-Fi   Wi-Fi 5 bars\n",
			wantStatus: map[string][]string{
				"Online": {"phone"},
			},
			wantConnection: map[string][]string{
				"Wi-Fi": {"phone"},
			},
		},
		{
			name: "skips short rows",
			input: header +
				"phone\tOnline\tWi-Fi\t-\n" +
				"broken\tOnline\n",
			wantStatus: map[string][]string{
				"Online": {"phone"},
			},
			wantConnection: map[string][]string{
				"Wi-Fi": {"phone"},
			},
			wantSkipped: []int{3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusGroups, connectionGroups, skipped, err := parseLanHosts(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("parseLanHosts: %v", err)
			}
			name := func(d lanDevice) string { return d.Name }
			if got := names(statusGroups, name); !reflect.DeepEqual(got, tt.wantStatus) {
				t.Errorf("status groups: got %v, want %v", got, tt.wantStatus)
			}
			if got := names(connectionGroups, name); !reflect.DeepEqual(got, tt.wantConnection) {
				t.Errorf("connection groups: got %v, want %v", got, tt.wantConnection)
			}
			if !reflect.DeepEqual(skipped, tt.wantSkipped) {
				t.Errorf("skipped lines: got %v, want %v", skipped, tt.wantSkipped)
			}
		})
	}
}

func TestRun(t *testing.T) {
	dir := t.TempDir()
	hostsFile := filepath.Join(dir, "hosts.tsv")
	lanFile := filepath.Join(dir, "lan-hosts.tsv")
	missingHosts := filepath.Join(dir, "missing-hosts.tsv")
	missingLan := filepath.Join(dir, "missing-lan.tsv")

	files := map[string]string{
		hostsFile: "Status\tDevice Name\tIP Address\tMAC Address\tConnection Type\n" +
			"On\tlaptop\t10.0.0.2\taa:bb\tWiFi\n",
		lanFile: "Name\tStatus\tConnection\tSignal\n" +
			"phone\tOnline\tWi-Fi\t-\n",
	}
	for path, data := range files {
		if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name      string
		hostsFile string
		lanFile   string
		wantErrs  []string
	}{
		{name: "both files", hostsFile: hostsFile, lanFile: lanFile},
		{name: "skip both", hostsFile: "", lanFile: ""},
		{name: "skip lan", hostsFile: hostsFile, lanFile: ""},
		{name: "missing hosts file", hostsFile: missingHosts, lanFile: lanFile, wantErrs: []string{missingHosts}},
		{
			// Both files are attempted even after the first one fails
			name:      "both missing",
			hostsFile: missingHosts,
			lanFile:   missingLan,
			wantErrs:  []string{missingHosts, missingLan},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(tt.hostsFile, tt.lanFile)
			if len(tt.wantErrs) == 0 {
				if err != nil {
					t.Fatalf("run: unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("run: got nil error, want errors mentioning %v", tt.wantErrs)
			}
			for _, want := range tt.wantErrs {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("run: error %q does not mention %q", err, want)
				}
			}
		})
	}
}
