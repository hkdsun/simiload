default: run

linux_server:
		CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o server server.go

builddocker: linux_server
		docker build -t hkdsun/simiload .

run: builddocker
	docker-compose up
