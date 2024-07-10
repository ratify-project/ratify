package refresh 

import(
	"context"
)

type Refresher interface {
	Refresh(ctx context.Context) error
}
