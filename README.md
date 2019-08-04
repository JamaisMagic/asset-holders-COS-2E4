# asset-holders-COS-2E4
asset-holders-COS-2E4

## Check information

GET https://explorer.picoluna.com/api/v1/asset/holders/cos-2e4/item?address=bnb-address

For example:

```bash
curl "https://explorer.picoluna.com/api/v1/asset/holders/cos-2e4/item?address=bnb1jxfh2g85q3v0tdq56fnevx6xcxtcnhtsmcu64m"

curl "https://explorer.picoluna.com/api/v1/asset/holders/cos-2e4/item?address=bnb1u9j9hkst6gf09dkdvxlj7puk8c7vh68a0kkmht"

```

Data syncs every 30 minutes.

## pprof example

1. Send a http request to the url.

```bash
curn -v https://explorer.picoluna.com/api/v1/test/cpu
```

2. Run go tool pprof to check profile

```bash
go tool pprof https://explorer.picoluna.com/debug/pprof/profile
```

### Referrers

1. [https://blog.golang.org/profiling-go-programs](https://blog.golang.org/profiling-go-programs)
2. [https://flaviocopes.com/golang-profiling/](https://flaviocopes.com/golang-profiling/)
3. [https://www.integralist.co.uk/posts/profiling-go/](https://www.integralist.co.uk/posts/profiling-go/)
