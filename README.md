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
git clone https://github.com/scribble-rs/scribble.rs.git
cd scribble.rs
go mod tidy
make bundle build
./start.sh
```

This will produce a portable binary called `scribblers`. The binary doesn't
have any dependencies and should run on every system as long as it has the
same architecture and OS family as the system it was compiled on.

The default port will be `8080`, and can be configured with the `-portHTTP` flag.

The agora key is provided by environment variable `AGORA_CERT`

It should run on any system that go supports as a compilation target.

This application uses go modules, therefore you need to make sure that you
have go version `1.13` or higher.

## Docker

Alternatively there's a docker container (which is out of date, ds0nt is guessing):

```shell
docker pull biosmarcel/scribble.rs
```

It uses port `80`.

## Contributing

There are many ways you can contribute:

- Update / Add documentation in the wiki of the GitHub repository
- Extend this README
- Create issues
- Solve issues by creating Pull Requests
- Tell your friends about the project
- Curating the word lists

## Connected Canadians setups

1. Install go
1. Install CompileDaemon `go get github.com/githubnemo/CompileDaemon`
1. `./run.sh`
1. When you make changes to any file, refresh your browser tab http://localhost:8080/

## Building and deploying on cloud builds from CLI

1. Install gcloud cli tool https://cloud.google.com/sdk/docs/install
1. Authenticate and login `gcloud auth login`
1. Set the project `gcloud config set project socialgaming`
1. `gcloud app deploy` to deploy the application

### Build and upload to a google cloud storage

1. `gcloud builds submit`
1. `gcloud config set run/platform gke`
1. `gcloud config set compute/zone us-central1-c`
1. `gcloud services enable container.googleapis.com containerregistry.googleapis.com cloudbuild.googleapis.com`
1. `gcloud config set run/cluster socialgame-cluster`
1. `gcloud config set run/cluster_location us-central1-c`

## Credits

- Favicon - [Fredy Sujono](https://www.iconfinder.com/freud)
- Rubber Icon - Made by [Pixel Buddha](https://www.flaticon.com/authors/pixel-buddha) from [flaticon.com](https://flaticon.com)
- Fill Bucket Icon - Made by [inipagistudio](https://www.flaticon.com/authors/inipagistudio) from [flaticon.com](https://flaticon.com)
