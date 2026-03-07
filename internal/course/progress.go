package course

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"time"
)

type Progress struct {
	CourseID         string    `json:"course_id"`
	CompletedLessons []string  `json:"completed_lessons"`
	CurrentLesson    string    `json:"current_lesson"`
	LastAccessed     time.Time `json:"last_accessed"`
	TotalLessons     int       `json:"total_lessons"`
}

type ProgressStatus int

const (
	NotStarted ProgressStatus = iota
	InProgress
	Completed
)

func SaveProgress(courseID, completedLessonID, currentLessonID, progressPath string, totalLessons int) error {

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

	// コース未開始の場合
	if idx == -1 {
		progresses = append(progresses, Progress{courseID, []string{}, "", time.Now(), totalLessons})
		idx = len(progresses) - 1
	}

	// レッスン完了済みは除外
	if !slices.Contains(progresses[idx].CompletedLessons, completedLessonID) {
		progresses[idx].CompletedLessons = append(progresses[idx].CompletedLessons, completedLessonID)
	}

	progresses[idx].CurrentLesson = currentLessonID
	progresses[idx].LastAccessed = time.Now()
	progresses[idx].TotalLessons = totalLessons

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

func LoadProgressStatus(courseID, progressPath string) (ProgressStatus, error) {

	progressJson, err := os.ReadFile(progressPath)
	if err != nil {
		return NotStarted, err
	}

	var progresses []Progress
	json.Unmarshal(progressJson, &progresses)

	for _, p := range progresses {
		if p.CourseID == courseID {

			switch {
			case len(p.CompletedLessons) == p.TotalLessons:
				return Completed, nil
			case len(p.CompletedLessons) == 0:
				return NotStarted, nil
			default:
				return InProgress, nil
			}
		}
	}
	return NotStarted, fmt.Errorf("course not found: %s", courseID)
}
