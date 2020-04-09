
win:
	go build -o ./bin/orderServer.exe ./src/cmd/orderServer.go
linux:
	go build -o ./bin/orderServer ./src/cmd/orderServer.go