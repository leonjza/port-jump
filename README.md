# port-jump

Some security by obscurity using "port-jumping". A silly PoC to use HOTP to update port numbers to a service as time progresses.

![logo](./images/port-jump.png)

> [!WARNING]
> This is a PoC-scratching-an-itch project. Don't actually use this somewhere important, okay? Nothing beats a firewall actually blocking stuff.

## introduction

Port jumping is a post-wake up "hmmm" idea that I wanted to PoC. This code is that result.

The idea is simple. Instead of having a service like SSH statically listen on port 22 (or whatever port you use) forever, what if that port number changed every `$interval`? Sounds like an excellent security by obscurity choice! This project does that by implementing an HOTP generator based on a secret, generating valid TCP port numbers within a range to use.

Using a simple config file in `~/.config/port-jump/config.yml`, shared secrets and port mappings are read, and rotated on a configured interval, just like a TOTP does! An example configuration is:

```yml
jumps:
  - enabled: false
    dstport: 23
    interval: 30
    sharedsecret: YIHWTYNSBRGWFPR4
  - enabled: true
    dstport: 22
    interval: 30
    sharedsecret: FWX2CC3PLA4ZYGCI
  - enabled: true
    dstport: 80
    interval: 60
    sharedsecret: HPQY7R45TFSZWTST
```

This configuration has three jumps configured, with one being disabled.

Assuming we're targeting SSH, you can now get the remote service port by running `port-jump get port` as follows:

```console
ssh -p $(port-jump get port -p22) user@10.211.55.6
```

Or, if its say a web service, how about:

```console
curl $(port-jump get uri --url remote-service.local -p 443)
```

Of course, you could also just update the port section of a URL, just like the SSH example:

```console
curl https://remote-service.local:$(port-jump get port -p 443)/
```

## example run

In the below image, in the bottom panes I have an ubuntu server running the `port-jump jump` command that reads the configuration file and updates `nftables` to NAT incoming connections to port 22. In the top pane is a macOS SSH client that uses the `port-jump get port` command to get the current port to use to connect to the remote SSH service. This command is run every 30 seconds as an example as the configured interval changes the port.

![example](./images/example.png)

## building & installing

How you want to install this depends on what you prefer. You can run it as is, as a [systemd unit](#systemd-unit) or as a [docker container](#docker). Do what works for you.

In most cases you'd need to build the program though, so do that with:

```console
go build -o port-jump
```

You can also install it using `go install` which would typically install `port-jump` to wherever `GOBIN` points to.

```console
go install github.com/leonjza/port-jump@latest
```

### systemd unit

A systemd [unit](./port-jump.service) is available that will start the `port-jump jump` command as a systemd service. To install:

*Note:* If you have no jumps configured in the configuration file, one will be added as an example, but will be disabled. With no enabled jumps, the service will exit. Be sure to check out `~/.config/port-jump/config.yml` to configure your jumps.

- Copy the example unit file over to something like `/etc/systemd/system/port-jump.servive`
- Make sure the contents reflects the correct paths where you put your build of `port-jump`.
- Reload the available daemons with `systemctl daemon-reload`.
- Enable the service with `systemctl enable port-jump.service`.
- Start the service with `systemctl start port-jump.service`.
- Check out the status of `port-jump` with `systemctl status port-jump.service`.
- Check out the logs with `journalctl -fu port-jump.service`.

### docker

It's possible to run `port-jump` using Docker. It’s going to require the `--privileged` flag which is generally discouraged. However, assuming you trust this code and understand what that flag means, you could get a docker container up and running.

Build it with:

```console
docker build -t portjump:local .
```

Then run it with:

```console
docker run --rm -it \
  -v /root/.config/port-jump/config.yml:/root/.config/port-jump/config.yml \
  --network host \
  --privileged \
  portjump:local jump
```

Note the volume mapping with `-v`. This is where the jump mapping lives.

## configuration

Configuration for `port-jump` lives in a configuration file. Depending on your Operating System, this may be relative to wherever the Golang [os.UserHomeDir](https://pkg.go.dev/os#UserHomeDir) call resolves to as `$HOME`. i.em, `$HOME/.config/port-jump/config.yml`. Feel free to edit this file manually, or use the `port-jump config` set of commands to perform create, toggle and delete operations on the configuration.

```console
$ port-jump config
Work with port-jump configurations

Usage:
  port-jump config [command]

Available Commands:
  add         Add a new jump
  delete      Delete a jump
  list        List the current jumps
  toggle      Toggle jump status

Flags:
  -h, --help   help for config

Global Flags:
  -D, --debug   debug

Use "port-jump config [command] --help" for more information about a command.
```

## todo

This is a PoC, but to give you an idea of stuff to do includes:

- Adding a floor / ceiling limit to a jump so that ports do not overlap with existing services that may already be running.
- Add some more firewall support. Right now only `nftables` is supported on Linux.
- IPv6 Suport.
- Potentially faster interval support <https://infosec.exchange/@singe@chaos.social/113057901149163673>
- Use names / id's to identify jumps. Right now, it's just a port map, but what if you want more than one service, on the same port but separate keys?
