# ── Build stage ──────────────────────────────────────────────────────────────
FROM node:22-alpine AS build

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build -- --configuration production

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM nginx:1.27-alpine AS runtime

# Replace default nginx config with our reverse-proxy config
COPY ../deployments/nginx/nginx.conf /etc/nginx/conf.d/default.conf

# Copy built app to nginx static root
COPY --from=build /app/dist/tradekai/browser /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
