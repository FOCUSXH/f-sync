package global

import (
	"fsync/client/models"

	"go.uber.org/zap"
)

var (
	Configs *models.Config
	Logger  *zap.Logger
)
