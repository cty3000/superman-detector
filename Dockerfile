FROM golang:1.11-alpine AS builder

ENV APP_NAME superman-detector

RUN set -eux \
    && apk --no-cache add --virtual build-dependencies cmake g++ make unzip curl upx git

WORKDIR ${GOPATH}/src/superman-detector

COPY . .

RUN make

RUN CGO_ENABLED=1 \
    CGO_CXXFLAGS="-g -Ofast -march=native" \
    CGO_FFLAGS="-g -Ofast -march=native" \
    CGO_LDFLAGS="-g -Ofast -march=native" \
    GOOS=$(go env GOOS) \
    GOARCH=$(go env GOARCH) \
    go build --ldflags '-s -w -linkmode "external" -extldflags "-static -fPIC -pthread -std=c++11 -lstdc++"' -a -tags "cgo netgo" -installsuffix "cgo netgo" -o "${APP_NAME}" \
    && mv "${APP_NAME}" "/usr/bin/${APP_NAME}"
    #&& mv "${GOPATH}/bin/${APP_NAME}" "/usr/bin/${APP_NAME}"

RUN apk del build-dependencies --purge \
    && rm -rf "${GOPATH}"

# Start From Scratch For Running Environment
FROM scratch

ENV APP_NAME superman-detector

COPY --from=builder /usr/bin/${APP_NAME} /usr/bin/${APP_NAME}

ENTRYPOINT ["/usr/bin/superman-detector"]
