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

Clone this repo into `/etc/docker/compose/calendar/`.
Copy the `docker-compose@.service` file to `/etc/systemd/system/` directory.

You need a Docker installation with the [Compose plugin](https://docs.docker.com/compose/install/linux/).

Enable the service: `$ systemctl enable --now docker-compose@calendar`.

TODO: describe HTTPS reverse proxy (most probably Caddy) and domain name config.
