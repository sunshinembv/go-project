package tasks

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	taskErrors "go-project/internal/domain/task/errors"
	tasksDomain "go-project/internal/domain/task/models"
	taskMocks "go-project/internal/mocks/tasks"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errService = errors.New("service error")

func newTaskBenchmarkRequest(
	b *testing.B,
	handler gin.HandlerFunc,
	method string,
	route string,
	target string,
	body string,
) *resty.Request {
	b.Helper()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		ctx.Set("userID", "user-123")
		ctx.Next()
	})
	router.Handle(method, route, handler)

	httpSrv := httptest.NewServer(router)
	b.Cleanup(httpSrv.Close)

	req := resty.New().R()
	req.Method = method
	req.URL = httpSrv.URL + target

	if body != "" {
		req.Body = body
		req.SetHeader("Content-Type", "application/json")
	}

	return req
}

func BenchmarkGetTasks(b *testing.B) {
	service := taskMocks.NewMockTaskService(b)
	service.EXPECT().
		GetTasks("user-123").
		Return([]tasksDomain.Task{
			{
				TID:   "task-123",
				UID:   "user-123",
				Title: "title",
			},
		}, nil)

	handler := New(service)
	req := newTaskBenchmarkRequest(
		b,
		handler.GetTasks,
		http.MethodGet,
		"/tasks",
		"/tasks",
		"",
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req.Send()
	}
}

func BenchmarkGetTaskByTID(b *testing.B) {
	service := taskMocks.NewMockTaskService(b)
	service.EXPECT().
		GetTaskByTID("user-123", "task-123").
		Return(tasksDomain.Task{
			TID:   "task-123",
			UID:   "user-123",
			Title: "title",
		}, nil)

	handler := New(service)
	req := newTaskBenchmarkRequest(
		b,
		handler.GetTaskByTID,
		http.MethodGet,
		"/tasks/:id",
		"/tasks/task-123",
		"",
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req.Send()
	}
}

func BenchmarkCreateTask(b *testing.B) {
	reqBody := `{"title":"title","description":"description","status":"new"}`
	reqObj := tasksDomain.TaskRequest{
		Title:       "title",
		Description: "description",
		Status:      tasksDomain.NewStatus,
	}

	service := taskMocks.NewMockTaskService(b)
	service.EXPECT().
		CreateTask("user-123", reqObj).
		Return("task-123", nil)

	handler := New(service)
	req := newTaskBenchmarkRequest(
		b,
		handler.CreateTask,
		http.MethodPost,
		"/tasks",
		"/tasks",
		reqBody,
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req.Send()
	}
}

func BenchmarkUpdateTaskByTID(b *testing.B) {
	reqBody := `{"title":"updated","description":"description","status":"completed"}`
	reqObj := tasksDomain.TaskRequest{
		Title:       "updated",
		Description: "description",
		Status:      tasksDomain.CompletedStatus,
	}

	service := taskMocks.NewMockTaskService(b)
	service.EXPECT().
		UpdateTaskByTID("user-123", "task-123", reqObj).
		Return(tasksDomain.Task{
			TID:         "task-123",
			UID:         "user-123",
			Title:       reqObj.Title,
			Description: reqObj.Description,
			Status:      reqObj.Status,
		}, nil)

	handler := New(service)
	req := newTaskBenchmarkRequest(
		b,
		handler.UpdateTaskByTID,
		http.MethodPut,
		"/tasks/:id",
		"/tasks/task-123",
		reqBody,
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req.Send()
	}
}

func BenchmarkDeleteTaskByTID(b *testing.B) {
	service := taskMocks.NewMockTaskService(b)
	service.EXPECT().
		DeleteTaskByTID("user-123", "task-123").
		Return(nil)

	handler := New(service)
	req := newTaskBenchmarkRequest(
		b,
		handler.DeleteTaskByTID,
		http.MethodDelete,
		"/tasks/:id",
		"/tasks/task-123",
		"",
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req.Send()
	}
}

