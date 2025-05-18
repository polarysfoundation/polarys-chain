package params

type Config struct {
	MaxProposalSize int64 `mapstructure:"max_proposal_size"`
	MaxTxSize       int64 `mapstructure:"max_tx_size"`
}


