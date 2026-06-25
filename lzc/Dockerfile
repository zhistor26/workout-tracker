ARG NODE_IMAGE=registry.lazycat.cloud/u64111927/library/node:36ce4090eb3a433e
ARG GOLANG_IMAGE=registry.lazycat.cloud/u64111927/library/golang:d0fe39dc4e27a703
ARG ALPINE_IMAGE=registry.lazycat.cloud/u64111927/library/alpine:55ee87a30da74d0a
ARG SOURCE_REPO=https://github.com/zhistor26/workout-tracker.git
ARG SOURCE_REF=master

FROM ${NODE_IMAGE} AS frontend
ARG SOURCE_REPO
ARG SOURCE_REF

RUN sed -i 's|https://dl-cdn.alpinelinux.org/alpine|https://mirrors.tuna.tsinghua.edu.cn/alpine|g' /etc/apk/repositories \
    && apk --no-cache add ca-certificates git make

WORKDIR /src

RUN git clone --depth=1 --branch "${SOURCE_REF}" "${SOURCE_REPO}" .
RUN cd frontend \
    && npm config set registry https://registry.npmmirror.com \
    && npm ci

RUN make build-frontend

FROM ${GOLANG_IMAGE} AS backend
ARG BUILD_TIME
ARG GIT_COMMIT
ARG GIT_REF
ARG GIT_REF_NAME
ARG GIT_REF_TYPE

WORKDIR /app

COPY --from=frontend /src ./
RUN go env -w GOPROXY=https://goproxy.cn,direct \
    && go mod download

ENV CGO_ENABLED=0
RUN go mod vendor && go build \
    -ldflags "-X 'main.buildTime=${BUILD_TIME}' -X 'main.gitCommit=${GIT_COMMIT}' -X 'main.gitRef=${GIT_REF}' -X 'main.gitRefName=${GIT_REF_NAME}' -X 'main.gitRefType=${GIT_REF_TYPE}'" \
    -o /commands/ ./cmd/...

FROM ${ALPINE_IMAGE}

RUN sed -i 's|https://dl-cdn.alpinelinux.org/alpine|https://mirrors.tuna.tsinghua.edu.cn/alpine|g' /etc/apk/repositories \
    && apk add --no-cache tzdata curl
COPY --from=backend /commands/* /app/

VOLUME /data /imports
WORKDIR /data
HEALTHCHECK CMD curl -sSLf http://localhost:8080/health
ENTRYPOINT ["/app/workout-tracker"]
EXPOSE 8080