func performTaskRequest(
	t *testing.T,
	handler gin.HandlerFunc,
	method string,
	route string,
	target string,
	body string,
	uid string,
) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(func(ctx *gin.Context) {
		if uid != "" {
			ctx.Set("userID", uid)
		}
		ctx.Next()
	})
	router.Handle(method, route, handler)

	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func TestGetTasks(t *testing.T) {
	type want struct {
		statusCode int
		tasks      []tasksDomain.Task
		err        error
		body       string
	}

	type test struct {
		name   string
		uid    string
		called bool
		want   want
	}

	tests := []test{
		{
			name: "unauthorized",
			want: want{
				statusCode: http.StatusUnauthorized,
				body:       "unauthorized",
			},
		},
		{
			name:   "success",
			uid:    "user-123",
			called: true,
			want: want{
				statusCode: http.StatusOK,
				tasks: []tasksDomain.Task{
					{
						TID:   "task-123",
						UID:   "user-123",
						Title: "title",
					},
				},
				body: "task-123",
			},
		},
		{
			name:   "not found",
			uid:    "user-123",
			called: true,
			want: want{
				statusCode: http.StatusNotFound,
				err:        taskErrors.ErrTaskNotFound,
				body:       taskErrors.ErrTaskNotFound.Error(),
			},
		},
		{
			name:   "service error",
			uid:    "user-123",
			called: true,
			want: want{
				statusCode: http.StatusInternalServerError,
				err:        errService,
				body:       errService.Error(),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := taskMocks.NewMockTaskService(t)
			if tc.called {
				service.EXPECT().GetTasks(tc.uid).Return(tc.want.tasks, tc.want.err)
			}

			resp := performTaskRequest(t, New(service).GetTasks, http.MethodGet, "/tasks", "/tasks", "", tc.uid)

			require.Equal(t, tc.want.statusCode, resp.Code)
			assert.Contains(t, resp.Body.String(), tc.want.body)
		})
	}
}

func TestGetTaskByTID(t *testing.T) {
	type want struct {
		statusCode int
		task       tasksDomain.Task
		err        error
	}

	type test struct {
		name   string
		uid    string
		called bool
		want   want
	}

	tests := []test{
		{
			name: "unauthorized",
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name:   "success",
			uid:    "user-123",
			called: true,
			want: want{
				statusCode: http.StatusOK,
				task: tasksDomain.Task{
					TID:   "task-123",
					UID:   "user-123",
					Title: "title",
				},
			},
		},
		{
			name:   "not found",
			uid:    "user-123",
			called: true,
			want: want{
				statusCode: http.StatusNotFound,
				err:        taskErrors.ErrTaskNotFound,
			},
		},
		{
			name:   "service error",
			uid:    "user-123",
			called: true,
			want: want{
				statusCode: http.StatusInternalServerError,
				err:        errService,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := taskMocks.NewMockTaskService(t)
			if tc.called {
				service.EXPECT().GetTaskByTID(tc.uid, "task-123").Return(tc.want.task, tc.want.err)
			}

			resp := performTaskRequest(t, New(service).GetTaskByTID, http.MethodGet, "/tasks/:id", "/tasks/task-123", "", tc.uid)

			require.Equal(t, tc.want.statusCode, resp.Code)
			if tc.want.err != nil {
				assert.Contains(t, resp.Body.String(), tc.want.err.Error())
			}
			if tc.want.statusCode == http.StatusOK {
				assert.Contains(t, resp.Body.String(), tc.want.task.TID)
			}
		})
	}
}

func TestCreateTask(t *testing.T) {
	validBody := `{"title":"title","description":"description","status":"new"}`
	req := tasksDomain.TaskRequest{
		Title:       "title",
		Description: "description",
		Status:      tasksDomain.NewStatus,
	}

	type want struct {
		statusCode int
		tid        string
		err        error
	}

	type test struct {
		name   string
		uid    string
		body   string
		called bool
		want   want
	}

	tests := []test{
		{
			name: "unauthorized",
			body: validBody,
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "invalid json",
			uid:  "user-123",
			body: `{invalid`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "success",
			uid:    "user-123",
			body:   validBody,
			called: true,
			want: want{
				statusCode: http.StatusCreated,
				tid:        "task-123",
			},
		},
		{
			name:   "service error",
			uid:    "user-123",
			body:   validBody,
			called: true,
			want: want{
				statusCode: http.StatusInternalServerError,
				err:        errService,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := taskMocks.NewMockTaskService(t)
			if tc.called {
				service.EXPECT().CreateTask(tc.uid, req).Return(tc.want.tid, tc.want.err)
			}

			resp := performTaskRequest(t, New(service).CreateTask, http.MethodPost, "/tasks", "/tasks", tc.body, tc.uid)

			require.Equal(t, tc.want.statusCode, resp.Code)
			if tc.want.tid != "" {
				assert.Contains(t, resp.Body.String(), tc.want.tid)
			}
			if tc.want.err != nil {
				assert.Contains(t, resp.Body.String(), tc.want.err.Error())
			}
		})
	}
}

