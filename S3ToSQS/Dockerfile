
FROM golang:1.12



RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh


ENV GOROOT=/usr/local/go
ENV GOBIN=$GOROOT/bin
ENV GOPATH=/go
ENV PATH=$PATH:$GOBIN

EXPOSE 80 8000

#RUN go dep ensure -update

RUN go get github.com/canthefason/go-watcher

RUN go install github.com/canthefason/go-watcher/cmd/watcher

RUN apt update

RUN apt install python3-pip -y

RUN apt install zip -y

RUN pip3 install awscli --upgrade --user