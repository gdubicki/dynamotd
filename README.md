# Dynamotd

We know: servers, VMs and bare metals are not cool nowadays, but you may still have them and have to
take care of them, so how about making your life a little better?
 
This app provides a [MOTD (Message Of The Day)](https://en.wikipedia.org/wiki/Motd_(Unix)) optimized for 
providing the info about your server that YOU need to troubleshoot stuff when you SSH to it (because if you are doing
it, then something IS wrong with the server, isn't it?).   

## Screenshot 

<img src="https://raw.githubusercontent.com/gdubicki/dynamotd/main/dynamotd.png" width="1000px" alt="logo">

## Features

* **Useful default info** - CPU load, memory use, disk space use etc.,
* **Semantic colors** - <span style="color:red">red</span> → problem, <span style="color:yellow">yellow</span> → warning, <span style="color:green">green</span> → ok.
* **Customizable** - reorder or remove lines,
* **Fast** - native app,

## Installation

1. Download the lastest binary for your OS and arch and make it executable:
```
curl -L https://github.com/gdubicki/dynamotd/releases/latest/download/dynamotd-linux-amd64 -o /usr/local/bin/dynamotd && chmod +x /usr/local/bin/dynamotd
```
2. (Optionally) Configure with `/etc/dynamotd.yaml` file. See the example config in [dynamotd.yaml](./dynamotd.yaml).

3. Make it shown instead of or after your default static MOTD by editing `/etc/pam.d/sshd`. Find line with `pam_motd.so` and either replace or append this line to it:
```
session    optional pam_exec.so stdout /usr/local/bin/dynamotd -force-color
```
(or `-no-color` if you prefer plain black-and-white output).

## Troubleshooting

**Problem**: dynamic MOTD is not shown during login.

**Solution(s)**: there could be multiple reason for this.

One of them is if you use SSH multiplexing (`ControlPath`, `ControlMaster`, `ControlPersist` in your SSH config) - then the MOTD will be shown only during creation of the first connection to the server. Read more about it [here](https://blog.plover.com/Unix/sshd.html).

TODO: expand this section as more cases are identified.

## Configuration

See the example config in [dynamotd.yaml](./dynamotd.yaml).

Check out all the available command-line arguments by running `dynamotd -help`.

## Contributing

Requirements:
* Go v. 1.17+

How to build:
```
./build.sh
```

How to run tests:
```
go test -v tests/*
```

Optional, but recommended: use [pre-commit](https://pre-commit.com) and install use apps.

## TODO

These features will be implemented in the future:

* **Pluggable** - add custom lines with any label and any command(s) output as the value,
* **Interactive mode** - show the output like `htop`, updated every 2 seconds,
* **Keyboard shortcuts** - in interactive mode only, to run popular tools like htop, iftop etc.
* **Circuit-breaker** - minimal info mode in case of server overload detection,
* **Native packages for RedHat/Centos/Rocky and Debian/Ubuntu** - easy install it on (almost) all of your servers.

We also don't have user docs, developer docs and contribution guide.

What we DO have is some non-idiomatic code and brain-dead solutions because the original author is only learning Golang.
:P
