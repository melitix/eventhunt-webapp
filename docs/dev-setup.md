# Setup local dev environment

## Dependencies

- [Go 1.22+](https://go.dev/)
- PostgreSQL v16.3 - this can be installed directly or via Docker
- [Tern](https://github.com/jackc/tern) - (DB migrations)
- [GoTestSum](https://github.com/gotestyourself/gotestsum) (optional)
- Docker (optional)


## Get code

```bash
git clone https://github.com/eventhunt-org/webapp.git
cd webapp
```


## Create/Start PostgreSQL instance

### If you are using Docker:

First time run:

```bash
docker run --name=eventhunt-db -e POSTGRES_USER=app -e POSTGRES_PASSWORD=APass -e POSTGRES_DB=app -p 9001:5432 -d postgis/postgis:16-3.4
```

You can swap out the password `APass` if you'd like, but as this is a local development environment, this isn't important.
Do not use a bad password like this for production.

After the first time, you can start the container with:

```bash
docker start eventhunt-db
```


## Run migrations

```bash
cd ./migrations
tern migrate --port=9001
```


## Seed geo data

```bash
cd ../seeder
go run . countries --filter=US
go run . spr --filter=US
go run . tz --filter=US
go run . cities --filter=US
```


## Seed environment variables

```bash
cd ..
cp example.env .env
```

At the very least, two envars should be set:

- **AUTH_SESSION_KEY** - A secret value.
- **APP_MAP_KEY** -  The Google Maps API key.


## Run the application

```bash
cd ./webapp
go run .
```


