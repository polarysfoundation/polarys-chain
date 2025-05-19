package params

type Config struct {
	MaxProposalSize int64 `mapstructure:"max_proposal_size"`
	MaxTxSize       int64 `mapstructure:"max_tx_size"`
	MaxBlockSize    int64 `mapstructure:"max_block_size"`
	MaxTxPerBlock   int64 `mapstructure:"max_tx_per_block"`
	MinimalGasTip   int64 `mapstructure:"minimal_gas_tip"`
	
}

func LoadConfig() *Config {
	return DefaultConfig
}
