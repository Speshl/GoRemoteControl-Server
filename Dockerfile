# syntax=docker/dockerfile:1

FROM golang:1.19-bullseye
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY / ./
RUN ls -la ./*.go

RUN apt-get update -qq  
#RUN apt-get install -y build-essential
RUN apt-get install -y v4l-utils

#RUN go get github.com/vladimirvivien/go4vl/v4l2

RUN go build -o /GoRemoteControl-Server

EXPOSE 1054/tcp
EXPOSE 1053/udp

CMD [ "/GoRemoteControl-Server"]


#****************** To Build and Deploy to DockerHub**************************

#docker login
#docker buildx build --platform=linux/arm64,linux/amd64 -t speshl/goremotecontrol-server:latest --push .

#***************************************************************************


#****************** Build and Test Local**************************
#docker build --platform=linux/amd64 -t speshl/goremotecontrol-server:latest .

#docker run  -d -p 1054:1054 -p 1053:1053 speshl/goremotecontrol-server:latest



#*******************Push to docker Hub OR Save/Load locally ******************************
#docker push speshl/thermo_status_server:tagname

#docker save --output thermo_status_server.tar thermo_status_server
#docker load --input thermo_status_server.tar