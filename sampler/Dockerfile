FROM golang:latest
COPY ./ /go/src/github.com/armsnyder/somethings-funny/sampler/
RUN env GOOS=linux GOARCH=arm GOARM=7 go build github.com/armsnyder/somethings-funny/sampler
FROM arm32v7/busybox:latest
COPY --from=0 /go/sampler /usr/local/bin/sampler
CMD ["/usr/local/bin/sampler"]
