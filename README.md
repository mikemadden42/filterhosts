# filterhosts

Summarize router host exports. `filterhosts` reads tab-separated device
lists and prints the devices grouped by connection type and by status.

## Install

```sh
go install github.com/mikemadden42/filterhosts@latest
```

Requires Go 1.26 or newer.

## Usage

```sh
filterhosts                                  # read hosts.tsv and lan-hosts.tsv
filterhosts -hosts router.tsv -lan lan.tsv   # custom paths
filterhosts -lan ""                          # process only the hosts file
```

| Flag     | Default         | Description                                                       |
| -------- | --------------- | ----------------------------------------------------------------- |
| `-hosts` | `hosts.tsv`     | Hosts export to group by connection type (`""` to skip)           |
| `-lan`   | `lan-hosts.tsv` | LAN hosts export to group by status and connection (`""` to skip) |

Errors and warnings go to stderr. The exit status is 0 on success, 1 if a
file can't be read, and 2 for bad usage.

## Input files

Both files are tab-separated. The first line is always treated as a header
and skipped, and blank lines are ignored. Rows with too few columns are
skipped with a warning such as
`filterhosts: lan-hosts.tsv:5: skipping row with too few fields`.
Extra columns are ignored.

### `hosts.tsv`

| Column | Field           |
| ------ | --------------- |
| 1      | Status          |
| 2      | Device Name     |
| 3      | IP Address      |
| 4      | MAC Address     |
| 5      | Connection Type |

### `lan-hosts.tsv`

| Column | Field                                    |
| ------ | ---------------------------------------- |
| 1      | Name                                     |
| 2      | Status                                   |
| 3      | Connection                               |
| 4      | Required but unused (for example signal) |

## Example

`hosts.tsv`:

```
Status	Device Name	IP Address	MAC Address	Connection Type
Online	laptop	192.168.1.20	aa:bb:cc:dd:ee:01	Wi-Fi
Online	desktop	192.168.1.21	aa:bb:cc:dd:ee:02	Ethernet
Offline	phone	192.168.1.22	aa:bb:cc:dd:ee:03	Wi-Fi
```

`lan-hosts.tsv`:

```
Name	Status	Connection	Signal
laptop	Online	Wi-Fi	Wi-Fi 5 bars
tv	Offline	Ethernet	-
phone	Online	Wi-Fi	Wi-Fi 3 bars
```

Output:

```
Devices Grouped by Connection Type (hosts.tsv):
Ethernet:
	- desktop
Wi-Fi:
	- laptop
	- phone

Devices Grouped by Status (lan-hosts.tsv):
Offline:
	- tv
Online:
	- laptop
	- phone

Devices Grouped by Connection (lan-hosts.tsv):
Ethernet:
	- tv
Wi-Fi:
	- laptop
	- phone
```

Groups are sorted by name; devices within a group keep their order from
the file.

## Development

```sh
go test ./...   # run tests
make check      # golangci-lint
make build      # snapshot build with goreleaser
```

## License

MIT. See [LICENSE](LICENSE).
