package gaspool

var (

	// Unit have 15 decimals
	// 1 PRY = 1000000000000000
	// 0.1 PRY = 100000000000000
	// 0.01 PRY = 10000000000000
	// 0.001 PRY = 1000000000000
	// 0.0001 PRY = 100000000000
	// 0.00001 PRY = 10000000000
	// 0.000001 PRY = 1000000000
	// 0.0000001 PRY = 100000000
	// 0.00000001 PRY = 10000000
	// 0.000000001 PRY = 1000000
	// 0.0000000001 PRY = 100000
	// 0.00000000001 PRY = 10000
	// 0.000000000001 PRY = 1000
	// 0.0000000000001 PRY = 100
	// 0.00000000000001 PRY = 10
	// 0.000000000000001 PRY = 1

	BaseGas         = 8000 // 0.000000000008 PRY
	BytesCost       = 20   // 0.00000000000002 PRY
	PayloadByteCost = 32   // 0.000000000000032 PRY
	ValueByteCost   = 16   // 0.000000000000016 PRY
	MinTipPerGas    = 5    // 0.000000000000005 PRY

	GasDivisor       = 10000 // 100%
	MaxGasMultiplier = 6000  // 60%
	MinGasMultiplier = 500   // 5%
	GasTipMultiplier = 1000  // 10%
	GasRefactor      = 3000  // 30%

	DefaultGasTarget    = 1000000                                        // 0.000000001 PRY
	DefaultMinGas       = DefaultGasTarget - MinGasMultiplier*GasDivisor // default_min_gas = default_gas_target - min_gas_multiplier * gas_divisor
	DefaultMaxGasTarget = DefaultGasTarget + MaxGasMultiplier*GasDivisor // default_max_target = default_gas_target + max_gas_multiplier * gas_divisor
	DefaultGasInterval  = 60                                             // 60 seconds
)
