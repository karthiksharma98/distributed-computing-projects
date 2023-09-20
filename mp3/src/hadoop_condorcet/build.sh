cd reducer && go build -o ../reduce1 reduce.go reduce1.go
cd ../reducer && go build -o ../reduce2 reduce.go reduce2.go
cd ../mapper && go build -o ../map1 map.go map1.go
cd ../mapper && go build -o ../map2 map.go map2.go