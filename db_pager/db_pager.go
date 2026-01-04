package db_pager

import (
	"context"
)

type FetchPage[T any] func(limit uint, offset uint) ([]T, error)

type Pager[T any] struct {
	limit  uint
	offset uint
	fetch  FetchPage[T]
	out    chan T
}

func NewPager[T any](fetch FetchPage[T], limit uint) *Pager[T] {
	return &Pager[T]{
		limit:  limit,
		offset: 0,
		fetch:  fetch,
		out:    make(chan T),
	}
}

func (p *Pager[T]) Out() <-chan T {
	return p.out
}

func (p *Pager[T]) Run(ctx context.Context) {
	for {
		results, err := p.fetch(p.limit, p.offset)

		if err != nil || len(results) == 0 {
			break
		}

		p.offset += p.limit

		for _, result := range results {
			p.out <- result
		}
	}

	close(p.out)
}
