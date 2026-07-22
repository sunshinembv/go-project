package taskservice

import (
	"errors"
	"testing"

	tasksDomain "go-project/internal/domain/task/models"
	taskMocks "go-project/internal/mocks/taskservice"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errRepository = errors.New("repository error")

func TestGetTasks(t *testing.T) {
	type want struct {
		tasks []tasksDomain.Task
		err   error
	}

	type test struct {
		name string
		uid  string
		want want
	}

	tests := []test{
		{
			name: "success",
			uid:  "user-123",
			want: want{
				tasks: []tasksDomain.Task{
					{
						TID: "task-123",
						UID: "user-123",
					},
				},
			},
		},
		{
			name: "repository error",
			uid:  "user-456",
			want: want{
				err: errRepository,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := taskMocks.NewMockRepository(t)
			repo.EXPECT().GetTasks(tc.uid).Return(tc.want.tasks, tc.want.err)

			service := New(repo)
			got, err := service.GetTasks(tc.uid)

			require.ErrorIs(t, err, tc.want.err)
			assert.Equal(t, tc.want.tasks, got)
		})
	}
}

func TestGetTaskByTID(t *testing.T) {
	type want struct {
		task tasksDomain.Task
		err  error
	}

	type test struct {
		name string
		uid  string
		tid  string
		want want
	}

	tests := []test{
		{
			name: "success",
			uid:  "user-123",
			tid:  "task-123",
			want: want{
				task: tasksDomain.Task{
					TID:   "task-123",
					UID:   "user-123",
					Title: "title",
				},
			},
		},
		{
			name: "repository error",
			uid:  "user-123",
			tid:  "task-456",
			want: want{
				err: errRepository,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := taskMocks.NewMockRepository(t)
			repo.EXPECT().GetTaskByTID(tc.uid, tc.tid).Return(tc.want.task, tc.want.err)

			got, err := New(repo).GetTaskByTID(tc.uid, tc.tid)

			require.ErrorIs(t, err, tc.want.err)
			assert.Equal(t, tc.want.task, got)
		})
	}
}

func TestCreateTask(t *testing.T) {
	tid := uuid.MustParse("2d7c583d-240d-4fb9-a690-6c352387f613")
	req := tasksDomain.TaskRequest{
		Title:       "title",
		Description: "description",
		Status:      tasksDomain.InProgressStatus,
	}

	type want struct {
		tid uuid.UUID
		err error
	}

	type test struct {
		name string
		want want
	}

	tests := []test{
		{
			name: "success",
			want: want{
				tid: tid,
			},
		},
		{
			name: "repository error",
			want: want{
				tid: uuid.Nil,
				err: errRepository,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := taskMocks.NewMockRepository(t)
			repo.EXPECT().CreateTask("user-123", tasksDomain.Task{
				Title:       req.Title,
				Description: req.Description,
				Status:      req.Status,
			}).Return(tc.want.tid, tc.want.err)

			got, err := New(repo).CreateTask("user-123", req)

			if tc.want.err != nil {
				require.ErrorIs(t, err, tc.want.err)
				assert.Empty(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want.tid.String(), got)
		})
	}
}

func TestUpdateTaskByTID(t *testing.T) {
	type want struct {
		task tasksDomain.Task
		err  error
	}

	type test struct {
		name string
		uid  string
		tid  string
		req  tasksDomain.TaskRequest
		want want
	}

	tests := []test{
		{
			name: "success",
			uid:  "user-123",
			tid:  "task-123",
			req: tasksDomain.TaskRequest{
				Title:  "updated",
				Status: tasksDomain.CompletedStatus,
			},
			want: want{
				task: tasksDomain.Task{
					TID:    "task-123",
					UID:    "user-123",
					Title:  "updated",
					Status: tasksDomain.CompletedStatus,
				},
			},
		},
		{
			name: "repository error",
			uid:  "user-123",
			tid:  "task-456",
			req: tasksDomain.TaskRequest{
				Title:  "updated",
				Status: tasksDomain.CompletedStatus,
			},
			want: want{
				err: errRepository,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := taskMocks.NewMockRepository(t)
			repo.EXPECT().UpdateTaskByTID(tc.uid, tc.tid, tc.req).Return(tc.want.task, tc.want.err)

			got, err := New(repo).UpdateTaskByTID(tc.uid, tc.tid, tc.req)

			require.ErrorIs(t, err, tc.want.err)
			assert.Equal(t, tc.want.task, got)
		})
	}
}

func TestDeleteTaskByTID(t *testing.T) {
	type want struct {
		err error
	}

	type test struct {
		name string
		uid  string
		tid  string
		want want
	}

	tests := []test{
		{
			name: "success",
			uid:  "user-123",
			tid:  "task-123",
			want: want{
				err: nil,
			},
		},
		{
			name: "repository error",
			uid:  "user-123",
			tid:  "task-456",
			want: want{
				err: errRepository,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := taskMocks.NewMockRepository(t)
			repo.EXPECT().DeleteTaskByTID(tc.uid, tc.tid).Return(tc.want.err)

			err := New(repo).DeleteTaskByTID(tc.uid, tc.tid)

			require.ErrorIs(t, err, tc.want.err)
		})
	}
}
