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

## pprof cpu profile example

1. add "runtime/pprof" package to you app.
2. Setup a http server
    ```go
    log.Println(http.ListenAndServe(":" + port, nil))
    ```
3. Start cpu profile
    ```bash
    go tool pprof https://explorer.picoluna.com/debug/pprof/profile
    ```
4. At the same time, send requests to the url.
    ```bash
    ab -c 1 -n 4 https://explorer.picoluna.com/api/v1/test/cpu
    ```
5. After that, you can get cpu profile, and use pprof command to check information.

    ```text
    Fetching profile over HTTP from https://explorer.picoluna.com/debug/pprof/profile
    Saved profile in /Users/jamais/pprof/pprof.main.samples.cpu.005.pb.gz
    File: main
    Type: cpu
    Time: Aug 4, 2019 at 7:19pm (CST)
    Duration: 30s, Total samples = 17.40s (58.00%)
    Entering interactive mode (type "help" for commands, "o" for options)
    (pprof) top
    Showing nodes accounting for 17.40s, 100% of 17.40s total
          flat  flat%   sum%        cum   cum%
        17.40s   100%   100%     17.40s   100%  main.createHeavy
             0     0%   100%     17.40s   100%  main.handleCpuTest
             0     0%   100%     17.40s   100%  net/http.(*ServeMux).ServeHTTP
             0     0%   100%     17.40s   100%  net/http.(*conn).serve
             0     0%   100%     17.40s   100%  net/http.HandlerFunc.ServeHTTP
             0     0%   100%     17.40s   100%  net/http.serverHandler.ServeHTTP
    ```
### Referrers

1. [https://blog.golang.org/profiling-go-programs](https://blog.golang.org/profiling-go-programs)
2. [https://flaviocopes.com/golang-profiling/](https://flaviocopes.com/golang-profiling/)
3. [https://www.integralist.co.uk/posts/profiling-go/](https://www.integralist.co.uk/posts/profiling-go/)
4. [https://golang.org/pkg/net/http/pprof/](https://golang.org/pkg/net/http/pprof/)
5. [https://matoski.com/article/golang-profiling-flamegraphs/](https://matoski.com/article/golang-profiling-flamegraphs/)
