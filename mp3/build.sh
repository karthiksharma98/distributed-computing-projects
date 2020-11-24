#go build -o master main.go master.go monitor.go net.go util.go logs.go detector.go
#go build -o client main.go monitor.go net.go util.go logs.go detector.go client.go

go build -o main
cd wordcount && go build -o ../wc wordcount.go