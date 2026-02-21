package redis

import (
	"time"

	"github.com/ugorji/go/codec"

	userDomain "github.com/danicc097/todo-ddd-example/internal/modules/user/domain"
	"github.com/danicc097/todo-ddd-example/internal/modules/workspace/domain"
)

type workspaceCacheDTO struct {
	ID          domain.WorkspaceID                         `json:"id"`
	Name        string                                     `json:"name"`
	Description string                                     `json:"description"`
	CreatedAt   time.Time                                  `json:"created_at"`
	Members     map[userDomain.UserID]domain.WorkspaceRole `json:"members"`
}

type WorkspaceCacheCodec struct {
	handle *codec.MsgpackHandle
}

func NewWorkspaceCacheCodec() *WorkspaceCacheCodec {
	return &WorkspaceCacheCodec{
		handle: &codec.MsgpackHandle{},
	}
}

func (c *WorkspaceCacheCodec) Marshal(w *domain.Workspace) ([]byte, error) {
	dto := workspaceCacheDTO{
		ID:          w.ID(),
		Name:        w.Name().String(),
		Description: w.Description().String(),
		CreatedAt:   w.CreatedAt(),
		Members:     w.Members(),
	}

	var b []byte

	err := codec.NewEncoderBytes(&b, c.handle).Encode(dto)

	return b, err
}

func (c *WorkspaceCacheCodec) Unmarshal(data []byte) (*domain.Workspace, error) {
	var dto workspaceCacheDTO

	dec := codec.NewDecoderBytes(data, c.handle)
	if err := dec.Decode(&dto); err != nil {
		return nil, err
	}

	name, _ := domain.NewWorkspaceName(dto.Name)
	description, _ := domain.NewWorkspaceDescription(dto.Description)

	w := domain.ReconstituteWorkspace(
		dto.ID,
		name,
		description,
		dto.CreatedAt,
		dto.Members,
	)

	return w, nil
}
