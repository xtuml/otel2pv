# TLS Configuration
The `tlsConfig` object is used to configure the TLS settings for the producers and consumers. The object has the following structure:
```json
{
    "certFile": <string - the path to the client certificate file>,
    "keyFile": <string - the path to the client private key file, must be unencrypted>,
    "clientCaFile": <string, optional - the path to the CA certificate file>,
    "rootCaFile": <string, optional - the path to the root CA certificate file>,
    "tlsVersion": <string, optional - the minimum TLS version to use, default is "TLSv1.2">,
}
```

This configuration provides enough flexibility to configure the TLS settings for the producers and consumers. The `certFile` and `keyFile` are required to be set in order for the client to work correctly. The `clientCaFile` and `rootCaFile` are optional and can be set if the client requires them. The `tlsVersion` is also optional and can be set to the minimum TLS version that the client should use.