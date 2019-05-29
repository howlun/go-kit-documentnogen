# DOCNOGEN_BE
A Golang microservice that generate document number

> API Doc: https://documenter.getpostman.com/view/5502222/S1TR3z87

## Golang Installation
```
$ cd  /tmp
$ wget –c https://storage.googleapis.com/golang/go1.11.5.linux-amd64.tar.gz
$ sudo tar -C /usr/local -xvzf go1.11.5.linux-amd64.tar.gz
$ mkdir –p ~/go/bin
$ mkdir –p ~/go/src
$ mkdir –p ~/go/pkg
$ sudo nano ~/.profile
```
 Add in the following lines to the bottom of the file:
```
export PATH=$PATH:/usr/local/go/bin
export GOPATH="$HOME/go"
export GOBIN="$GOPATH/bin"
```
Save and exit

```
$ source ~/.profile
$ go version
$ go env
```

## Source Code Installation
```
$ cd /tmp
$ git clone https://github.com/howlun/go-kit-documentnogen.git
```
Type in Username and Password
```
$ cd go-kit-documentnogen/deployment/{dev or staging or prod}
$ bash deploy.sh
$ sudo systemctl status docnogen-api
```

## Steps to deploy to different environment
1. create **deploy.sh** and **docnogen-api.service** for {env} environment under the folder **go-kit-documentnogen/deployment/{env}**
2. configure the system to run with different options (Change **docnogen-api.service** file)
```
GLOBAL OPTIONS:
   --httpaddr value           Http Server Address (default: ":12000")
   --grpcaddr value           GRPC Server Address (default: ":13000")
   --mongoaddr value          Mongo DB Server Address (default: "localhost:27017")
   --mongodbname value        Mongo DB Name (default: "docnogen_v1")
   --mongoauthusername value  Mongo DB Auth Username
   --mongoauthpassword value  Mongo DB Auth Password
   --httplog value            HTTP log directory and filename (default: "log/http.log")
   --help, -h                 show help
   --version, -v              print the version
```
3. rebuild the source by
```
$ cd ~/go/src/github.com/howlun/go-kit-documentnogen/cmd/server/
$ go build
$ sudo systemctl restart docnogen-api
```