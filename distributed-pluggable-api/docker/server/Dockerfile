# syntax = docker/dockerfile:1

FROM alpine/git AS clone

WORKDIR /repo
RUN git clone https://github.com/xscotophilic/async-fibonacci-lab.git .

FROM node:20-alpine AS base

ENV NODE_ENV=production
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app \
  && chown -R app:app /app
USER app

ARG SERVER_PATH=/repo/server

COPY --from=clone ${SERVER_PATH}/package*.json ./
RUN npm ci --omit=dev && npm cache clean --force

COPY --from=clone ${SERVER_PATH}/src ./src
COPY --from=clone ${SERVER_PATH}/migrations ./migrations

EXPOSE 5000

CMD ["npm", "run", "start"]
