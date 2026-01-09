package helper

import (
	"api-jellyfish/internal/pkg/logger"
)

func HandleAppError(err error, function, step string, fatal bool) error {
	if err != nil {
		if fatal {
			logger.Error.Println("Fatal error in function: ", function, "Step: ", step, "Details: ", err)
		} else {
			logger.Error.Println("Error in function: ", function, "Step: ", step, "Details: ", err)
		}
		return err
	}
	return nil
}
