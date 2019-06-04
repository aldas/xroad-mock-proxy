# Installing application as Systemd service

This guide assumes that you are inside folder that contains following:

* `xroad-mock-proxy` - application executable
* `.xroad-mock-proxy-example.yaml` - application configuration file example
* `xroad-mock-proxy.service` - Systemd service file
* `certificates` - folder containing X-road server certificates
* `mock` - folder containing X-road mock service response templates
* `web` - folder containing mock/proxy frontend files


Hypothetical use case for proxy/mock: 

1. You are required to communicate with X-road through client auth TLS security server.
2. X-road security server is located at (host:port) `xroad.lan.ee:443`
3. Client Apps must use client auth TLS even with Xroad mock/proxy
4. You decide to serve proxy on all network interfaces on port `18443` because we are going to connect to it from different server


## Setup

1. Create certificates for proxy server. This is where we serve your proxy for Xroad
    * read [/docs/certificates.md](../../docs/certificates.md)
    
    ```bash
    # generate private key with password
    export PASSWORD=SuperSecret1 
    openssl genrsa -passout env:PASSWORD -aes128 -out proxy_server_key.pem 4096
    
    # generate certificate
    openssl req -x509 -key proxy_server_key.pem -passin env:PASSWORD -out proxy_server_cert.pem -nodes -days 3650 -subj "/CN=localhost/O=Xroad\ mock\ proxy"
    ```
    
2. Create client certificates for you application. 
    This needed as we server proxy over client certificate auth TLS  (as our X-road security server does)

    Create key and certificate signing request (CSR) for `client app1` with `clientAuth` extension
    NB: drop `-nodes` from args if you want private key to have password
    
    ```bash
    # Ubuntu: /etc/pki/tls/openssl.cnf
    # Centos: /etc/pki/tls/openssl.cnf
    openssl req -newkey rsa:4096 -keyout client_app1_key.pem -out client_app1_csr.pem -nodes -subj "/CN=client-app1" \
        -reqexts cert_ext -config <(cat /etc/pki/tls/openssl.cnf <(printf "\n[cert_ext]\nextendedKeyUsage=serverAuth,clientAuth"))
    ```
    
    Sign `client app1` CSR with server key to create client certificate.
    NB: If you are going to issue multiple certificates then increment `-set_serial 01` value for unique ID for certificates
    
    ```bash
    openssl x509 -req -in client_app1_csr.pem -CA proxy_server_cert.pem -CAkey proxy_server_key.pem -out client_app1.pem -set_serial 01 -days 3650
    ```
    
    Create PKCS#12 keystore for our `client-app1` certificate and key. This is used as client app JAVA keystore
    
    ```bash
    # NB: password is a must. The KeyStore fails to work with JSSE without a password
    openssl pkcs12 -export -in client_app1.pem -inkey client_app1_key.pem \
                   -out client_app1.p12 -name xroad-proxy \
                   -CAfile proxy_server_cert.pem -caname root
    ```

3. Extract X-road security server certificates and private key from existing java keystore

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

4. Create system user for application. For security reasons no shell, no home, no nothing
    ```bash
    sudo adduser -r -s /usr/sbin/nologin xroad
    ```

5. Create folder for application and its configuration
    ```bash
    sudo mkdir -p /opt/app/xroad-mock-proxy
    sudo chown xroad:xroad /opt/app/xroad-mock-proxy
    sudo chmod 750 /opt/app/xroad-mock-proxy
    ```

6. Copy certificates to conf folder
    ```bash
    sudo mkdir -p /opt/app/xroad-mock-proxy/certificates/
 
    sudo cp proxy_server_*.pem /opt/app/xroad-mock-proxy/certificates/
    sudo cp xroad-*.pem /opt/app/xroad-mock-proxy/certificates/
    ```

7. Copy mock response files for predefined mock responses
    ```bash
    sudo cp -r mock /opt/app/xroad-mock-proxy/
    ```

8. Copy `web` assets for proxy and mock frontend
    ```bash
    sudo cp -r web /opt/app/xroad-mock-proxy/
    ```

9. Copy application and it's configuration files to `/opt/app/xroad-mock-proxy`
    ```bash
    sudo cp xroad-mock-proxy /opt/app/xroad-mock-proxy/
    sudo cp .xroad-mock-proxy-example.yaml /opt/app/xroad-mock-proxy/.xroad-mock-proxy.yaml
 
    # fix permissions for this and previous steps
    sudo chown xroad:xroad -R /opt/app/xroad-mock-proxy
    sudo chmod 750 -R /opt/app/xroad-mock-proxy
    ```

10. Edit configuration file `/opt/app/xroad-mock-proxy/.xroad-mock-proxy.yaml`
    
    Configure network interface, port and certificates for proxy server
    ```yaml
    proxy:
      server:
        address: '0.0.0.0:18443'
        tls:
          ca_file: '/opt/app/xroad-mock-proxy/certificates/proxy_server_cert.pem'
          cert_file: '/opt/app/xroad-mock-proxy/certificates/proxy_server_cert.pem'
          key_file: '/opt/app/xroad-mock-proxy/certificates/proxy_server_key.pem'
          key_password: 'SuperSecret1' # (optional)
          force_client_cert_auth: true # force client auth TLS instead of regular HTTPS
    ```
    
    Configure X-road security server as default route to proxy requests
    ```yaml
    proxy:
       routes:
         servers:
           - name: 'real-xroad'
             is_default: true
             address: 'https://xroad.lan.ee:443'
             tls:
               ca_file: '/opt/app/xroad-mock-proxy/certificates/xroad-cert.pem'
               cert_file: '/opt/app/xroad-mock-proxy/certificates/xroad-cert.pem'
               key_file: '/opt/app/xroad-mock-proxy/certificates/xroad-key.pem'
               key_password: '' # (optional) fill if you created private key with password
    ```

11. Copy systemd service file, enable service to start after reboot and start it right now
    ```bash
    sudo cp xroad-mock-proxy.service /lib/systemd/system/.
    sudo chmod 755 /lib/systemd/system/xroad-mock-proxy.service
    
    sudo systemctl daemon-reload
    sudo systemctl enable xroad-mock-proxy.service
    sudo systemctl start xroad-mock-proxy.service
    ```
    
    Check service status
    ```bash
    sudo systemctl start xroad-mock-proxy.service
    ```

    Tail service logs from `journalctl` to see if everything is ok
    ```bash
    sudo journalctl -u xroad-mock-proxy -f
    ```

12. Test proxy with CURL
    ```bash
    curl -v \
        --cacert ./proxy_server_cert.pem \
        --cert ./client_app1.pem \
        --key ./client_app1_key.pem \
        -X POST -d @../mock/rr.rr456.v1/rr456.paring.xml \
        --header "Content-Type: text/xml;charset=UTF-8" https://localhost:18443/cgi-bin/consumer_proxy
    ```
