FROM node:23-bookworm AS assets

WORKDIR /build

COPY package.json package-lock.json ./
RUN npm ci

COPY tailwind.config.js tailwind.css ./
COPY internal/html ./internal/html
RUN npx tailwindcss -i tailwind.css -o ./internal/dist/app.css --minify \
    && cp ./node_modules/htmx.org/dist/htmx.min.js ./internal/dist/htmx.min.js \
    && cp ./node_modules/mark.js/dist/mark.min.js ./internal/dist/mark.min.js


FROM golang:1.23-bookworm AS build

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
COPY --from=assets /build/internal/dist /build/internal/dist

ENV CGO_ENABLED=0
RUN go build -trimpath -tags timetzdata,strictdist -o calendar ./cmd/calendar


FROM gcr.io/distroless/base-debian12

COPY --from=build /build/calendar /
