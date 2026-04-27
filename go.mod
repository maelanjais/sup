module github.com/pressly/sup

go 1.24

require (
	github.com/goware/prefixer v0.0.0-20160118172347-395022866408
	github.com/hashicorp/go-plugin v1.7.0
	github.com/maelanjais/sup-hcl2-plugin v0.0.0-20260427114201-79dec5cbec35
	github.com/mikkeloscar/sshconfig v0.0.0-20190102082740-ec0822bcc4f4
	github.com/pkg/errors v0.9.1
	golang.org/x/crypto v0.38.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/fatih/color v1.13.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/grpc v1.61.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace github.com/maelanjais/sup-hcl2-plugin => ../Hcl2Plugin
