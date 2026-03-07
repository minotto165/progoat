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
	var progressJson []byte

	if exists(progressPath) {
		// ファイルがある場合
		progressJson, err := os.ReadFile(progressPath)
		if err != nil {
			return err
		}
		err = json.Unmarshal(progressJson, &progresses)
		if err != nil {
			return err
		}
	} else {
		// ファイルがない場合
		progresses = []Progress{}
	}

	// progressesからcourseIDを検索し、インデックスを取得 -> idx int
	found := false
	idx := 0
	for i, p := range progresses {
		if p.CourseID == courseID {
			found = true
			idx = i
			break
		}
	}

	if !found {
		progresses = append(progresses, Progress{courseID, []string{}, "", time.Now()})
		idx = len(progresses) - 1
	}

	if !slices.Contains(progresses[idx].CompletedLessons, completedLessonID) {
		progresses[idx].CompletedLessons = append(progresses[idx].CompletedLessons, completedLessonID)
	}

	progresses[idx].CurrentLesson = currentLessonID
	progresses[idx].LastAccessed = time.Now()

	progressJson, err := json.Marshal(progresses)
	if err != nil {
		return err
	}

	os.WriteFile(progressPath, progressJson, 0755)

	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
