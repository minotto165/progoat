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

	progresses, err := loadProgresses(progressPath)
	if err != nil {
		return err
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

	progressJson, err := json.MarshalIndent(progresses, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(progressPath, progressJson, 0644)
	if err != nil {
		return err
	}

	return nil
}

func ResetProgress(courseID, progressPath string) error {

	progresses, err := loadProgresses(progressPath)
	if err != nil {
		return err
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
		return fmt.Errorf("course not found: %s", courseID)
	}

	progresses = append(progresses[:idx], progresses[idx+1:]...)

	progressJson, err := json.MarshalIndent(progresses, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(progressPath, progressJson, 0644)
	if err != nil {
		return err
	}

	return nil
}

func LoadProgressStatus(courseID, progressPath string) (ProgressStatus, string, error) {

	progresses, err := loadProgresses(progressPath)
	if err != nil {
		return NotStarted, "", err
	}

	for _, p := range progresses {
		if p.CourseID == courseID {

			switch {
			case len(p.CompletedLessons) == p.TotalLessons:
				return Completed, "", nil
			case len(p.CompletedLessons) == 0:
				return NotStarted, "", nil
			default:
				return InProgress, p.CurrentLesson, nil
			}
		}
	}
	return NotStarted, "", nil
}

func loadProgresses(progressPath string) ([]Progress, error) {
	progressJson, err := os.ReadFile(progressPath)
	if err != nil {
		if !os.IsNotExist(err) {
			// "ファイルが存在しない"以外のエラーの場合
			return []Progress{}, err
		}
	}

	if len(progressJson) > 0 {
		// 中身がある場合
		var progresses []Progress
		err = json.Unmarshal(progressJson, &progresses)
		if err != nil {
			return []Progress{}, err
		}
		return progresses, nil
	} else {
		// 中身がない場合
		return []Progress{}, nil
	}
}
