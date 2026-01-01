module github.com/zoobzio/atom/pkg/bson

go 1.24

toolchain go1.25.5

require (
	github.com/zoobzio/atom v0.0.0
	go.mongodb.org/mongo-driver/v2 v2.4.1
)

require github.com/zoobzio/sentinel v0.1.4 // indirect

replace github.com/zoobzio/atom => ../..
