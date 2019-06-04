# HOW TO for certificate related things

## Generate server key and certificate (for serving HTTPS on proxy/mock endpoints)

if you are to use password protected private key and you are not using following commands to create private key 
then make sure that private key contains `DEK-Info:` header or app is unable to decrypt private key
```bash
# generate private key with password
export PASSWORD=SuperSecret1 
openssl genrsa -passout env:PASSWORD -aes128 -out server_key.pem 4096

# generate certificate
openssl req -x509 -key server_key.pem -passin env:PASSWORD -out server_cert.pem -nodes -days 3650 -subj "/CN=localhost/O=Xroad\ mock\ proxy"
```

## Client cert authentication setup (both server and clients)

Prerequisite: Create self-signed certificate for server with previous `Generate server key and certificate` paragraph. 
`server_cert.pem` will also be our CA cert.


Create key and certificate signing request (CSR) for `client app1` with `clientAuth` extension
NB: drop `-nodes` from args if you want private key to have password
```bash
openssl req -newkey rsa:4096 -keyout client_app1_key.pem -out client_app1_csr.pem -nodes -subj "/CN=client-app1" \
    -reqexts cert_ext -config <(cat /etc/ssl/openssl.cnf <(printf "\n[cert_ext]\nextendedKeyUsage=serverAuth,clientAuth"))
```

Sign `client app1` CSR with server key to create client certificate.
NB: If you are going to issue multiple certificates then increment `-set_serial 01` value for unique ID for certicates
```bash
openssl x509 -req -in client_app1_csr.pem -CA server_cert.pem -CAkey server_key.pem -out client_app1.pem -set_serial 01 -days 3650
```

Create PKCS#12 keystore for our `client-app1` certificate and key. This is used for JAVA keystore also
```bash
openssl pkcs12 -export -clcerts -in client_app1.pem -inkey client_app1_key.pem -out client_app1.p12
```

Test with curl
```bash
curl -v -k --cert client_app1.p12 --cert-type p12 \
    -X POST -d @testdata/rr.rr456.v1/rr456.paring.xml \
    --header "Content-Type: text/xml;charset=UTF-8" https://localhost:18443/cgi-bin/consumer_proxy
```

## Extract X-road security server cert/key out of existing JAVA Keystore

In case you already have existing java application that is configured to communicate with tls
to X-road security server you can use following commands to extract security server cert/key
so our proxy can also connect to server

```bash
# convert Java keystore to PKCS12 so we can get cert/key out of it
keytool -importkeystore \
    -srckeystore keystore.jks \
    -destkeystore keystore.p12 \
    -deststoretype PKCS12 \
    -srcalias xroad-client-alias \
    -deststorepass keystore-password \
    -destkeypass privatekey-password

# Export certificate using openssl:
openssl pkcs12 -in keystore.p12 -nokeys -out xroad-cert.pem
# Export unencrypted private key:
openssl pkcs12 -in keystore.p12 -nodes -nocerts -out xroad-key.pem
```

