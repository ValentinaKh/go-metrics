## Результат оптимизации
```
Type: inuse_space                                                                                                                                        
Time: Nov 16, 2025 at 2:14pm (MSK)                                                                                                                       
Duration: 120.01s, Total samples = 902.59kB                                                                                                              
Showing nodes accounting for -2349.84kB, 260.35% of 902.59kB total                                                                                       
flat  flat%   sum%        cum   cum%                                                                                                               
-1805.17kB 200.00% 200.00% -2349.84kB 260.35%  compress/flate.NewWriter (inline)                                                                         
-544.67kB 60.35% 260.35%  -544.67kB 60.35%  compress/flate.(*compressor).initDeflate (inline)                                                           
0     0% 260.35%  -544.67kB 60.35%  compress/flate.(*compressor).init                                                                           
0     0% 260.35% -2349.84kB 260.35%  compress/gzip.(*Writer).Close
0     0% 260.35% -2349.84kB 260.35%  compress/gzip.(*Writer).Write
0     0% 260.35% -2349.84kB 260.35%  github.com/ValentinaKh/go-metrics/internal/handler/middleware.(*compressWriter).Close
0     0% 260.35% -2349.84kB 260.35%  github.com/ValentinaKh/go-metrics/internal/handler/middleware.GzipMW.func1
0     0% 260.35% -2349.84kB 260.35%  github.com/ValentinaKh/go-metrics/internal/handler/middleware.LoggingMw.func1
0     0% 260.35% -2349.84kB 260.35%  github.com/ValentinaKh/go-metrics/internal/server.createServer.ValidateHashMW.func4.1
0     0% 260.35% -2349.84kB 260.35%  github.com/go-chi/chi/v5.(*ChainHandler).ServeHTTP
0     0% 260.35% -2349.84kB 260.35%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
0     0% 260.35% -2349.84kB 260.35%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
0     0% 260.35% -2349.84kB 260.35%  net/http.(*conn).serve
0     0% 260.35% -2349.84kB 260.35%  net/http.HandlerFunc.ServeHTTP
0     0% 260.35% -2349.84kB 260.35%  net/http.serverHandler.ServeHTTP
```
