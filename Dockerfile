FROM golang:1.10 AS builder

ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

WORKDIR $GOPATH/src/gocode

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only

COPY . ./
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix nocgo -o /app .

FROM scratch
COPY --from=builder /app ./
EXPOSE 80
ENTRYPOINT ["./app"]