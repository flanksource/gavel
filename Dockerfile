FROM debian:12-slim

# Hosted gavel serve runs git for pushed repositories, GitHub mirrors and run branches.
RUN apt-get update \
  && apt-get install -y --no-install-recommends git ca-certificates \
  && rm -rf /var/lib/apt/lists/* \
  && groupadd --gid 65532 nonroot \
  && useradd --uid 65532 --gid 65532 --create-home --home-dir /home/nonroot nonroot

COPY gavel /usr/local/bin/gavel

USER 65532:65532
ENTRYPOINT ["/usr/local/bin/gavel"]
CMD ["--help"]
