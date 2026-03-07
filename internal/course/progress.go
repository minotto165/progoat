package course

import (
	"encoding/json"
	"os"
	"slices"
	"time"
)

type Progress struct {
	CourseID         string    `json:"course_id"`
	CompletedLessons []string  `json:"completed_lessons"`
	CurrentLesson    string    `json:"current_lesson"`
	LastAccessed     time.Time `json:"last_accessed"`
}

func SaveProgress(courseID, completedLessonID, currentLessonID, progressPath string) error {

	var progresses []Progress

	progressJson, err := os.ReadFile(progressPath)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
	} else if len(progressJson) > 0 {
		err = json.Unmarshal(progressJson, &progresses)
		if err != nil {
			return err
		}
	} else {
		progresses = []Progress{}
	}

	// progressesからcourseIDを検索し、インデックスを取得 -> idx int

	idx := -1
	for i, p := range progresses {
		if p.CourseID == courseID {
			idx = i
			break
		}
	}

	if idx == -1 {
		progresses = append(progresses, Progress{courseID, []string{}, "", time.Now()})
		idx = len(progresses) - 1
	}

	if !slices.Contains(progresses[idx].CompletedLessons, completedLessonID) {
		progresses[idx].CompletedLessons = append(progresses[idx].CompletedLessons, completedLessonID)
	}

	progresses[idx].CurrentLesson = currentLessonID
	progresses[idx].LastAccessed = time.Now()

	progressJson, err = json.MarshalIndent(progresses, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(progressPath, progressJson, 0644)
	if err != nil {
		return err
	}

	return nil
}
