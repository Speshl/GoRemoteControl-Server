# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM crazymax/goxx:latest AS base

ENV OUTPUT="simple-cam"
ENV CGO_ENABLED=1
WORKDIR /src

FROM base AS build
ARG TARGETPLATFORM
RUN --mount=type=cache,sharing=private,target=/var/cache/apt \
  --mount=type=cache,sharing=private,target=/var/lib/apt/lists \
  goxx-apt-get install -y binutils gcc g++ pkg-config
RUN --mount=type=bind,source=. \
  --mount=type=cache,target=/root/.cache \
  --mount=type=cache,target=/go/pkg/mod \
  goxx-go build -o /out/${OUTPUT} .

FROM scratch AS artifact
COPY --from=build /out /


## Build with the following command
# docker build --platform "linux/arm/v6" --output "./build" .


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