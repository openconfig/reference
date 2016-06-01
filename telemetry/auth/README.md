# auth
Provides authentication used for gRPC and TACACS user/password integration.
Today gRPC only supports SSL authentication. By providing a username/password
credential into the gRPC metadata, the OpenConfig server can use the credential
to link device authentication via TACACS or local authentication.

# Example
```
pc := credential.NewPassCred(*user, *passwd, true)
DialOptions = append(DialOptions, grpc.WithPerRPCCredentials(pc))
conn, err := grpc.Dial(Target, DialOptions...)
```
