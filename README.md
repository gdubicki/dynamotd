# Dynamotd

We know: servers, VMs and bare metals are not cool nowadays but you may still have them and have to
take care of them, so how about making your life a little bit better?
 
This app provides a [MOTD (Message Of The Day)](https://en.wikipedia.org/wiki/Motd_(Unix)) optimized for 
providing the info about your server that YOU need to troubleshoot stuff when you SSH to it (because if you are doing
it, then something IS wrong with the server, isn't it?).   

## Features

* Useful default info - CPU Load, Memory use, Network use, Disk space use etc.,
* Customizable - reorder and remove default lines,
* Pluggable - add custom lines with any label and any command(s) output as the value,
* Semantic colors - for example red if load is higher than no of cores,
* Static and dynamic mode - the former just prints the current state, the latter updates every 2 seconds,
  like htop,
* Keyboard shortcuts to run popular tools (in dynamic mode only) - like htop, iftop etc.
* Fast - written in Go for that,
* Circuit-breaker - minimal info mode in case of server overload detection,
* Native packages for Centos (RPM) and Ubuntu (DEB) - so you can have on almost all of your servers.

## Status

POC.

TODO: Almost all of the above. :P

We have some default info, limited customizability, semantic colors, static and dynamic mode (although the latter is
not really dynamic yet).

We also don't have tests and the building pipeline (but they WILL be created if we decide to go ahead with the project).

Finally we don't have user docs, developer docs and contribution guide.

What we DO have is some non-idiomatic code and brain-dead solutions because the original author is only learning Golang.
:P

## Building

Requirements:
* Go v. ? (I am using 1.13),
* ?

How-to build:
```
go build
```
