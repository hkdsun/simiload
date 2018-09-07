# Simiload

Simulation of an Overload Protection System

![image](https://user-images.githubusercontent.com/6955854/45006533-66a14d80-afc7-11e8-95a8-88c9c13546d9.png)

# Running

- `make` compiles `server.go` and runs a docker-compose cluster with the simulation server at: `localhost:8080`
- `make metrics` starts a metrics collection cluster, the Grafana cluster is at: `localhost:3000`
- To generate some load use: `go run -race generate.go -config flash_sale.json "http://localhost:8080"`

I used the logs for quick development and metrics for tuning:
![image](https://user-images.githubusercontent.com/6955854/45006491-39549f80-afc7-11e8-8225-0cadca0cee56.png)
