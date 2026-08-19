# syntax=docker/dockerfile:1

# ---------- build ----------
FROM node:20-alpine AS build
WORKDIR /src

# npm ci installs exactly what the lockfile pins and fails loudly if it has
# drifted from package.json, which is what makes the image reproducible.
COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# Vite inlines env vars at BUILD time, not runtime — this value ends up baked
# into the JS bundle, so it cannot be changed by restarting the container.
# "/api" keeps every call on the page's own origin, where nginx proxies it to
# the backend (see nginx.conf).
ARG VITE_API_URL=/api
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

# ---------- runtime ----------
FROM nginx:1.27-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /src/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1/ >/dev/null 2>&1 || exit 1
