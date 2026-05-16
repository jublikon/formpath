FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bin/formpath ./cmd/server

FROM alpine:3.21

WORKDIR /app
COPY --from=build /bin/formpath /app/formpath
COPY migrations /app/migrations

EXPOSE 8080
CMD ["/app/formpath"]
