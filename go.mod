module github.com/sagernet/quic-go

go 1.20

//replace github.com/JimmyHuang454/qtls-go1-20-JLS => ../qtls-go1-20-JLS
require (
	github.com/JimmyHuang454/JLS-go v0.0.0-20230831150107-90d536585ba0 // indirect
	github.com/JimmyHuang454/qtls-go1-20-JLS v0.0.0-20230909050831-219b978e3919
)

require (
	github.com/cloudflare/circl v1.3.3 // indirect
	github.com/francoispqt/gojay v1.2.13
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/google/pprof v0.0.0-20210407192527-94a9f03dee38 // indirect
	github.com/onsi/ginkgo/v2 v2.9.5 // indirect
	github.com/quic-go/qpack v0.4.0
	github.com/sagernet/cloudflare-tls v0.0.0-20230829051644-4a68352d0c4a
	golang.org/x/crypto v0.13.0
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9
	golang.org/x/net v0.15.0
	golang.org/x/sys v0.12.0
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/tools v0.13.0 // indirect
)

require github.com/stretchr/testify v1.8.4

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
