FROM armsnyder/arm32v7-golang-cross:latest
RUN apt-get update && apt-get install -y portaudio19-dev
RUN go get github.com/gordonklaus/portaudio
RUN go get github.com/aws/aws-sdk-go
RUN go get github.com/orcaman/writerseeker
RUN go get github.com/mjibson/go-dsp/fft
COPY ./ $HOME/go/src/github.com/armsnyder/somethings-funny/sampler/
CMD ["sh", "-c", "go build github.com/armsnyder/somethings-funny/sampler && ./sampler"]
