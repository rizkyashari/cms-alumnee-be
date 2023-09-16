package model

import "github.com/google/uuid"

type RelationClassroomSubject struct {
	Base
	ClassroomID uuid.UUID
	SubjectID   uuid.UUID
}
