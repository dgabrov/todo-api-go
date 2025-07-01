FROM amd64/alpine
EXPOSE 3001
WORKDIR /app
COPY todo-api-go .
CMD ["/app/todo-api-go"]
