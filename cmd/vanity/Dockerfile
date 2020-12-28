FROM golang:1.15 AS build
RUN mkdir /app
WORKDIR /app
RUN go mod init vanity
COPY main.go .
RUN CGO_ENABLED=0 go build -o=vanity

FROM busybox
RUN mkdir /app
WORKDIR /app
COPY --from=build /app/vanity /app/vanity
ENTRYPOINT [ "/app/vanity" ]
