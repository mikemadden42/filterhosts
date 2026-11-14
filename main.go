// Command filterhosts summarizes router host exports (hosts.tsv and
// lan-hosts.tsv), grouping devices by connection type and status.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type Device struct {
	Status         string
	Name           string
	IPAddress      string
	MACAddress     string
	ConnectionType string
}

type LanDevice struct {
	Name       string
	Status     string
	Connection string
}

func main() {
	// Process "hosts.tsv"
	processHostsFile("hosts.tsv")

	// Process "lan-hosts.tsv"
	processLanHostsFile("lan-hosts.tsv")
}

func processHostsFile(filename string) {
	// Open the data file
	file, err := os.Open(filepath.Clean(filename))
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println("Error closing file:", err)
		}
	}()

	connectionTypeGroups, err := parseHosts(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Print the summary by Connection Type
	printGroups(fmt.Sprintf("Devices Grouped by Connection Type (%s)", filename),
		connectionTypeGroups, func(d Device) string { return d.Name })
}

// parseHosts reads hosts.tsv data and groups devices by connection type.
func parseHosts(r io.Reader) (map[string][]Device, error) {
	// Initialize a map to store devices grouped by connection type
	connectionTypeGroups := make(map[string][]Device)

	// Read the file line by line
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip the header line
		if strings.Contains(line, "Status\tDevice Name\tIP Address\tMAC Address\tConnection Type") {
			continue
		}

		// Split the line using the tab character as the delimiter
		fields := strings.Split(line, "\t")

		// Ensure the line has enough fields
		if len(fields) >= 5 {
			// Extract Device information
			device := Device{
				Status:         fields[0],
				Name:           fields[1],
				IPAddress:      fields[2],
				MACAddress:     fields[3],
				ConnectionType: fields[4],
			}

			// Group devices by Connection Type
			connectionTypeGroups[device.ConnectionType] = append(connectionTypeGroups[device.ConnectionType], device)
		}
	}

	return connectionTypeGroups, scanner.Err()
}

func processLanHostsFile(filename string) {
	// Open the data file
	file, err := os.Open(filepath.Clean(filename))
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Println("Error closing file:", err)
		}
	}()

	statusGroups, connectionGroups, err := parseLanHosts(file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	name := func(d LanDevice) string { return d.Name }

	// Print the summary by Status type
	printGroups(fmt.Sprintf("Devices Grouped by Status (%s)", filename), statusGroups, name)

	// Print the summary by Connection type
	printGroups(fmt.Sprintf("Devices Grouped by Connection (%s)", filename), connectionGroups, name)
}

// parseLanHosts reads lan-hosts.tsv data and groups devices by status and
// by connection.
func parseLanHosts(r io.Reader) (statusGroups, connectionGroups map[string][]LanDevice, err error) {
	// Initialize maps to store devices
	statusGroups = make(map[string][]LanDevice)
	connectionGroups = make(map[string][]LanDevice)

	// Track if the current line is the header
	isHeader := true

	// Read the file line by line
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip the header line
		if isHeader {
			isHeader = false
			continue
		}

		fields := strings.Split(line, "\t")

		// Ensure the line has enough fields
		if len(fields) >= 4 {
			// Extract Device information
			device := LanDevice{
				Name:       fields[0],
				Status:     fields[1],
				Connection: fields[2],
			}

			// Group devices by Status and by Connection type
			statusGroups[device.Status] = append(statusGroups[device.Status], device)
			connectionGroups[device.Connection] = append(connectionGroups[device.Connection], device)
		}
	}

	return statusGroups, connectionGroups, scanner.Err()
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
