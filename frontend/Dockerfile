# ── Build stage ──────────────────────────────────────────────────────────────
FROM node:22-alpine AS build

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build -- --configuration production

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM nginx:1.27-alpine AS runtime

# Replace default nginx config with SPA-friendly static serving config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built app to nginx static root
COPY --from=build /app/dist/tradekai/browser /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
	CMD wget -qO- http://127.0.0.1/ >/dev/null 2>&1 || exit 1

CMD ["nginx", "-g", "daemon off;"]
