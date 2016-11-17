# Tunneling tool for Nitrous Solo

solo-tunnel opens an SSH tunnel to your [Nitrous Solo](https://github.com/nitrous-io/nitrous-solo) container. This allows you to use the Nitrous Solo Web IDE securely and also access any other ports (e.g. for previewing your web application).

## Usage

`solo-tunnel start <SSH URL> [FLAGS]`

Example:

`solo-tunnel start ssh://nitrous@1.2.3.4:22 -i ~/.ssh/id_rsa -p 8000:6877 -p 3000:3000`

An explanation of the command's arguments:

* `ssh://nitrous@1.2.3.4:22`: This is URL to the machine that your Nitrous Solo container is running on:
  * `nitrous`: The username of an account on the machine.
  * `1.2.3.4`: IP address of the machine.
  * `22`: Port that sshd is running on.
* `-i ~/.ssh/id_rsa`: Path to SSH private key for the user you're trying to access the machine as.
* `-p 8000:6877`: Forwards port 8000 to 6877. Nitrous Solo usually runs on port 6877, so you can access it from http://localhost:8000/.
* `-p 3000:3000`: Forwards port 3000 to 3000.


## Development

### Dependencies

We use [govendor](https://github.com/kardianos/govendor) to manage dependencies.

### Run tests

`script/test`

### Building the binaries

`script/build`
