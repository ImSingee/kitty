FROM gcr.io/distroless/static-debian12:nonroot

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/kitty /usr/bin/kitty
WORKDIR /worktree

ENTRYPOINT ["/usr/bin/kitty"]
