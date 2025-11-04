# Frontend Dockerfile for static HTML deployment
FROM nginx:alpine

# Accept build arguments from Coolify
ARG COMMIT_SHA
ARG BUILD_DATE

# Set environment variables that nginx can use (these will be used by the backend API)
ENV FRONTEND_BUILD_COMMIT=${COMMIT_SHA:-unknown}
ENV FRONTEND_BUILD_DATE=${BUILD_DATE:-unknown}

# Copy static files to nginx
COPY index.html /usr/share/nginx/html/
COPY templates /usr/share/nginx/html/templates/

# Expose port 80
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
