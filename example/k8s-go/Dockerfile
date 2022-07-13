FROM golang:1.18.3-bullseye as build
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server

###

FROM gcr.io/distroless/base-debian11
COPY --from=build /app/server /server
CMD ["/server"]