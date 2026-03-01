package course

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Course struct {
	ID                  string   `json:"course_id"`
	Title               string   `json:"title"`
	Description         string   `json:"description"`
	ProgrammingLanguage string   `json:"programming_language"`
	Lessons             []Lesson `json:"lessons"`
}

type Lesson struct {
	ID              string   `json:"lesson_id"`
	Title           string   `json:"title"`
	Slides          []string `json:"slides"`
	TaskDescription string   `json:"task_description"`
	InitialCode     string   `json:"initial_code"`
	CorrectOutput   string   `json:"correct_output"`
	FileName        string   `json:"file_name"`
}

func GetCourses(coursesPath string) ([]Course, error) {
	files, err := os.ReadDir(coursesPath)
	if err != nil {
		return nil, err
	}
	var courses []Course
	for _, file := range files {
		if file.IsDir() {
			dirName := file.Name()
			coursesJsonPath := filepath.Join(coursesPath, dirName, "course.json")
			coursesJson, err := os.ReadFile(coursesJsonPath)
			if err != nil {
				return nil, err
			}
			// Convert to struct
			var course Course
			err = json.Unmarshal(coursesJson, &course)
			if err != nil {
				return nil, fmt.Errorf("failed to parse JSON: %w", err)
			}

			// Add to slice
			courses = append(courses, course)
		}
	}
	return courses, nil
}

func GetCourseStruct(courseID, coursesPath string) (Course, error) {
	courseJsonPath := filepath.Join(coursesPath, filepath.Base(courseID), "course.json")
	courseJson, err := os.ReadFile(courseJsonPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Course{}, fmt.Errorf("no such a course: %s", courseID)
		}
		return Course{}, err
	}

	var course Course
	if err := json.Unmarshal(courseJson, &course); err != nil {
		return Course{}, fmt.Errorf("failed to parse JSON for course %s: %w", courseID, err)
	}

	return course, nil
}

func SaveCourse(response, coursesPath string) (string, error) {

	// JSON to struct
	var course Course
	err := json.Unmarshal([]byte(response), &course)
	if err != nil {
		return "", fmt.Errorf("failed to parse JSON:%w", err)
	}

	// Crate course directory
	coursePath := filepath.Join(coursesPath, filepath.Base(course.ID))
	os.MkdirAll(coursePath, 0755)

	// Update courses.json
	coursesJsonPath := filepath.Join(coursePath, "course.json")
	coursesJson, err := json.Marshal(course) // Convert to string(JSON)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON:%w", err)
	}

	os.WriteFile(coursesJsonPath, coursesJson, 0755)

	// Create lessons direcotries
	for _, lesson := range course.Lessons {
		lessonPath := filepath.Join(coursePath, filepath.Base(lesson.ID))
		os.MkdirAll(lessonPath, 0755)

		// Create slides - not used
		// slides := lesson.Slides
		// slidesContent, err := json.Marshal(slides)
		// if err != nil {
		// 	return "", fmt.Errorf("failed to marshal JSON:%w", err)
		// }

		// Write Files
		// os.WriteFile(filepath.Join(lessonPath, "slide.json"), slidesContent, 0644)
		os.WriteFile(filepath.Join(lessonPath, "task.md"), []byte(lesson.TaskDescription), 0644)
		os.WriteFile(filepath.Join(lessonPath, filepath.Base(lesson.FileName)), []byte(lesson.InitialCode), 0644)

	}
	return course.Title, nil

}
