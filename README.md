# EventHunt
[![CI Status](https://dl.circleci.com/status-badge/img/gh/eventhunt-org/webapp/tree/trunk.svg?style=shield)](https://dl.circleci.com/status-badge/redirect/gh/eventhunt-badge/webapp/tree/trunk)
[![Codecov](https://codecov.io/gh/eventhunt-badge/webapp/graph/badge.svg)](https://codecov.io/gh/eventhunt-badge/webapp)
[![Software License](https://img.shields.io/badge/license-BSD3-blue.svg)](https://raw.githubusercontent.com/eventhunt-org/webapp/trunk/LICENSE)

Meet up locally with like-minded people.


## Dependencies

### Dev

- Go 1.22
- PostgreSQL v16.3
- GoTestSum
- Docker (optional)
- Tern - (DB migrations) https://github.com/jackc/tern

### Production

- PostgreSQL v16.3
- Tern - (DB migrations) https://github.com/jackc/tern


## Setup

### Without Docker Compose

```
docker run --name=eventhunt-db -e POSTGRES_USER=app -e POSTGRES_PASSWORD=APass -e POSTGRES_DB=app -p 9001:5432 -d postgis/postgis:16-3.4
```


## Environment Variables

Check the `example.env` file for the latest variables and default values.


## Running

First the database needs to be started.
Then the app can be run:

```
docker start eventhunt-db
cd ./webapp
go run .
```


## Info

Times stored in the database should be assumed to be UTC.


## Setup

```bash
cd ./seeder
go run . countries --filter=US
go run . spr --filter=US
go run . tz --filter=US
go run . cities --filter=US
```
