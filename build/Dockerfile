FROM node:22 as nodebuilder
WORKDIR /app

RUN npm install -g @angular/cli

COPY frontend /app

RUN ng build 
RUN ls /app/dist

FROM golang:1.25 as gobuilder


ARG CGO_ENABLED=0
ARG GOOS=linux

ENV CGO_ENABLED=${CGO_ENABLED} \
    GOOS=${GOOS}

WORKDIR /app
COPY . /app

RUN go build \
    -trimpath \
    -ldflags="-s -w" \
    -gcflags="all=-l -B" \
    -o /app/ff-routing .



FROM scratch
COPY --from=nodebuilder /app/dist /frontend
COPY --from=gobuilder /app/ff-routing /bin/ff-routing
ENTRYPOINT ["/bin/ff-routing"]