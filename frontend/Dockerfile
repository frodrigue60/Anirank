# Build stage
FROM oven/bun:alpine AS builder

WORKDIR /app

# Install dependencies
COPY package.json bun.lockb ./
RUN bun install

# Copy source code and build
COPY . .
ARG VITE_API_URL
ARG PUBLIC_API_URL
ENV VITE_API_URL=$VITE_API_URL
ENV PUBLIC_API_URL=$PUBLIC_API_URL
RUN bun run svelte-kit sync && bun run build

# Final stage - Serve with Nginx
FROM nginx:stable-alpine

# Copy built assets to Nginx html directory
COPY --from=builder /app/build /usr/share/nginx/html

# Copy custom nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port (Frontend port 80)
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
