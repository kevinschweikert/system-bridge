package data_module

import (
	"context"
	"math"
	"time"

	"github.com/charmbracelet/log"
	"github.com/distatus/battery"
	"github.com/timmo001/system-bridge/types"
)

type BatteryModule struct{}

func (batteryModule BatteryModule) Name() types.ModuleName { return types.ModuleBattery }

func (batteryModule BatteryModule) Update(ctx context.Context) (any, error) {
	log.Info("Getting battery data")

	// Get all batteries
	batteries, err := battery.GetAll()
	// If there's an error getting battery info or no batteries found, return empty data
	// This handles both error cases and systems without batteries
	if err != nil {
		log.Debug("No battery present or error getting battery info", "err", err)
		return types.BatteryData{
			IsCharging:    nil,
			Percentage:    nil,
			TimeRemaining: nil,
		}, nil
	}

	// If no batteries found, return empty data
	if len(batteries) == 0 {
		log.Debug("No batteries found")
		return types.BatteryData{
			IsCharging:    nil,
			Percentage:    nil,
			TimeRemaining: nil,
		}, nil
	}

	// Use the first battery (most systems only have one)
	bat := batteries[0]

	// Calculate percentage and ensure it's between 0 and 100
	percentage := math.Min(100, math.Max(0, math.Round((bat.Current/bat.Full)*100)))

	// Determine if charging based on state
	isCharging := bat.State.Raw == battery.Charging

	// Calculate time remaining (in seconds)
	var timeRemaining float64

	// Only calculate if we have valid rate information
	if bat.ChargeRate > 0.001 { // Avoid division by very small numbers
		if isCharging {
			// Time to full
			timeRemaining = ((bat.Full - bat.Current) / bat.ChargeRate) * 3600
		} else {
			// Time to empty - use ChargeRate as discharge rate
			timeRemaining = (bat.Current / bat.ChargeRate) * 3600
		}

		week := 7 * 24 * time.Hour
		// Ensure time remaining is not negative or unreasonably large
		if timeRemaining < 0 || timeRemaining > week.Seconds() {
			timeRemaining = 0 // Reset if unreasonable
		}
	} else {
		timeRemaining = 0
	}

	return types.BatteryData{
		IsCharging:    &isCharging,
		Percentage:    &percentage,
		TimeRemaining: &timeRemaining,
	}, nil
}
