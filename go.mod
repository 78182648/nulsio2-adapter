module github.com/blocktree/nulsio2-adapter

go 1.12

require (
	github.com/asdine/storm v2.1.2+incompatible
	github.com/astaxie/beego v1.12.0
	github.com/blocktree/go-owcdrivers v1.2.0
	github.com/blocktree/go-owcrypt v1.1.1
	github.com/blocktree/openwallet v1.7.0
	github.com/imroc/req v0.2.4
	github.com/pkg/errors v0.8.1
	github.com/shopspring/decimal v0.0.0-20200105231215-408a2507e114
	github.com/siddontang/go v0.0.0-20180604090527-bdc77568d726
	github.com/tidwall/gjson v1.3.5
	golang.org/x/crypto v0.0.0-20191227163750-53104e6ec876
)

//replace github.com/blocktree/openwallet => ../../openwallet