func TestUpdateTaskByTID(t *testing.T) {
	validBody := `{"title":"updated","description":"description","status":"completed"}`
	req := tasksDomain.TaskRequest{
		Title:       "updated",
		Description: "description",
		Status:      tasksDomain.CompletedStatus,
	}
	wantTask := tasksDomain.Task{
		TID:         "task-123",
		UID:         "user-123",
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	}

	type want struct {
		statusCode int
		task       tasksDomain.Task
		err        error
	}

	type test struct {
		name   string
		uid    string
		body   string
		called bool
		want   want
	}

	tests := []test{
		{
			name: "unauthorized",
			body: validBody,
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name: "invalid json",
			uid:  "user-123",
			body: `{invalid`,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:   "success",
			uid:    "user-123",
			body:   validBody,
			called: true,
			want: want{
				statusCode: http.StatusOK,
				task:       wantTask,
			},
		},
		{
			name:   "not found",
			uid:    "user-123",
			body:   validBody,
			called: true,
			want: want{
				statusCode: http.StatusNotFound,
				err:        taskErrors.ErrTaskNotFound,
			},
		},
		{
			name:   "service error",
			uid:    "user-123",
			body:   validBody,
			called: true,
			want: want{
				statusCode: http.StatusInternalServerError,
				err:        errService,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := taskMocks.NewMockTaskService(t)
			if tc.called {
				service.EXPECT().UpdateTaskByTID(tc.uid, "task-123", req).Return(tc.want.task, tc.want.err)
			}

			resp := performTaskRequest(t, New(service).UpdateTaskByTID, http.MethodPut, "/tasks/:id", "/tasks/task-123", tc.body, tc.uid)

			require.Equal(t, tc.want.statusCode, resp.Code)
			if tc.want.statusCode == http.StatusOK {
				assert.Contains(t, resp.Body.String(), wantTask.Title)
			}
			if tc.want.err != nil {
				assert.Contains(t, resp.Body.String(), tc.want.err.Error())
			}
		})
	}
}

func TestDeleteTaskByTID(t *testing.T) {
	type want struct {
		statusCode int
		err        error
	}

	type test struct {
		name   string
		uid    string
		called bool
		want   want
	}

	tests := []test{
		{
			name: "unauthorized",
			want: want{
				statusCode: http.StatusUnauthorized,
			},
		},
		{
			name:   "success",
			uid:    "user-123",
			called: true,
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:   "not found",
			uid:    "user-123",
			called: true,
			want: want{
				statusCode: http.StatusNotFound,
				err:        taskErrors.ErrTaskNotFound,
			},
		},
		{
			name:   "service error",
			uid:    "user-123",
			called: true,
			want: want{
				statusCode: http.StatusInternalServerError,
				err:        errService,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			service := taskMocks.NewMockTaskService(t)
			if tc.called {
				service.EXPECT().DeleteTaskByTID(tc.uid, "task-123").Return(tc.want.err)
			}

			resp := performTaskRequest(t, New(service).DeleteTaskByTID, http.MethodDelete, "/tasks/:id", "/tasks/task-123", "", tc.uid)

			require.Equal(t, tc.want.statusCode, resp.Code)
			if tc.want.err != nil {
				assert.Contains(t, resp.Body.String(), tc.want.err.Error())
			} else if tc.want.statusCode == http.StatusOK {
				assert.Contains(t, resp.Body.String(), "task deleted")
			}
		})
	}
}
