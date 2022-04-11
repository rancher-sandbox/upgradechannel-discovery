FROM golang:1.17-alpine as build
ENV CGO_ENABLED=0
WORKDIR /src
COPY go.mod go.sum /src/
RUN go mod download
COPY main.go /src/
COPY pkg /src/pkg
RUN go build -ldflags "-extldflags -static -s" -o /usr/bin/upgradechannel-discovery

FROM scratch
COPY --from=build /usr/bin/upgradechannel-discovery /usr/bin/upgradechannel-discovery
ENTRYPOINT "/usr/bin/upgradechannel-discovery"