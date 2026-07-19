# ── Build stage ───────────────────────────────────────────────────────────────
FROM node:20-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json* ./
RUN npm ci

COPY . .

# In Docker the frontend talks to nginx which proxies /api to the agent,
# so VITE_API_URL is empty (same-origin requests via nginx proxy).
RUN VITE_API_URL="" npm run build

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM nginx:1.25-alpine

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Copy our config and built assets
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/txline.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
