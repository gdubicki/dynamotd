# Dynamotd

We know: servers, VMs and bare metals are not cool nowadays, but you may still have them and have to
take care of them, so how about making your life a little better?
 
This app provides a [MOTD (Message Of The Day)](https://en.wikipedia.org/wiki/Motd_(Unix)) optimized for 
providing the info about your server that YOU need to troubleshoot stuff when you SSH to it (because if you are doing
it, then something IS wrong with the server, isn't it?).   

## Screenshot 

<img src="https://raw.githubusercontent.com/gdubicki/dynamotd/main/dynamotd.png" width="1000px" alt="screenshot">

## Features

* **Useful default info** - CPU load, memory use, disk space use etc.,
* **Semantic colors** - <span style="color:red">red</span> → problem, <span style="color:yellow">yellow</span> → warning, <span style="color:green">green</span> → ok.
* **Customizable** - reorder or remove lines,
* **Fast** - native app,

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
| `ip` | Local and external IP | - |
| `uptime` | Time since the last boot | `warning: 90`, `critical: 365` - in days |
| `reboot` | That a reboot is required - only if it is: on Debian and Ubuntu if any package requires it, on other distros if the most recently installed kernel is not the running one | - |
| `clock` | That the clock is not synchronized - only if it isn't, and only on Linux. Determined the same way as in `timedatectl`. | - |
| `cores` | Number of CPU cores and their clock speed | - |
| `load` | Load averages and number of CPU cores | `warning: 0.9`, `critical: 1.1` - per CPU core |
| `cpu` | How the CPU time is spent: in user space, in the kernel, waiting for I/O and stolen by the hypervisor (the last two only on Linux) | `iowait: {warning: 10, critical: 25}`, `steal: {warning: 5, critical: 15}` - in % |
| `memory` | RAM usage | `warning: 90`, `critical: 95` - in % |
| `swap` | Swap usage, and the number of processes killed by the OOM killer since boot - only if there were any | `warning: 50`, `critical: 80` - in % of swap used |
| `topprocesses` | The processes using the most CPU - only if the load is high, and the ones using the most RAM - only if its usage is high | `load: 0.9` - per CPU core, `memory: 90` - in %, `count: 3` - number of processes, `always: false` - to show them always |
| `diskspace` | Disk space usage of `/`, and of any other writable filesystem that reached the warning threshold for space or inodes. Inodes usage is shown only if it reached its warning threshold. | `warning: 75`, `critical: 90` - in % of space used, `inodes: {warning: 75, critical: 90}` - in % of inodes used, `always: ["/"]` - filesystems always shown, `others: true` - whether to show the other filesystems |
| `failedunits` | Failed systemd units - only if there are any | - |
| empty or `emptyLine` | Empty line | - |

Check out all the available command-line arguments by running `dynamotd -help`.

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

How to release: push a `v*` tag. GitHub Actions will build the binaries and packages and create a draft release.

Optional, but recommended: use [pre-commit](https://pre-commit.com) and install the apps used by our config.

## TODO

These features ~~will~~ may be implemented in the future:

* **Pluggable** - add custom lines with any label and any command(s) output as the value,
* **Interactive mode** - show the output like `htop`, updated every 2 seconds,
* **Keyboard shortcuts** - in interactive mode only, to run popular tools like htop, iftop etc.
* **Circuit-breaker** - minimal info mode in case of server overload detection,
* **Native packages for RedHat/Centos/Rocky and Debian/Ubuntu** - easy install it on (almost) all of your servers.

We also don't have user docs, developer docs and contribution guide.

What we DO have is some non-idiomatic code because the original author is only learning Golang.
:P

## License

[MIT](./LICENSE)
