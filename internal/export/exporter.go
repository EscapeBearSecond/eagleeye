package export

import (
	"context"
)

type Exporter interface {
	Export(c context.Context, v []any) error
	Close()
}
