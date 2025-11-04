# Frontend Dockerfile for static HTML deployment
FROM nginx:alpine

# Accept build arguments from Coolify (for reference/documentation purposes)
# Note: The backend will receive frontend build info via runtime environment variables,
# not from this container. These args are kept for build-time logging/debugging.
ARG COMMIT_SHA
ARG BUILD_DATE

# Copy static files to nginx
COPY index.html /usr/share/nginx/html/
COPY templates /usr/share/nginx/html/templates/

# Expose port 80
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
