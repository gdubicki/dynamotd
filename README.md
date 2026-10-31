# Dynamotd

We know: servers, VMs and bare metals are not cool nowadays, but you may still have them and have to
take care of them, so how about making your life a little better?
 
This app provides a [MOTD (Message Of The Day)](https://en.wikipedia.org/wiki/Motd_(Unix)) optimized for 
providing the info about your server that YOU need to troubleshoot stuff when you SSH to it (because if you are doing
it, then something IS wrong with the server, isn't it?).   

## Screenshots

When all is good, it shows the basic info about your server:

<img src="https://raw.githubusercontent.com/gdubicki/dynamotd/main/screenshot-all-good.png" width="923px" alt="screenshot of a server without problems">

When there are problems, they are shown in yellow and red, and some rows appear only then:

<img src="https://raw.githubusercontent.com/gdubicki/dynamotd/main/screenshot-all-problems.png" width="1000px" alt="screenshot of a server with all kinds of problems">

## Features

* **Useful default info** - CPU load and usage (incl. waiting for disk and stolen by hypervisor), memory and swap use, disk space use of all the filesystems that need attention, top processes when the load or memory use is high etc.,
* **Shows problems only when there are any** - failed systemd units, OOM kills, required reboot, unsynchronized clock, other logged in users,
* **Semantic colors** - <span style="color:red">red</span> → problem, <span style="color:yellow">yellow</span> → warning, <span style="color:green">green</span> → ok.
* **Never blocks your login** - the rows that don't finish in 1s (configurable), f.e. because of a broken DNS, are shown as "timed out",
* **Customizable** - reorder or remove rows, change the thresholds, add rows with the output of your own commands,
* **Fast** - native app, generating all the rows concurrently,
* **Easy to install** - static binaries working on any Linux distro, and packages for Debian/Ubuntu and RedHat/Rocky/Fedora etc.

## Installation

1. Download the latest binary for your OS and arch and make it executable, f.e. for Linux on AMD64:
```
curl -L https://github.com/gdubicki/dynamotd/releases/latest/download/dynamotd-linux-amd64 -o /usr/local/bin/dynamotd && chmod +x /usr/local/bin/dynamotd
```
As it will run as root on every login, it's worth verifying that the binary was built by this repo's GitHub Actions from its source code, using the [GitHub CLI](https://cli.github.com/):
```
gh attestation verify /usr/local/bin/dynamotd --repo gdubicki/dynamotd
```
Or at least compare its checksum with the `checksums.txt` file from the same release.

The Linux binaries are static, so they work on any distro. Alternatively, install the `.deb` or `.rpm` package from the [latest release](https://github.com/gdubicki/dynamotd/releases/latest) - it installs the binary as `/usr/bin/dynamotd` and the example config as `/etc/dynamotd.yaml`.

2. (Optionally) Configure with `/etc/dynamotd.yaml` file (or `~/.dynamotd/dynamotd.yaml` - the first one found is used, or any file passed with `-config`). See the example config in [dynamotd.yaml](./dynamotd.yaml).

3. Make it shown on login:

* If you installed the `.deb` package on Debian or Ubuntu, then there is nothing to do - it is shown by the `/etc/update-motd.d/50-dynamotd` script that comes with it. You may want to remove the other scripts in that directory and the static `/etc/motd` though. To use `-no-color` (see below), edit that script. To disable it, delete it - package upgrades won't restore it. If you already run dynamotd from `/etc/pam.d/` (see below), the package removes that script itself, so the MOTD isn't shown twice.

* Otherwise, make it shown instead of or after your default static MOTD by editing `/etc/pam.d/sshd`. Find line with `pam_motd.so` and either replace or append this line to it:
```
session    optional     pam_exec.so stdout /usr/local/bin/dynamotd -force-color
```
(or `-no-color` if you prefer plain black-and-white output). Use `/usr/bin/dynamotd` if you installed a package.

## Troubleshooting

**Problem**: dynamic MOTD is not shown during login.

**Solution(s)**: there could be multiple reason for this.

One of them is if you use SSH multiplexing (`ControlPath`, `ControlMaster`, `ControlPersist` in your SSH config) - then the MOTD will be shown only during creation of the first connection to the server. Read more about it [here](https://blog.plover.com/Unix/sshd.html).

**Problem**: the OOM kills row shows only their number, without when the last one happened.

**Solution**: that info comes from the kernel log, which on most distros only root can read. It is shown when dynamotd runs on login, as it runs as root then, and when you run it with `sudo`.

TODO: expand this section as more cases are identified.

## Configuration

See the example config in [dynamotd.yaml](./dynamotd.yaml).

After changing the config, check it with `dynamotd -check`.

### Row types

These are the row types you can use in the `rows` list. Some of them have settings, which you can change like this:
```yaml
rows:
  - memory:
      warning: 80
      critical: 90
```
The `warning` and `critical` thresholds define when the value is shown in yellow or red.

| Row type | Shows | Settings (defaults) |
|---|---|---|
| `timestamp` | Current date and time | - |
| `fqdn` | Fully qualified domain name | - |
| `ip` | Local and external IPv4 and IPv6 (the latter only if the machine has it), shown once if they are the same ("Public IPv4") | - |
| `uptime` | Time since the last boot | `warning: 90`, `critical: 365` - in days |
| `users` | Who else is logged in - only if anyone is | - |
| `reboot` | That a reboot is required - only if it is: on Debian and Ubuntu if any package requires it, on other distros if the most recently installed kernel is not the running one | - |
| `clock` | That the clock is not synchronized - only if it isn't, and only on Linux. Determined the same way as in `timedatectl`. | - |
| `cores` | Number of CPU cores and their clock speed | - |
| `load` | Load averages and number of CPU cores | `warning: 0.9`, `critical: 1.1` - per CPU core |
| `cpu` | How busy the CPUs are (running programs' code - "user", or the kernel's), how much time they wait for the disk, and how much time the hypervisor gives to other VMs instead of this one - "stolen" (the last two only on Linux, and only if they reach the `show` threshold) | `iowait: {show: 5, warning: 10, critical: 25}`, `steal: {show: 2, warning: 5, critical: 15}` - in % |
| `memory` | RAM usage, and swap usage - only if there is any swap | `warning: 90`, `critical: 95` - in % of RAM used, `swap: {warning: 50, critical: 80}` - in % of swap used |
| `oomkills` | Number of processes killed by the OOM killer since boot - only if there were any, with when the last one was killed and its name (if the kernel log is readable, which requires root; if the kill is no longer in the kernel's log buffer, then at least that it happened before the oldest entry there) - in red if it was in the last 24h, in yellow if earlier | - |
| `topprocesses` | The processes using the most CPU - only if the load is high, and the ones using the most RAM - only if its usage is high | `load: 0.9` - per CPU core, `memory: 90` - in %, `count: 3` - number of processes, `always: false` - to show them always |
| `diskspace` | Disk space usage of `/`, and of any other writable filesystem that reached the warning threshold for space or inodes. Inodes usage is shown only if it reached its warning threshold. | `warning: 75`, `critical: 90` - in % of space used, `inodes: {warning: 75, critical: 90}` - in % of inodes used, `always: ["/"]` - filesystems always shown, `others: true` - whether to show the other filesystems |
| `failedunits` | Failed systemd units - only if there are any | - |
| `command` | First line of the output of a custom command, see below | `label` and `run` - both required |
| empty or `emptyLine` | Empty line | - |

Check out all the available command-line arguments by running `dynamotd -help`.

### Custom commands

To show the output of any command, use the `command` row type, f.e.:
```yaml
rows:
  - command:
      label: Containers
      run: docker ps -q | wc -l
```
The command is run with `/bin/sh` and the first line of its output is shown. Its exit code determines the color, like for the [Nagios plugins](https://nagios-plugins.org/doc/guidelines.html#AEN78), so you can use them as is: 0 - normal, 1 - warning (yellow), 2 and other - critical (red). Commands that don't finish within the `timeout` are killed and shown as "timed out".

As the commands are run as the user running dynamotd, usually root, the config file with them must be owned by root (or that user) and not writable by group or others.

## Contributing

Requirements:
* Go v. 1.26+

How to build for your current OS and arch (the output name is needed, as `dynamotd` is a directory):
```
go build -o dynamotd-dev
```

How to build all the release binaries and packages locally (requires [GoReleaser](https://goreleaser.com)):
```
goreleaser release --snapshot --clean
```

How to run tests and the linter (requires [golangci-lint](https://golangci-lint.run)):
```
go test ./...
golangci-lint run
```

How to add a new row type: create a file in [plugins/](./plugins) with a function returning a `dynamotd.Row` and register it under a name (to be used in the config file) in an `init()` function - see the existing plugins.

How to regenerate the screenshots above after changing the output (requires ImageMagick, pngquant and the Monaco font, included in macOS):
```
tools/screenshots/generate.sh
```

How to release: push a `v*` tag. GitHub Actions will build the binaries and packages and create a draft release.

Optional, but recommended: use [pre-commit](https://pre-commit.com) and install the apps used by our config.

## TODO

These features ~~will~~ may be implemented in the future:

* **Interactive mode** - show the output like `htop`, updated every 2 seconds,
* **Keyboard shortcuts** - in interactive mode only, to run popular tools like htop, iftop etc.
* **Circuit-breaker** - minimal info mode in case of server overload detection.

## License

[MIT](./LICENSE)
