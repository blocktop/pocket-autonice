# pocket-autonice



## Dependencies

### ZeroMQ required development libraries and build tools

Install zeromq dependencies from the system package manager. For example on Ubuntu 20.04:

```shell
apt install -y libtool pkg-config gcc g++ make libsodium-dev libczmq-dev
```

## Installation

Clone the repo and install dependencies

```sh
git clone https://github.com/blocktop/pocket-autonice.git 
cd pocket-autonice
npm install
```

## Usage

There are two processes in pocket-autorenice, a client and a server. Run the client on all servers as
a privileged user. This user will need to be able to execute the `renice` program.

```sh
npm start client
```

Run the server as an unprivileged user on the pocket server. The nginx mirror configuration below is required
to make any messages flow to the server.

```sh
npm start server
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
