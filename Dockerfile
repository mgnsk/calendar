FROM node:23-bookworm AS assets

WORKDIR /build

COPY package.json package-lock.json ./
RUN npm ci

COPY tailwind.config.js tailwind.css ./
COPY internal/html ./internal/html
RUN npx tailwindcss -i tailwind.css -o app.css --minify


FROM golang:1.23-bookworm AS build

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
COPY --from=assets /build/app.css ./
COPY --from=assets /build/node_modules ./

ENV CGO_ENABLED=0
RUN go build -trimpath -tags timetzdata,strictdist -o calendar ./cmd/calendar


FROM gcr.io/distroless/base-debian12

COPY --from=build /build/calendar /
