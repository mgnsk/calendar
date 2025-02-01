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

### Docker Compose deployment

The Docker image is published at https://github.com/mgnsk/calendar/pkgs/container/calendar

Describes an easy deployment pattern I personally use elsewhere.
Choose whichever system you're familiar with.

You need a Docker installation with the [Compose plugin](https://docs.docker.com/compose/install/linux/).

Copy `docker-compose.example.yml` to `/etc/docker/compose/calendar/docker-compose.yml`
and `.env` to `/etc/docker/compose/calendar/.env`.

Edit the environment variables. Make sure to set a new 32-byte SESSION_SECRET.

You may also set the environment variables directly in `docker-compose.yml`.
In this case, the `.env` file is not needed. Remove the `env_file` setting from the compose file.

Copy `docker-compose@.service` to `/etc/systemd/system/`.

Enable the service: `$ systemctl enable --now docker-compose@calendar`.

Docker Compose automatically creates a database volume for the service. Do not delete the volume or all data will be lost!

To update the service:

- `$ cd /etc/docker/compose/calendar`
- `$ docker compose pull`
- `$ systemctl restart docker-compose@calendar`

TODO: automatic update
TODO: describe HTTPS reverse proxy (most probably Caddy) and domain name config.
