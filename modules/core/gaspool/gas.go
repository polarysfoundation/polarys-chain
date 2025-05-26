package gaspool

import (
	"fmt"
	"time"
)

type GasPool struct {
	gasTarget    uint64
	gasPrice     uint64
	maxGasTarget uint64
	gasInterval  uint64
	latestSync   uint64
}

func InitGasPool() *GasPool {
	return &GasPool{
		gasTarget:    uint64(DefaultGasTarget),
		gasInterval:  uint64(DefaultGasInterval),
		latestSync:   uint64(time.Now().Unix()),
		maxGasTarget: uint64(DefaultMaxGasTarget),
	}
}

func (gp *GasPool) GasPrice() uint64 {
	return gp.gasPrice
}

func (gp *GasPool) GasTarget() uint64 {
	return gp.gasTarget
}

func (gp *GasPool) MaxGasTarget() uint64 {
	return gp.maxGasTarget
}

func (gp *GasPool) SyncGasPool(latestGasTarget uint64, latestBlock uint64, transactionCount int) error {
	if latestGasTarget > uint64(gp.maxGasTarget) {
		return fmt.Errorf("gas target exceeds max value")
	}

	if !gp.expired() {
		return nil
	}

	totalGasTip := latestGasTarget / uint64(GasDivisor) * uint64(GasTipMultiplier)

	gasTargetWithoutTip := latestGasTarget - totalGasTip

	gasTargetRefactor := gasTargetWithoutTip * uint64(GasRefactor) / uint64(GasDivisor)

	gasTarget := gasTargetWithoutTip + gasTargetRefactor

	gasPrice := gasTarget / uint64(GasDivisor) * uint64(MinGasMultiplier)

	maxGasTarget := gasTarget * uint64(MaxGasMultiplier) / uint64(GasDivisor)

	gp.gasPrice = gasPrice
	gp.gasTarget = gasTarget
	gp.maxGasTarget = maxGasTarget
	gp.latestSync = uint64(time.Now().Unix())

	return nil
}

func (gp *GasPool) CalcGas(data []byte, payloadLen int, valueLen int, tipGas uint64) (uint64, error) {
	gasCost, _, err := calcGasCost(gp.gasTarget, gp.gasPrice, len(data), payloadLen, valueLen, tipGas)
	if err != nil {
		return 0, err
	}

	return gasCost, nil
}

func (gp *GasPool) expired() bool {
	return gp.latestSync <= uint64(time.Now().Unix())-gp.gasInterval
}

func CalcGas(gasTarget uint64, gasPrice uint64, dataLength int, payloadLen int, valueLength int) (uint64, uint64, error) {
	return calcGasCost(gasTarget, gasPrice, dataLength, valueLength, payloadLen, 0)
}

func calcGasCost(gasTarget uint64, gasPrice uint64, dataLength int, payloadLen int, valueLen int, tipGas uint64) (uint64, uint64, error) {
	gasCost := BaseGas

	gasCost += dataLength * BytesCost

	gasCost += payloadLen * PayloadByteCost

	if valueLen > 0 {
		gasCost += valueLen * ValueByteCost
	}

	exactGasTip := calcGasTip(uint64(gasCost))

	if tipGas > uint64(exactGasTip) {
		return 0, 0, fmt.Errorf("gas tip does not match")
	}

	if tipGas > 0 && tipGas < exactGasTip {
		exactGasTip = tipGas
	}

	gasCost += int(exactGasTip)

	if gasCost > int(gasTarget) {
		return 0, 0, fmt.Errorf("gas cost exceeds target")
	}

	if gasCost < int(gasPrice) {
		gasCost = int(gasPrice)
	}

	return uint64(gasCost), uint64(exactGasTip), nil
}

func calcGasTip(gasCost uint64) uint64 {
	return gasCost * uint64(GasTipMultiplier) / uint64(GasDivisor)
}
