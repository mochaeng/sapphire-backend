# Certificates for Postgres SSL

[read more here](https://stackoverflow.com/questions/55072221/deploying-postgresql-docker-with-ssl-certificate-and-key-with-volumes)

```sh
mkdir certs && cd certs
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -subj "/CN=localhost"
openssl x509 -req -in server.csr -signkey server.key -out server.crt -days 3650
```

`postgres` as owner of the `server.key`:

```sh
chown 70:70 server.key # 70:70 for alpine, 999:999 for debian
chmod 600 server.key
```

`root` as owner of the `server.key`:

```sh
chown 70:70 server.key # 70:70 for alpine, 999:999 for debian
chmod 640 server.key
```
