FROM golang:1.19

RUN go install github.com/audibleblink/passdb@latest
CMD ["passdb"]

# $ docker build -t passdb-server .
# $ docker run --env-file .env -p 3000:3000  passdb-server
