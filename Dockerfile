FROM ubuntu

RUN apt-get update
RUN apt-get install -y ca-certificates
COPY ./project-gandalf /project-gandalf
