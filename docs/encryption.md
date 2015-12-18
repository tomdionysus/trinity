![Trinity DB Logo](../gfx/trinity_m.png) 

# Trinity DB - Certificates and Encryption

## Requirements

Trinity features mandatory node-to-node encryption, meaning that a node must be supplied with:

* A CA Certificate
* A Node Certificate and Key

Encryption, TLS and x509 are complex subjects and thorough study is recommended before going out in the wild in production, please see the Further Reading section below.

## Testing / Evaluation

To ease evaluation, a 'snakeoil' CA and Node certificate are provided in the [cert/](../cert) directory. **IMPORTANT: You should NEVER use these for anything except development**.

The two files `ca.pem` and `localhost.pem` represent a CA cert and a certificate for `localhost` signed by that CA.

## EasyRSA

The excellent [EasyRSA](https://github.com/OpenVPN/easy-rsa) project from [OpenVPN](http://community.openvpn.net/openvpn) makes running your own private CA trivial. Please follow the [Getting Started](https://github.com/OpenVPN/easy-rsa/blob/master/doc/EasyRSA-Readme.md) guide.

**Caveat:** Trinity nodes use the same cert for both authenticating incoming node connections as a server and authenticating to other nodes as a client - as such, the certificate needs both `clientAuth` and `serverAuth` set in the x509v3 extension `extendedKeyUsage`. In EasyRSA, this can be acheived by modifying the `x509-types/server` file and replacing the `extendedKeyUsage = serverAuth` line with `extendedKeyUsage = clientAuth,serverAuth` before building the server cert.

Briefly:

```bash
wget https://github.com/OpenVPN/easy-rsa/releases/download/3.0.1/EasyRSA-3.0.1.tgz
tar -xzf EasyRSA-3.0.1.tgz
cd EasyRSA-3.0.1

./easyrsa init-pki
./easyrsa build-ca # Enter a good password for the CA key
nano x509-types/server # Modify as per Caveat above
./easyrsa build-server-full <hostname> # Enter a short password for the cert

openssl rsa -in pki/private/<hostname>.key -out <hostname>.key # Enter the short password to decrypt the key
cat pki/issued/<hostname>.crt <hostname>.key > <hostname>.pem 
rm <hostname>.key
cp pki/ca.crt ./ca.pem

```

After this, the CA cert `ca.pem` and your certificate `<hostname>.pem` can be specified to the `--ca` and `--cert` flags in `trinity-server`. 

## Further Reading

* [x.509](https://en.wikipedia.org/wiki/X.509)
* [Transport Layer Security](https://tools.ietf.org/html/rfc5246)
* [EasyRSA](https://github.com/OpenVPN/easy-rsa) 

