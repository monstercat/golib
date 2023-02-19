module github.com/monstercat/golib

go 1.19

require (
	cloud.google.com/go/logging v1.4.2
	cloud.google.com/go/storage v1.18.2
	github.com/Masterminds/squirrel v1.5.1
	github.com/aws/aws-sdk-go v1.41.12
	github.com/fsouza/fake-gcs-server v1.30.2
	github.com/gin-contrib/sse v0.1.0
	github.com/gin-gonic/gin v1.8.2
	github.com/golang/protobuf v1.5.2
	github.com/jmoiron/sqlx v1.3.4
	github.com/keimoon/gore v0.0.0-20160317032603-1ac8b93b5fdb
	github.com/lib/pq v1.10.3
	github.com/monstercat/pgnull v0.0.0-20211008053451-c7be7177fe76
	github.com/monstercat/websocket v0.0.0-20211027191942-c20559603674
	github.com/pkg/errors v0.9.1
	github.com/shopspring/decimal v1.3.1
	github.com/tidwall/gjson v1.11.0
	github.com/trubitsyn/go-zero-width v1.0.1
	google.golang.org/api v0.59.0
	google.golang.org/genproto v0.0.0-20211027162914-98a5263abeca
)

require (
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.11.2 // indirect
	github.com/goccy/go-json v0.10.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/paulrosania/go-charset v0.0.0-20190326053356-55c9d7a5834c // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/ugorji/go/codec v1.2.9 // indirect
	golang.org/x/crypto v0.6.0 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/monstercat/pgnull => ../pgnull

replace github.com/monstercat/websocket => ../websocket
