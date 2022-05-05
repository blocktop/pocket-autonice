# pocket-autonice

When running a [pocket node](https://docs.pokt.network/core/) and blockchain cluster, it is advantageous to the node 
runner and good for the network overall that the cluster perform to its maximum capability. This project helps achieve 
that aim by boosting the priority of the pocket and the relay chain processes when relays are being served.

## Requirements

The program uses the Linux `renice` command to boost the processes of *the user* under which the pocket and relay chain
processes are running. Why not use the specific pids of the pocket and relay chian processes? That would require
reconfiguring `pocket-autonice` everytime a process was restarted and receives a new pid. 

So that means that the pocket and relay chain processes must each be running under a different Linux user. For services
running under `systemd`, the user can be set with

```text
[Service]
User=<username>
Group=<groupname>
ExecStart=...
```

For services running in docker, start the docker container with the `--user` argument specifying the user it should
run under.

## Installation

Start with just a few servers in your cluster to get the hang of this. Maybe even start locally on your pocket node.
Eventually you will want to install pocket-autonice on all your servers and configure it to coordinate process
boosting across your cluster.

**Dependencies:**

* Install [zeromq](https://github.com/zeromq/goczmq) dependencies from the system package manager. For example on 
Ubuntu 20.04:

```shell
apt install -y libtool pkg-config gcc g++ make libsodium-dev libczmq-dev
```

* Install [Go 1.17+](https://go.dev/doc/install). Ensure that the `go` executables as well as `$GOPATH/bin` are in 
your `$PATH`.


* Install `pocket-autonice`:

```shell
go install github.com/blocktop/pocket-autonice@latest
```

## How to boost your pocket and relay chain performance

Follow the steps below carefully. 

### Chains configuration

As root, dump the `config_example.yaml` file by running:

```shell
pocket-autonice dump-config
```

The file will be saved in the `$HOME/.pocket-autonice` directory. Rename this file to `config.yaml` in the same
directory.

On each server, set the relay chains that are active on that server. This should be a map of relay chain ID to 
the Linux user running that relay chain process. You should also include pocket-core in this configuration on the 
server running pocket-core.

```yaml
chains:
  "0001": pocket
  "0009": polygon
  "0021": geth
```

### Networking configuration

TL;DR:

On pocket node:

```yaml
subscriber_address: 10.0.0.2:5555
publish_to_endpoints:
  - 10.0.0.2:5555
  - 10.0.0.3:5555
  - all servers... 
```

On servers with relay chains only:

```yaml:
subscriber_address: 10.0.0.3:5555
```

Any open port can be used. For a single server setup, replace the IP address with `127.0.0.1`.

#### Details

Pocket-autonice works by looking for changes in the relay counts in the prometheus output of the pocket-core process.
When a change is detected, a message is published on a pub-sub channel containing the chain ID that is currently
being served. All servers in the cluster (including the pocket node) will receive that message and adjust the
priority of that relay chain if it is configured (see above). In addition, a message of `0001` is published when
any relay chain is being served to tell pocket-core to boost itself.

The server running pocket-core is the "poller", meaning it polls prometheus output. In order for publushed messages
to reach all servers, their subscriber addresses must be listed in the config item `publish_to_endpoints`:

```yaml
publish_to_endpoints:
  - 10.0.0.2:5555
  - 10.0.0.3:5555
  - etc.. 
```

On all servers (including the pocket node), the `subscriber_address` configuration must be provided in order for that
server to receive messages and for pocket-autonice to boost pocket or relay chain processes.

```yaml
subscriber_address: 10.0.0.2:5555
```

That's it. All other configuration values would only need to be changed for an unusual setup. See the
documentation in the config_example.yaml file for information on all configuration options.

### Running pocket-autonice

Try running with the `--dry-run` flag for a few relay sessions until you get the hang of what pocket-autonice
will be doing to your processes.

Pocket-autonice must be run as `root` or under a process that has passwordless `sudo`. This is because the `renice` 
command is a privileged command. If using `sudo` set the `run_with_sudo` config value.

On the server running pocket-core:

```shell
pocket-autonice --with-poller
```

On all other servers:

```shell
pocket-autonice
```

## Testing

On a macOS device, install the ZeroMQ libraries.

```sh
brew install libsodium zeromq czmq
```

Ensure that Go has the ability to compile C modules.

```sh
go env CGO_ENABLED
-> 1
```
