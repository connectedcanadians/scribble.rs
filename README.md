<h1 align="center">Scribble.rs</h1>

<p align="center">
  <a href="https://github.com/scribble-rs/scribble.rs/actions">
    <img src="https://github.com/scribble-rs/scribble.rs/workflows/scribble-rs/badge.svg">
  </a>
  <a href="https://codecov.io/gh/scribble-rs/scribble.rs">
    <img src="https://codecov.io/gh/scribble-rs/scribble.rs/branch/master/graph/badge.svg">
  </a>
  <a href="https://discord.gg/3sntyCv">
    <img src="https://img.shields.io/discord/693433417395732531.svg?logo=discord">
  </a>
</p>

This project is intended to be a clone of the web-based drawing game
[skribbl.io](https://skribbl.io). In my opinion skribbl.io has several
usability issues, which I'll try addressing in this project.

Even though there is an official instance at
[scribble.rs](http://scribble.rs), you can still host your own instance.

The site will not display any ads or share any data with third parties.

## Building / Running

Run the following to build the application:

Install Go, Yarn.

```shell
# after cloning this repository...
go mod tidy

# bundle webpack
make bundle 

# build the go binary
make build

# run tests
make test

# run locally
make up

# dockerize
make dockerize

# push docker image to docker repository
make push

```

This will produce a portable binary called `scribblers`. The binary doesn't
have any dependencies and should run on every system as long as it has the
same architecture and OS family as the system it was compiled on.

# Configuration via Environment Variables

`HTTP_PORT` defaults to `8080`

`REDIS_ADDR` defaults to `127.0.0.1:6379`

`AGORA_APP_ID` defaults to `89f97462e28540e68a6a90760c9ca113` ... get one at https://www.agora.io/en/

`AGORA_CERT`


## Contributing

There are many ways you can contribute:

* Update / Add documentation in the wiki of the GitHub repository
* Extend this README
* Create issues
* Solve issues by creating Pull Requests
* Tell your friends about the project
* Curating the word lists

## Connected Canadians setups
1. Install go
1. Install CompileDaemon `go get github.com/githubnemo/CompileDaemon`
1. `./run.sh`
1. When you make changes to any file, refresh your browser tab http://localhost:8080/

## Credits

* Favicon - [Fredy Sujono](https://www.iconfinder.com/freud)
* Rubber Icon - Made by [Pixel Buddha](https://www.flaticon.com/authors/pixel-buddha) from [flaticon.com](https://flaticon.com)
* Fill Bucket Icon - Made by [inipagistudio](https://www.flaticon.com/authors/inipagistudio) from [flaticon.com](https://flaticon.com)
