FROM arm32v7/golang:latest
RUN go get github.com/aws/aws-sdk-go
RUN go get github.com/orcaman/writerseeker
RUN go get github.com/mjibson/go-dsp/fft
COPY ./ $HOME/go/src/github.com/armsnyder/somethings-funny/sampler/
RUN go build github.com/armsnyder/somethings-funny/sampler

FROM arm32v7/debian:latest
RUN apt-get update && apt-get install -y ca-certificates
COPY --from=0 /go/sampler /usr/local/bin/sampler
CMD ["/usr/local/bin/sampler"]
