# RPM
Project IDA Remote Power Monitoring Application

### Install GO
__On FBSD system as user *nrts*:__

* mkdir -p build/go
* cd build/go
* curl -L -o go1.18.1.freebsd-386.tar.gz https://go.dev/dl/go1.18.1.freebsd-386.tar.gz
_Golang v18.1 current as of this writing_

__as *root*:__

* cd ~nrts/build/go
* tar -C /usr/local -xzf go1.18.1.freebsd-386.tar.gz

__as user *nrts*:__

* Check ~nrts/.pathrc for `set MyPath = ($MyPath /usr/local/go/bin)` and if not there, add before the line `set path = ($MyPath $path)`.
* log out (`exit`)
* log back in
* Check go installation by viewing the go version: `go version`. You should see: *go version go1.18.1 freebsd/386*

### Get RPM Source from a Release on Github and Build
You must use a Personal Access Token from your GitHub account
* cd ~/build
* set TOKEN = "_your token goes here in quotes"
#### _This example downloads and builds release version 1.2 created on Github_
* curl -sL --header "Authorization: token $TOKEN" --header 'Accept: application/octet-stream' https://github.com/ProjectIDA/rpm/archive/refs/tags/v1.2.tar.gz -o rpm.v1.2.tar.gz
* tar xvf rpm.v1.2.tar.gz
* cd rpm-1.2
* go build
* _if ~/bin/rpm exists and you are upgrading, then_ `chmod 755 ~/bin/rpm`
* cp rpm ~/bin/
* cp rpm.toml ~/etc
* cd ~/etc
* Edit ~/etc/rpm.toml and set correct station code (uppercase) at top of file

### TODO
*better log msgs on signals and on shutdown
*static OID log msgs should come before "first packet received" msg