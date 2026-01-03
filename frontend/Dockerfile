# Frontend Dockerfile for static HTML deployment
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

# Expose port 80
EXPOSE 80

# Optimized health check with shorter intervals
HEALTHCHECK --interval=15s --timeout=2s --start-period=3s --retries=2 \
  CMD wget --quiet --tries=1 --spider http://localhost/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
