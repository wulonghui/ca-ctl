# ca-ctl

A simple certificate manager written in Go. Easy to use with limited capability.



## Common Uses

ca-ctl allows you to build your own certificate system:

1. Create certificate authority
2. Create, issue and export host certificates
3. Manage host identities
4. Deploy a Public Key Infrastructure

Primarly used for [coreos/etcd](https://github.com/coreos/etcd) SSL/TLS testing.

## Certificate architecture

ca-ctl inits a certificate authority, and issues certificates using the authority only. It indicates the length of authorization path is at most 2.

## Examples

### Create a new certificate authority:

```
$ ./ca-ctl init
Created ca/key
Created ca/crt
```

### Create a new host identity, including keypair and certificate request:

```
$ ./ca-ctl new-cert alice
Created alice/key
Created alice/csr
```

ca-ctl uses 127.0.0.1 for IP SAN in default. If etcd has peer address $etcd_ip other than 127.0.0.1, run `./ca-ctl new-cert -ip $etcd_ip alice` instead.

If your server has mutiple ip addresses or domains, use comma seperated ip/domain list with -ip/-domain. eg: `./ca-ctl new-cert -ip $etcd_ip1,$etcd_ip2 -domain $etcd_domain1,$etcd_domain2`

### Sign certificate request of host and generate the certificate:

```
$ ./ca-ctl sign alice
Created alice/crt from alice/csr signed by ca.key
```

### Export the certificate chain for host:

```
$ ./ca-ctl chain alice
----BEGIN CERTIFICATE-----
CA certificate body
-----END CERTIFICATE-----
----BEGIN CERTIFICATE-----
alice certificate body
-----END CERTIFICATE-----
```

### Package up the certificate and key of host:

```
$ ./ca-ctl export alice > alice.tar
```

Because etcd takes unencrypted key for `-key-file` and `-peer-key-file`, you should use `./ca-ctl export --insecure alice > alice.tar` to export private key.

### List the status of all certificates:

```
$ ./ca-ctl status
ca: WARN (60 days until expiration)
alice: OK (120 days until expiration)
bob: Unsigned
```

## Getting Started

### Building

ca-ctl must be built with Go 1.3+. You can build etcd-ca from source:

```
$ git clone https://github.com/wulonghui/ca-ctl
$ cd ca-ctl
$ ./build
```

This will generate a binary called `./bin/ca-ctl`

### Examples

[generate certificates for etcd](./hack/etcd_example.sh)

## Project Details

### Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches and contacting developers via IRC and mailing lists.

### License

ca-ctl is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.
