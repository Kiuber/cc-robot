FROM golang:1.15.14

ENV WORKDIR /opt/src

ADD app/main   $WORKDIR/main
RUN chmod +x $WORKDIR/main

ADD app/config   $WORKDIR/config

WORKDIR $WORKDIR
CMD ./main -env=prod
