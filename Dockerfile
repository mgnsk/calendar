FROM node:25-bookworm@sha256:6d362f0df70431417ef79c30e47c0515ea9066d8be8011e859c6c3575514a027 AS assets

WORKDIR /build

COPY package.json package-lock.json ./
ENV NODE_ENV=production
RUN npm ci

COPY tailwind.css ./
COPY html ./html
RUN npx @tailwindcss/cli -i tailwind.css -o app.css --minify


FROM golang:1.25.5-bookworm@sha256:2c7c65601b020ee79db4c1a32ebee0bf3d6b298969ec683e24fcbea29305f10e AS build

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
COPY --from=assets /build/app.css ./app.css
COPY --from=assets /build/node_modules ./node_modules

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GOCACHE=/go/cache

RUN --mount=type=cache,target=/go/cache/ go build -trimpath -tags timetzdata,strictdist -o calendar ./cmd/calendar

RUN mkdir /database /cache


FROM scratch

COPY --from=build /build/calendar /bin/calendar
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build --chown=65534:65534 /cache/ /cache/
COPY --from=build --chown=65534:65534 /database/ /database/

USER 65534:65534
