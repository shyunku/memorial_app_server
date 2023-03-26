mkdir -p certificates
openssl req -x509 -newkey rsa:4096 -keyout certificates/server.key -out certificates/server.crt -days 365 -nodes