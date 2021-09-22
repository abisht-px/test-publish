package models

import (
	"errors"
	"fmt"
	"strings"

	"github.anim.dreamworks.com/golang/logging"
	"github.com/google/uuid"

	"github.anim.dreamworks.com/DreamCloud/stella-api/utils"
)

type Task struct {
	TaskModel
	TotalSteps         uint     `json:"total_steps"`
	CurrentStep        uint     `json:"current_step"`
	Description        string   `json:"description"`
	Status             string   `json:"status"`
	AssociatedResource string   `json:"associated_resource"`
	Log                []string `json:"log" gorm:"-"`

	DomainID  uuid.UUID `json:"domain_id"`
	ProjectID uuid.UUID `json:"project_id"`
	UserID    uuid.UUID `json:"user_id" mapstructure:"-"`
	UserLogin string    `json:"user_login" mapstructure:"-"`

	LogString string `json:"-" gorm:"column:log"` // Used for DB

	HasWarnings bool `json:"-" gorm:"-"`
	HasErrors   bool `json:"-" gorm:"-"`
}

func NewTask() Task {
	return Task{}
}

func AllTasks() []Task {
	var tasks []Task
	db.Unscoped().Find(&tasks)
	return tasks
}

func TasksCount(
	archived string,
	domainIDs interface{},
	projectIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
) int {
	var count int
	var tasks []Task

	sDB := dbWithScope(archived, domainIDs, projectIDs)
	for k, v := range arrayQueryParams {
		inQuery := fmt.Sprintf("%s IN (?)", k)
		sDB = sDB.Where(inQuery, v)
	}
	if search != "" {
		searchParam := fmt.Sprintf("%%%v%%", search)
		sDB = sDB.Where("description LIKE ?", searchParam)
	}

	sDB.Find(&tasks).Count(&count)
	return count
}

func PaginatedTasks(
	archived string,
	domainIDs interface{},
	projectIDs interface{},
	queryParams map[string]interface{},
	arrayQueryParams map[string][]interface{},
	search string,
	page int,
	perPage int,
	orderBy string,
) []Task {
	var tasks []Task
	offset := (page - 1) * perPage

	if orderBy == "" {
		orderBy = "created_at desc"
	}

	sDB := dbWithScope(archived, domainIDs, projectIDs).Order(orderBy).Offset(offset).Limit(perPage).Where(queryParams)
	for k, v := range arrayQueryParams {
		inQuery := fmt.Sprintf("%s IN (?)", k)
		sDB = sDB.Where(inQuery, v)
	}
	if search != "" {
		searchParam := fmt.Sprintf("%%%v%%", search)
		sDB = sDB.Where("description LIKE ?", searchParam)
	}

	sDB.Find(&tasks)
	return tasks
}

func FindTask(id interface{}) (*Task, error) {
	task := NewTask()
	resp := db.First(&task, id)

	if resp.Error != nil {
		return nil, resp.Error
	} else {
		return &task, nil
	}
}

func FirstTask(query string, params ...interface{}) *Task {
	tasks := TasksWhere(query, params...)
	if len(tasks) == 0 {
		return nil
	}
	return &tasks[0]
}

func TasksWhere(query string, params ...interface{}) []Task {
	var tasks []Task
	db.Where(query, params...).Find(&tasks)
	return tasks
}

func AllTasksWhere(query string, params ...interface{}) []Task {
	var tasks []Task
	db.Unscoped().Where(query, params...).Find(&tasks)
	return tasks
}

func CreateTask(task Task) (*Task, error) {
	if !db.NewRecord(task) {
		err := task.Save()
		return &task, err
	}

	resp := db.Create(&task)
	return &task, resp.Error
}

func FindAndDeleteTask(id uint) error {
	task, err := FindTask(id)
	if err != nil {
		return err
	}
	if task.DeletedAt != nil {
		return nil // Record has already been deleted.
	}
	return task.Delete()
}

// Updates a single field
func (task *Task) Update(attrs ...interface{}) error {
	return db.Unscoped().Model(task).Update(attrs...).Error
}

// Updates multiple fields
func (task *Task) UpdateFields(fields map[string]interface{}) error {
	return db.Unscoped().Model(task).Updates(fields).Error
}

func (task *Task) Save() error {
	if db.NewRecord(*task) {
		_, err := CreateTask(*task)
		return err
	}

	resp := db.Save(task)
	return resp.Error
}

func (task *Task) Delete() error {
	if db.NewRecord(*task) || task.DeletedAt != nil {
		return nil // Record does not exist or has already been deleted.
	}

	resp := db.Delete(task)
	return resp.Error
}

func (task *Task) Reload() error {
	t, err := FindTask(task.ID)
	if err != nil {
		return err
	}
	*task = *t

	return nil
}

//================================================================================
// Helper Functions
//================================================================================

func (task *Task) AddLog(level string, text string) {
	logText := fmt.Sprintf("%v (%v): %v", utils.CurrentTimeString(), level, text)
	logging.Infoln(logText)
	task.Log = append(task.Log, logText)
	if err := task.Save(); err != nil {
		logging.Errorf("Saving task %v. %v", task, err)
	}
}

func (task *Task) AddInfo(text string) {
	task.AddLog("INFO", text)
}

func (task *Task) AddWarning(text string) {
	task.HasWarnings = true
	task.AddLog("WARNING", text)
}

func (task *Task) AddError(text string) {
	task.HasErrors = true
	task.AddLog("ERROR", text)
}

func (task *Task) ResourceURI() string {
	return fmt.Sprintf("/api/db-tasks/%v", task.ID)
}

//================================================================================
// Hooks
//================================================================================

func (task *Task) BeforeSave() (err error) {
	// Validate Required Params
	if err = task.validateRequiredParams(); err != nil {
		return err
	}

	// Serialize Log Array
	if len(task.Log) > 0 {
		task.LogString = strings.Join(task.Log[:], ";;")
	}

	return nil
}

func (task *Task) AfterFind() error {
	// Deserialize Log Array
	if len(task.LogString) > 0 {
		task.Log = strings.Split(task.LogString, ";;")
	} else {
		task.Log = []string{}
	}
	return nil
}

//================================================================================
// Model Validations
//================================================================================

func (task *Task) validateRequiredParams() error {
	var missingParams []string

	if task.Status == "" {
		missingParams = append(missingParams, "status")
	}
	if task.DomainID == EmptyUUID {
		missingParams = append(missingParams, "domain_id")
	}
	if task.ProjectID == EmptyUUID {
		missingParams = append(missingParams, "project_id")
	}

	if len(missingParams) > 0 {
		errMsg := fmt.Sprintf("Missing required param(s): %v", strings.Join(missingParams[:], ", "))
		return errors.New(errMsg)
	}

	return nil
}
