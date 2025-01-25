

#### Setup easyjson
```bash
go get -u github.com/mailru/easyjson/...
easyjson -all internal/types/sol.go
```


### Setup Timescale Server
```bash
https://docs.docker.com/engine/install/ubuntu/

sudo ufw allow from <IP_ADDRESS> to any port 5432
sudo ufw enable

docker pull timescale/timescaledb-ha:pg17

docker run -d \
  --name timescaledb \
  -p 5432:5432 \
  -e POSTGRES_PASSWORD=<PASSWORD> \
  -v timescaledb-data:/home/postgres/pgdata/data \
  timescale/timescaledb-ha:pg17
```

