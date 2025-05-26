package prydb

var (
	blocksByHeight            = "blocks/height/"
	blocksByHash              = "blocks/hash/"
	blocksLatest              = "blocks/latest/"
	transactionsByHash        = "transactions/confirmed/"
	transactionsByAccount     = "accounts/%s/transactions/"
	transactionsRejecteds     = "transactions/rejected/"
	transactionsByBlockHash   = "blocks/hash/%s/transactions/"
	transactionsByBlockHeight = "blocks/height/%s/transactions/"
	accounts                  = "accounts/block_%s/"
	txPools                   = "txpool/block_%s"
	transactionsByTxPool      = "txpool/%s/transactions/"
)
