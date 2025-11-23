FROM node:18-alpine AS base

# Install dependencies only when needed
FROM base AS deps
WORKDIR /app
COPY package.json package-lock.json* ./
COPY frontend/packages ./frontend/packages
# Copy specific app package.json (requires context to be root)
ARG APP_PATH
COPY frontend/${APP_PATH}/package.json ./frontend/${APP_PATH}/
RUN npm install

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

ARG APP_PATH
RUN npm run build --workspace=frontend/${APP_PATH}

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app

ENV NODE_ENV production

COPY --from=builder /app/frontend/${APP_PATH}/public ./public
COPY --from=builder /app/frontend/${APP_PATH}/.next/standalone ./
COPY --from=builder /app/frontend/${APP_PATH}/.next/static ./.next/static

EXPOSE 3000
CMD ["node", "server.js"]
