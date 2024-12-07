# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files
COPY user/ui/wrale-dashboard/package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY user/ui/wrale-dashboard/ ./

# Build the application
ENV NEXT_PUBLIC_API_URL=http://api:8083
RUN npm run build

# Runtime stage
FROM node:18-alpine

WORKDIR /app

# Copy built assets
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/package*.json ./
COPY --from=builder /app/node_modules ./node_modules

# Environment variables
ENV NODE_ENV=production
ENV NEXT_PUBLIC_API_URL=http://api:8083
ENV PORT=3000

# Expose port
EXPOSE 3000

CMD ["npm", "start"]
