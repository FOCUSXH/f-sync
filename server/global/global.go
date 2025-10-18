package global

import (
	"fsync/server/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	Configs *models.Config
	Logger  *zap.Logger
	DB      *gorm.DB
)
