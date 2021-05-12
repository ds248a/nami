package main

import (
	"context"
	"errors"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/patrickmn/go-cache"
)

var (
	ErrNil = errors.New("no matching record found")
	Ctx    = context.Background()
)

// --------------------------------
//    Hacker Model
// --------------------------------

type HackerModel struct {
	db *redis.Ring
	lc *cache.Cache
}

type dbHacker struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
	rank  int
}

// данные хакера
func (m *HackerModel) hacker(ctx context.Context, name string) (*dbHacker, error) {
	if c, ok := m.lc.Get(`h:` + name); ok {
		return c.(*dbHacker), nil
	}

	pipe := m.db.TxPipeline()
	score := pipe.ZScore(ctx, "hackers", name)
	rank := pipe.ZRank(ctx, "hackers", name)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	if score == nil {
		return nil, ErrNil
	}

	h := &dbHacker{
		Name:  name,
		Score: int(score.Val()),
		rank:  int(rank.Val()),
	}
	m.lc.Set(`h:`+name, h, mTimer)
	return h, nil
}

// список хакеров
func (m *HackerModel) hackersList(ctx context.Context) ([]*dbHacker, bool) {
	if cl, ok := m.lc.Get(`hl`); ok {
		return cl.([]*dbHacker), true
	}

	scores := m.db.ZRangeWithScores(ctx, "hackers", 0, -1)
	if scores == nil {
		return nil, false
	}

	count := len(scores.Val())
	hl := make([]*dbHacker, count)

	for idx, member := range scores.Val() {
		hl[idx] = &dbHacker{
			Name:  member.Member.(string),
			Score: int(member.Score),
		}
	}

	m.lc.Set(`hl`, hl, mTimer)
	return hl, true
}

// создает заданное кол-во записей в наборе хакеров
func (m *HackerModel) hackerNew(ctx context.Context, count int) error {
	m.lc.Delete(`hl`)

	pipe := m.db.TxPipeline()
	defer pipe.Close()

	z := &redis.Z{}
	for i := 0; i < count; i++ {
		z.Member = "name_" + strconv.Itoa(i)
		z.Score = float64(1024 + i)
		pipe.ZAdd(ctx, "hackers", z)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// очистка списка хакеров
func (m *HackerModel) hackerRecover(ctx context.Context) error {
	m.lc.Delete(`hl`)

	hl := []*dbHacker{
		&dbHacker{Name: "Richard Stallman", Score: 1953},
		&dbHacker{Name: "Alan Kay", Score: 1940},
		&dbHacker{Name: "Yukihiro Matsumoto", Score: 1965},
		&dbHacker{Name: "Claude Shannon", Score: 1916},
		&dbHacker{Name: "Linus Torvalds", Score: 1969},
		&dbHacker{Name: "Alan Turing", Score: 1912},
	}

	err := m.db.Del(ctx, "hackers").Err()
	if err != nil {
		return err
	}

	pipe := m.db.TxPipeline()
	defer pipe.Close()

	z := &redis.Z{}
	for _, h := range hl {
		z.Member = h.Name
		z.Score = float64(h.Score)
		pipe.ZAdd(ctx, "hackers", z)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}

	m.lc.Set(`hl`, hl, mTimer)
	return nil
}
