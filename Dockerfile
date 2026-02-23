FROM gcr.io/distroless/static-debian12:nonroot

COPY gavel /usr/local/bin/gavel

ENTRYPOINT ["/usr/local/bin/gavel"]
CMD ["--help"]
