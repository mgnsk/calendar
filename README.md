## Simple calendar app

### Building and running with docker

- `$ docker compose build`
- `$ docker compose up`

### Local development

Dependencies:

- [Go](https://go.dev/)
- [Node.js](https://nodejs.org/en)

Setup tools and run development environment:

- `$ make setup`
- `$ make dev`

### General info

The service binds itself to `:8080` by default. To change it, specify the `--addr` flag.

The service minimally needs a persistent volume to store the SQLite database.
Just mount the directory and specify via the `--database-dir` flag.

It automatically runs needed database migrations on startup.

### Docker Compose deployment

The Docker image is published at https://github.com/mgnsk/calendar/pkgs/container/calendar

Describes an easy deployment pattern I personally use elsewhere.
Choose whichever system you're familiar with.

You need a Docker installation with the [Compose plugin](https://docs.docker.com/compose/install/linux/).

Copy `docker-compose.yml` and `configuration.example.yml` to `/etc/docker/compose/calendar/`.
Edit the configuration and rename to `configuration.yml`.
Copy `docker-compose@.service` to `/etc/systemd/system/`.

Enable the service: `$ systemctl enable --now docker-compose@calendar`.

Docker Compose automatically creates a database volume for the service. Do not delete the volume or all data will be lost!

To update the service:

- `$ cd /etc/docker/compose/calendar`
- `$ docker compose pull`
- `$ systemctl restart docker-compose@calendar`

TODO: implement and describe the configuration file
TODO: automatic update
TODO: describe HTTPS reverse proxy (most probably Caddy) and domain name config.
