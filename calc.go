package calcapi

import (
	calc "calc-kit/gen/calc"
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
)

// calc service example implementation.
// The example methods log the requests and return zero values.
type calcsrvc struct {
	logger log.Logger
}

// NewCalc returns the calc service implementation.
func NewCalc(logger log.Logger) calc.Service {
	return &calcsrvc{logger}
}

// Multiply implements multiply.
func (s *calcsrvc) Multiply(ctx context.Context, p *calc.MultiplyPayload) (res int, err error) {
	s.logger.Log("info", fmt.Sprintf("calc.multiply"))
	return p.A * p.B, nil
}
