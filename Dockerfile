FROM golang AS builder

WORKDIR /app
COPY . .

ENV GOFLAGS="-buildvcs=false"
ENV CGO_ENABLED=0
RUN go build -o /usr/local/bin/reeve-step .

FROM docker

RUN apk add jq
COPY --chmod=755 --from=builder /usr/local/bin/reeve-step /usr/local/bin/

# FILES: Space separated list of files to be included (shell syntax)
ENV FILES=
# REVISION_VAR: Name of a runtime variable for setting the files' revision to
ENV REVISION_VAR=FILE_REV

ENTRYPOINT ["reeve-step"]
