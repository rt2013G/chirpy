FROM debian:stable-slim
COPY chirpy /bin/chirpy
ENV PORT 8080
CMD ["/bin/chirpy"]
