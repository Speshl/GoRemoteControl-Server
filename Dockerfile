# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM crazymax/goxx:latest AS base
FROM golang:1.19-bullseye
WORKDIR /server

COPY /GoRemoteControl_Server /GoRemoteControl_Server
COPY ./viewer.html /server/viewer.html

EXPOSE 1054/tcp
EXPOSE 1053/udp

CMD [ "/GoRemoteControl_Server"]


#scp dockerbox@192.168.1.41:~/go/src/github.com/Speshl/GoRemoteControl_Server/GoRemoteControl_Server .

#scp dockerbox@192.168.1.41:~/scripts/pi_compose.yml .
#scp pi_compose.yml dockerbox@192.168.1.41:~/scripts/pi_compose.yml


## Build with the following command
# docker build --platform "linux/arm/v6" --output "./build" .


#****************** To Build and Deploy to DockerHub**************************

#docker login
#docker buildx build --platform=linux/arm64,linux/amd64 -t speshl/goremotecontrol-server:latest --push .

#***************************************************************************


#****************** Build and Test Local**************************
#docker build --platform=linux/amd64 -t speshl/goremotecontrol-server:latest .

#docker run -d -p 1054:1054 -p 1053:1053 speshl/goremotecontrol-server:latest



#*******************Push to docker Hub OR Save/Load locally ******************************
#docker push speshl/thermo_status_server:tagname

#docker save --output thermo_status_server.tar thermo_status_server
#docker load --input thermo_status_server.tar