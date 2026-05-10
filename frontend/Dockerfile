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

# Final stage
FROM node:20-alpine

WORKDIR /app

# Copy the built application and dependencies
COPY --from=builder /app/build ./build
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json

# Environment variables for production
ENV NODE_ENV=production
ENV HOST=0.0.0.0

# Expose a default port (Railway will override this via the PORT env var)
EXPOSE 3000

# Run the SvelteKit server
CMD ["node", "build"]
