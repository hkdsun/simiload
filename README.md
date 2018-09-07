# Simiload

Simulation of different load shedding strategies. The high level architecture is as follows:

![image](https://user-images.githubusercontent.com/6955854/45006533-66a14d80-afc7-11e8-95a8-88c9c13546d9.png)

# Running

- `make` compiles `server.go` and runs a docker-compose cluster with the simulation server at: `localhost:8080`
- `make metrics` starts a metrics collection cluster, the Grafana cluster is at: `localhost:3000`
- To generate some load use: `go run -race generate.go -config flash_sale.json "http://localhost:8080"`

I used the logs for quick development:
![image](https://user-images.githubusercontent.com/6955854/45006491-39549f80-afc7-11e8-8225-0cadca0cee56.png)

And metrics for tuning:
<img width="1622" alt="screen shot 2018-09-07 at 7 07 06 am" src="https://user-images.githubusercontent.com/6955854/45215847-accb0b00-b26c-11e8-9e2c-da5e6890ad7f.png">
<img width="1624" alt="screen shot 2018-09-07 at 7 06 44 am" src="https://user-images.githubusercontent.com/6955854/45215849-accb0b00-b26c-11e8-8e7c-6950fe34e134.png">

