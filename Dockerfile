FROM busybox:glibc
COPY ./bookList /bookList
EXPOSE 8000
CMD ["./bookList"]
