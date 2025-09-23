package questionnaire

import (
	"github.com/google/uuid"
)

type QuestionnaireTargetRole struct {
	QuestionnaireID uuid.UUID `gorm:"primaryKey;type:uuid"`
	TargetRoleID    uuid.UUID `gorm:"primaryKey;type:uuid"`
}
