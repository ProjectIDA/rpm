# RPM
Project IDA Remote Power Monitoring Application

### Install GO
__On FBSD system as user *nrts*:__

* mkdir -p build/go
* cd build/go
* curl -L -o go1.16.3.freebsd-amd64.tar.gz https://golang.org/dl/go1.16.3.freebsd-amd64.tar.gz
_Golang v16.3 current as of this writing_

__as *root*:__

* tar -C /usr/local -xzf go1.16.3.freebsd-amd64.tar.gz

__as user *nrts*:__

* add to .pathrc: `set path = ($path /usr/local/go/bin)`

### Get RPM Source from a Release on Github and Build
You must use a Personal Access Token from your GitHub account
* cd ~/build
* set TOKEN = "_your token goes here in quotes"
#### _this example downloads adn builds release version 1.2 created on Github_
* curl -sL --header "Authorization: token $TOKEN" --header 'Accept: application/octet-stream' https://github.com/ProjectIDA/rpm/archive/refs/tags/v1.2.tar.gz -o rpm.v1.2.tar.gz
* tar xvf rpm.v1.2.tar.gz
* cd rpm-1.2
* go build
* chmod 755 ~/bin/rpm _(only if rpm already exists in ~/bin)_
* cp rpm ~/bin/
* cp rpm.toml ~/etc
* Edit ~/etc/rpm.toml and set correct station code at top of file

### TODO
*better log msgs on signals and on shutdown
*static OID log msgs should come before "first packet received" msg