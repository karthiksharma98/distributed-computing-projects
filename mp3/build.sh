#go build -o master main.go master.go monitor.go net.go util.go logs.go detector.go
#go build -o client main.go monitor.go net.go util.go logs.go detector.go client.go

go build -o main
cd maple && go build -o ../condorcet_maple_1 mapler.go condorcet_1.go
cd maple && go build -o ../condorcet_maple_1 mapler.go condorcet_2.go
#cd ../juice && go build -o ../wcjuice
