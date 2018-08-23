module github.com/michilu/cel-spec-go

replace (
	github.com/antlr/antlr4 v0.0.0-20180728001836-7d0787e29ca8 => github.com/michilu/antlr4 v0.0.0-20180803091604-411960b1878f
	github.com/google/cel-spec v0.0.2-0.20180709214636-040bf10482b7 => github.com/michilu/cel-spec v0.0.0-20180805110655-507a3c16c340
)

require (
	github.com/antlr/antlr4 v0.0.0-20180728001836-7d0787e29ca8 // indirect
	github.com/asaskevich/govalidator v0.0.0-20180720115003-f9ffefc3facf
	github.com/golang/protobuf v1.2.0 // indirect
	github.com/google/cel-go v0.0.0-20180713002128-97dcd37b146e
	github.com/google/cel-spec v0.0.2-0.20180709214636-040bf10482b7
	github.com/lkesteloot/astutil v0.0.0-20130122170032-b6715328cfa5
	github.com/michilu/boilerplate v0.0.0-20180819010701-3d1fe965c56b
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.1.0
	google.golang.org/grpc v1.14.0
)
