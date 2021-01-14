# rpm
Project IDA Remote Power Monitoring Application

### Install GO
as user *nrts*:
mkdir -p build/go
cd build/go
curl -L -o go1.15.6.freebsd-amd64.tar.gz https://golang.org/dl/go1.15.6.freebsd-amd64.tar.gz

as *root*:
tar -C /usr/local -xzf go1.15.6.freebsd-amd64.tar.gz

as user *nrts*
add to .pathrc
`set path = ($path /usr/local/go/bin)`

cd ~/build
git clone https://danauerbach@github.com/projectida/rpm.git
cd rpm
go get
go build
cp rpm/toml ~/etc
Edit rpm.toml with correct station code

### TODO
better log msgs on signals and on shutdown
set user to nrts
static OID log msgs should come before "first packet received" msg