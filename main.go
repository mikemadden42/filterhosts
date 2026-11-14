// Command filterhosts summarizes router host exports, grouping devices by
// connection type and status. The -hosts and -lan flags name the input files
// (hosts.tsv and lan-hosts.tsv by default).
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// hostDevice is one row of a hosts.tsv export.
type hostDevice struct {
	Status         string
	Name           string
	IPAddress      string
	MACAddress     string
	ConnectionType string
}

// lanDevice is one row of a lan-hosts.tsv export.
type lanDevice struct {
	Name       string
	Status     string
	Connection string
}

func main() {
	hostsFile := flag.String("hosts", "hosts.tsv", "hosts export to group by connection type (\"\" to skip)")
	lanFile := flag.String("lan", "lan-hosts.tsv", "LAN hosts export to group by status and connection (\"\" to skip)")
	flag.Parse()

	// All input comes from flags, so extra arguments are a usage error
	if flag.NArg() > 0 {
		flag.Usage()
		os.Exit(2)
	}

	if err := run(*hostsFile, *lanFile); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run processes the given hosts and LAN hosts files, skipping any whose name
// is empty. A failure in one file does not stop the other from being
// processed; all errors are returned joined.
func run(hostsFile, lanFile string) error {
	var errs []error

	if hostsFile != "" {
		if err := processHostsFile(hostsFile); err != nil {
			errs = append(errs, fmt.Errorf("filterhosts: %w", err))
		}
	}

	if lanFile != "" {
		if err := processLanHostsFile(lanFile); err != nil {
			errs = append(errs, fmt.Errorf("filterhosts: %w", err))
		}
	}

	return errors.Join(errs...)
}

// warnSkipped reports rows the parser skipped for having too few fields.
func warnSkipped(filename string, skipped []int) {
	for _, line := range skipped {
		fmt.Fprintf(os.Stderr, "filterhosts: %s:%d: skipping row with too few fields\n", filename, line)
	}
}

func processHostsFile(filename string) (err error) {
	// Open the data file
	file, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	connectionTypeGroups, skipped, err := parseHosts(file)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filename, err)
	}
	warnSkipped(filename, skipped)

	// Print the summary by Connection Type
	printGroups(fmt.Sprintf("Devices Grouped by Connection Type (%s)", filename),
		connectionTypeGroups, func(d hostDevice) string { return d.Name })
	return nil
}

// parseHosts reads hosts.tsv data and groups devices by connection type.
// It also returns the line numbers of rows skipped for having too few fields.
func parseHosts(r io.Reader) (map[string][]hostDevice, []int, error) {
	// Initialize a map to store devices grouped by connection type
	connectionTypeGroups := make(map[string][]hostDevice)

	skipped, err := scanRows(r, 5, func(fields []string) {
		// Extract Device information
		device := hostDevice{
			Status:         fields[0],
			Name:           fields[1],
			IPAddress:      fields[2],
			MACAddress:     fields[3],
			ConnectionType: fields[4],
		}

		// Group devices by Connection Type
		connectionTypeGroups[device.ConnectionType] = append(connectionTypeGroups[device.ConnectionType], device)
	})

	return connectionTypeGroups, skipped, err
}

func processLanHostsFile(filename string) (err error) {
	// Open the data file
	file, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	statusGroups, connectionGroups, skipped, err := parseLanHosts(file)
	if err != nil {
		return fmt.Errorf("reading %s: %w", filename, err)
	}
	warnSkipped(filename, skipped)

	name := func(d lanDevice) string { return d.Name }

	// Print the summary by Status type
	printGroups(fmt.Sprintf("Devices Grouped by Status (%s)", filename), statusGroups, name)

	// Print the summary by Connection type
	printGroups(fmt.Sprintf("Devices Grouped by Connection (%s)", filename), connectionGroups, name)
	return nil
}

// parseLanHosts reads lan-hosts.tsv data and groups devices by status and
// by connection. It also returns the line numbers of rows skipped for having
// too few fields.
func parseLanHosts(r io.Reader) (statusGroups, connectionGroups map[string][]lanDevice, skipped []int, err error) {
	// Initialize maps to store devices
	statusGroups = make(map[string][]lanDevice)
	connectionGroups = make(map[string][]lanDevice)

	skipped, err = scanRows(r, 4, func(fields []string) {
		// Extract Device information
		device := lanDevice{
			Name:       fields[0],
			Status:     fields[1],
			Connection: fields[2],
		}

		// Group devices by Status and by Connection type
		statusGroups[device.Status] = append(statusGroups[device.Status], device)
		connectionGroups[device.Connection] = append(connectionGroups[device.Connection], device)
	})

	return statusGroups, connectionGroups, skipped, err
}

// scanRows calls row with the tab-separated fields of each line after the
// header. Blank lines are ignored; lines with fewer than minFields fields are
// skipped and their 1-based line numbers returned.
func scanRows(r io.Reader, minFields int, row func(fields []string)) ([]int, error) {
	var skipped []int

	// Read the file line by line
	scanner := bufio.NewScanner(r)
	for lineNum := 1; scanner.Scan(); lineNum++ {
		line := scanner.Text()

		// Skip the header line and blank lines
		if lineNum == 1 || strings.TrimSpace(line) == "" {
			continue
		}

		// Split the line using the tab character as the delimiter
		fields := strings.Split(line, "\t")

		// Ensure the line has enough fields
		if len(fields) < minFields {
			skipped = append(skipped, lineNum)
			continue
		}

		row(fields)
	}

	return skipped, scanner.Err()
}

// printGroups prints each group's devices under its key, with keys sorted so
// output is stable.
func printGroups[T any](title string, groups map[string][]T, name func(T) string) {
	fmt.Printf("\n%s:\n", title)
	for _, key := range sortedKeys(groups) {
		fmt.Printf("%s:\n", key)
		for _, device := range groups[key] {
			fmt.Printf("\t- %s\n", name(device))
		}
	}
}

// sortedKeys returns the keys of m in sorted order so output is stable.
func sortedKeys[T any](m map[string][]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
