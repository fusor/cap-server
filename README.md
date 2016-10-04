<a name="overview"></a>
# Overview
cap-server is the [link http://golang.org  Golang]  backend for the CAP project. It's has a restful API.

<a name="usage"></a>
# Usage
<a name="installation"></a>
# Installation
If you just want to try out the server, you can use ```go get``` to install.

```shell
go get github.com/fusor/cap-server
```
This will install cap-server in ```$GOPATH/bin```. Then run ```cap-server``` and it will start to listen on ```localhost:3001```.

# Development
You can work on cap-server in one of two ways. Straight on your box or using our [link http://www.vagrant.com Vagrant] environment: [link http://github.com/fusor/cap/ Cap project]

## straight from the source repo
```shell
git clone https://github.com/fusor/cap-server.git
cd cap-server
go get . # get dependencies
go build # build binary
./cap-server # listens on port 3001
```

## using vagrant
[link https://github.com/fusor/cap/blob/master/README.md Vagrant environment]

### Golang primer
I typically checkout my go projects at the same level as the ```bin,pkg,src```.
```shell
mkdir -p $HOME/go/{bin,pkg,src}
export GOPATH=$HOME/go
cd $HOME/go
git clone https://github.com/fusor/cap-server.git
```
