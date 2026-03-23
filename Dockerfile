FROM --platform=$BUILDPLATFORM node:24-alpine3.23 AS nodebuild

RUN apk add make

COPY frontend src
WORKDIR /src

RUN make build && \
    tar -C /src/dist/ed-survey-tools/browser -cvf /src/webroot.tar .

FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.25.7-alpine3.23 AS gobuild

ARG TARGETOS
ARG TARGETARCH

RUN apk add make file

WORKDIR /src
COPY . .
COPY --from=nodebuild /src/webroot.tar /src/http/webroot.tar

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} make build && \
    file /src/dist/edst

FROM scratch

WORKDIR /bin
COPY --from=gobuild /src/dist/edst /bin/

ENTRYPOINT ["/bin/edst"]


