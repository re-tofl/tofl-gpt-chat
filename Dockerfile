FROM ubuntu:latest
LABEL authors="Artem"

ENTRYPOINT ["top", "-b"]