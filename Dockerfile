# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM crazymax/goxx:latest AS base
FROM golang:1.19-bullseye
WORKDIR /server

COPY /GoRemoteControl_Server /GoRemoteControl_Server
COPY ./viewer.html /server/viewer.html

EXPOSE 1054/tcp
EXPOSE 1053/udp

CMD [ "/GoRemoteControl_Server"]