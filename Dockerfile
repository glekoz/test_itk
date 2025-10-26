FROM golang:1.25.3-alpine3.22 AS build
WORKDIR /src

COPY ./internal/repository/migrations ./migrations
COPY config.env /etc/itkapp/config.env
COPY . .
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.26.0
RUN go build -o /usr/bin/itkapp ./cmd/itkapp/main.go

FROM ubuntu:24.04

COPY --from=build /go/bin/goose /usr/local/bin/goose
COPY --from=build /usr/bin/itkapp /usr/bin/itkapp  
COPY --from=build /src/migrations /etc/itkapp/migrations
COPY --from=build /etc/itkapp/ /etc/itkapp/