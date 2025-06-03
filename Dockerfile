FROM golang AS builder

WORKDIR /app
COPY . .

ENV GOFLAGS="-buildvcs=false"
ENV CGO_ENABLED=0
RUN go build -o /usr/local/bin/reeve-step .

FROM docker

COPY --chmod=755 --from=builder /usr/local/bin/reeve-step /usr/local/bin/

WORKDIR /reeve/src

# FILES: Space separated list of file patterns (see https://pkg.go.dev/github.com/bmatcuk/doublestar/v4#Match) to be searched (shell syntax)
ENV FILES=
# REVISION_VAR: Name of a runtime variable for setting the files' revision to
ENV REVISION_VAR=FILE_REV

ENTRYPOINT ["reeve-step"]
