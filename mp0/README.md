# Logly
Your friendly neighborhood distributed logging microservice from _Covfefe! Inc_.

## Installation
```
# Install Go
wget https://dl.google.com/go/go1.13.src.tar.gz
tar -C /usr/local -xzf go$VERSION.$OS-$ARCH.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

```
# Install Logly
mkdir -p ~/go
cd go/src
git clone git@gitlab.engr.illinois.edu:jjouett2/ece428.git
cd ece428
./bootstrap.sh
source ~/.bashrc
go build -o logly ece428/src/main
```

## Usage Examples
```
./logly --help
CONFIG=config.json ./logly -server &> serverlog.log &
CONFIG=config.json ./logly -client --expression="(A-Z)*" &> output.log
```

## Testing
```sh
go test ./...
go vet ./... # Check for program correctness
```
