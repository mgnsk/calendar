# Simple calendar app

## Local development

- `make setup` - Set up the development sandbox.
- `make dev` - Start the development sandbox.
- `make stop` - Stop the development sandbox.

The calendar application can be visited on `https://calendar.localhost`.

## Production installation

The Docker image is published at [https://github.com/mgnsk/calendar/pkgs/container/calendar](https://github.com/mgnsk/calendar/pkgs/container/calendar).

### Clone the repository

As root:

- `mkdir /opt/mgnsk/calendar`
- `cd /opt/mgnsk/calendar`
- `git clone https://github.com/mgnsk/calendar.git .`

### Create the `.env` file with your public domain name

```env
HOSTNAME="my-domain-name.com"
```

### Install and enable the service

As root:

`make install`

This will:

- Create a Docker volume for the database if not already exists.
- Symlink and reload the systemd service file.
- Enable and start the service.
