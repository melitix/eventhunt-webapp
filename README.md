# EventHunt
[![CI Status](https://dl.circleci.com/status-badge/img/gh/eventhunt-org/webapp/tree/trunk.svg?style=shield)](https://dl.circleci.com/status-badge/redirect/gh/eventhunt-badge/webapp/tree/trunk)
[![Codecov](https://codecov.io/gh/eventhunt-badge/webapp/graph/badge.svg)](https://codecov.io/gh/eventhunt-badge/webapp)
![Software License](https://img.shields.io/badge/license-proprietary-lightgrey.svg)

A local meet up group webapp.


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
docker run --name=eventhunt-db -e POSTGRES_USER=app -e POSTGRES_PASSWORD=AppleSucks -e POSTGRES_DB=app -p 9001:5432 -d postgis/postgis:16-3.4
```


## Environment Variables

Check the `example.env` file for the latest variables and default values.

`RAG_THEME_ROOT`: This environment variable tells the webapp where to find the theme files. The default value is "." which means to look in the same directory that the binary is running from. This is going to be ideal when running via Go (go run .) however usually not what you want when running the binary in a real server. The path should have a trailing forward slash.

`RAG_EMAIL_USER`: The SMTP username.  
`RAG_EMAIL_PWD`: The SMTP password.  
`RAG_EMAIL_HOST`: The SMTP hostname, not including the port.

`RAG_DB_USER`: The MariaDB username.  
`RAG_DB_PWD`: The MariaDB password.  
`RAG_DB_NAME`: The MariaDB database name.

`RAG_ENVIRONMENT`: The environment type the app is running in. The default value is "development". Other expected values is "production" and "staging".

`RAG_WEATHER_KEY`: The API key for OpenWeatherMap.org.


## Running

First the database needs to be started.
Then the app can be run:

```
docker start ra-garage_db
cd ./webapp
go run .
```


## Add new models

./scripts/import-models.sh <F_NAME>

where `<F_NAME>` is the filename of the .zip without the extension

mysqldump -h127.0.0.1 -uroot -pAppleSucks --lock-tables ra-garage_dev vehicle_models > sql/add-vehicle-models.sql


## Development vs Production

**envars** - For environment variables, in a dev environment they should be set in your local shell.
For `fish`, this would be `~/.config/fish/config.fish`.
For production, the SystemD service file loads `prod.env` as an environment file.


## Info

Times stored in the database should be assumed to be UTC.


## Setup

seeder countries --filter=US
seeder spr --filter=US
seeder tz --filter=US
seeder cities --filter=US
