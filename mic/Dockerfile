FROM arm32v7/golang:latest
RUN apt-get update && apt-get install -y portaudio19-dev
RUN go get github.com/gordonklaus/portaudio
COPY ./ $HOME/go/src/github.com/armsnyder/somethings-funny/mic/
RUN go build github.com/armsnyder/somethings-funny/mic

FROM arm32v7/debian:latest
RUN apt-get update && apt-get install -y libportaudio2
COPY --from=0 /go/mic /usr/local/bin/mic
CMD ["/usr/local/bin/mic"]
