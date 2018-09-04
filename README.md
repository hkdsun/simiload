# Simiload

Simulation of an Overload Protection System

`make` compiles `server.go` and runs a docker-compose cluster with:

- Simulation server: `localhost:8080`
- Metrics server: `localhost:3000`
- Load generator: `ruby generate.rb` (nicer one being developed in `generate.go`)

![image](https://user-images.githubusercontent.com/6955854/45006491-39549f80-afc7-11e8-8225-0cadca0cee56.png)
