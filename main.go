package main

import (
	"bufio"
	"fmt"
	"os"
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
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Initialize a map to store devices grouped by connection type
	connectionTypeGroups := make(map[string][]Device)

	// Read the file line by line
	scanner := bufio.NewScanner(file)
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

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Print the summary by Connection Type
	fmt.Printf("\nDevices Grouped by Connection Type (%s):\n", filename)
	for connectionType, devices := range connectionTypeGroups {
		fmt.Printf("%s:\n", connectionType)
		for _, device := range devices {
			fmt.Printf("\t- %s\n", device.Name)
		}
	}
}

func processLanHostsFile(filename string) {
	// Open the data file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Initialize maps to store devices
	connectionGroups := make(map[string][]LanDevice)
	statusGroups := make(map[string][]LanDevice)

	// Initialize slices to store Connection and Status summary lines
	var connectionSummaryLines []string
	var statusSummaryLines []string

	// Track if the current line is the header
	isHeader := true

	// Read the file line by line
	scanner := bufio.NewScanner(file)
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

			// Check for Connection summary
			if fields[3] == "Wi-Fi   Wi-Fi 5 bars" {
				connectionSummaryLines = append(connectionSummaryLines, line)
			} else {
				// Group devices by Connection type
				connectionGroups[device.Connection] = append(connectionGroups[device.Connection], device)
			}

			// Check for Status summary
			if fields[1] == "Status" {
				statusSummaryLines = append(statusSummaryLines, line)
			} else {
				// Group devices by Status type
				statusGroups[device.Status] = append(statusGroups[device.Status], device)
			}
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Print the Status summary lines first
	fmt.Printf("\nStatus Summary (%s):\n", filename)
	for _, line := range statusSummaryLines {
		fmt.Println(line)
	}

	// Print the summary by Status type
	fmt.Printf("\nDevices Grouped by Status (%s):\n", filename)
	for status, devices := range statusGroups {
		fmt.Printf("%s:\n", status)
		for _, device := range devices {
			fmt.Printf("\t- %s\n", device.Name)
		}
	}

	// Print the Connection summary lines
	fmt.Printf("\nConnection Summary (%s):\n", filename)
	for _, line := range connectionSummaryLines {
		fmt.Println(line)
	}

	// Print the summary by Connection type
	fmt.Printf("\nDevices Grouped by Connection (%s):\n", filename)
	for connection, devices := range connectionGroups {
		fmt.Printf("%s:\n", connection)
		for _, device := range devices {
			fmt.Printf("\t- %s\n", device.Name)
		}
	}
}
