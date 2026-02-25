FROM node:25-bookworm@sha256:ee8da47d6d7cdce02f246d7363dad4f687340854ab161816ce1bc3ce5ff1069b AS assets

WORKDIR /build

COPY package.json package-lock.json ./
ENV NODE_ENV=production
RUN npm ci

COPY tailwind.css ./
COPY html ./html
RUN npx @tailwindcss/cli -i tailwind.css -o app.css --minify


FROM golang:1.26.0-bookworm@sha256:878786d2372335058513ce271a4a5c2a1d3e6839b105144183464a05558b101c AS build

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
