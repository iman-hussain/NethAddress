# Frontend Dockerfile for static HTML deployment
# Multi-stage build: compile SCSS then serve with nginx

# =============================================================================
# Stage 1: SCSS Compilation
# =============================================================================
FROM node:20-alpine AS scss-builder

# Install Dart Sass (official, fast compiler)
RUN npm install -g sass

WORKDIR /build

# Copy only SCSS files for compilation
COPY static/scss ./scss/

# Compile SCSS to CSS with production optimisations
# --style=compressed minifies the output
# --no-source-map avoids leaking file paths
RUN sass scss/main.scss css/styles.css --style=compressed --no-source-map

# =============================================================================
# Stage 2: Production Nginx Image
# =============================================================================
FROM nginx:1.25-alpine

# Accept build arguments from Coolify (for reference/documentation purposes)
# Note: The backend will receive frontend build info via runtime environment variables,
# not from this container. These args are kept for build-time logging/debugging.
ARG COMMIT_SHA
ARG BUILD_DATE

# Copy static files to nginx in order of change frequency
# (least frequently changed first for better layer caching)
COPY static /usr/share/nginx/html/static/
COPY templates /usr/share/nginx/html/templates/
COPY manifest.json /usr/share/nginx/html/
COPY index.html /usr/share/nginx/html/

# Copy compiled CSS from builder stage (overwrites any existing styles.css)
COPY --from=scss-builder /build/css/styles.css /usr/share/nginx/html/static/css/styles.css

# Expose port 80
EXPOSE 80

# Optimised health check with shorter intervals
HEALTHCHECK --interval=15s --timeout=2s --start-period=3s --retries=2 \
  CMD wget --quiet --tries=1 --spider http://localhost/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
