module github.com/blocktree/nulsio2-adapter

go 1.12

require (
	github.com/DataDog/zstd v1.4.5 // indirect
	github.com/Sereal/Sereal v0.0.0-20200611165018-70572ef94023 // indirect
	github.com/astaxie/beego v1.12.0
	github.com/blocktree/go-owcdrivers v1.2.0
	github.com/blocktree/go-owcrypt v1.1.1
	github.com/blocktree/openwallet/v2 v2.0.5
	github.com/imroc/req v0.2.4
	github.com/pkg/errors v0.9.1
	github.com/shopspring/decimal v0.0.0-20200105231215-408a2507e114
	github.com/tidwall/gjson v1.3.5
	golang.org/x/crypto v0.0.0-20191227163750-53104e6ec876
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b // indirect
)

//replace github.com/blocktree/openwallet => ../../openwallet
