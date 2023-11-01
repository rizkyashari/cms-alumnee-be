package model

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	Base
	BeginDate                  time.Time
	EndDate                    time.Time
	Type                       string
	ClassroomID                uuid.UUID
	Classroom                  Classroom
	Title                      *string
	Description                *string
	RelationClassroomSubjectID *uuid.UUID
	RelationClassroomSubject   *RelationClassroomSubject
}
