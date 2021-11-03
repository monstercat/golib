module github.com/monstercat/golib

go 1.16

require (
	cloud.google.com/go/logging v1.4.2
	cloud.google.com/go/storage v1.18.2
	github.com/Masterminds/squirrel v1.5.1
	github.com/aws/aws-sdk-go v1.41.12
	github.com/fsouza/fake-gcs-server v1.30.2
	github.com/gin-contrib/sse v0.1.0
	github.com/gin-gonic/gin v1.7.4
	github.com/golang/protobuf v1.5.2
	github.com/jmoiron/sqlx v1.3.4
	github.com/keimoon/gore v0.0.0-20160317032603-1ac8b93b5fdb
	github.com/lib/pq v1.10.3
	github.com/monstercat/pgnull v0.0.0-20211008053451-c7be7177fe76
	github.com/monstercat/websocket v0.0.0-20211027191942-c20559603674
	github.com/pkg/errors v0.9.1
	github.com/shopspring/decimal v1.3.1
	github.com/tidwall/gjson v1.10.2
	github.com/trubitsyn/go-zero-width v1.0.1
	google.golang.org/api v0.59.0
	google.golang.org/genproto v0.0.0-20211027162914-98a5263abeca
)

replace github.com/monstercat/pgnull => ../pgnull

replace github.com/monstercat/websocket => ../websocket
