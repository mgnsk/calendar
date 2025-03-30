FROM node:23-bookworm@sha256:990d0ab35ae15d8a322ee1eeaf4f7cf14e367d3d0ee2f472704b7b3df4c9e7c1 AS assets

WORKDIR /build

COPY package.json package-lock.json ./
ENV NODE_ENV=production
RUN npm ci

COPY tailwind.css ./
COPY html ./html
RUN npx @tailwindcss/cli -i tailwind.css -o app.css --minify


FROM golang:1.24-bookworm@sha256:fa1a01d362a7b9df68b021d59a124d28cae6d99ebd1a876e3557c4dd092f1b1d AS build

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
COPY --from=assets /build/app.css ./app.css
COPY --from=assets /build/node_modules ./node_modules

ENV CGO_ENABLED=0
RUN go build -trimpath -tags timetzdata,strictdist -o calendar ./cmd/calendar


FROM scratch

COPY --from=build /build/calendar /
