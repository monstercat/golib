package postgres

import "github.com/Masterminds/squirrel"

type JoinType string

const (
	JoinTypeLeft  JoinType = "left"
	JoinTypeInner JoinType = "inner"
)

type Join struct {
	JoinSql  squirrel.Sqlizer
	JoinType JoinType
}

type JoinStr string

func (j *JoinStr) ToSql() (string, []interface{}, error) {
	return string(*j), nil, nil
}

// JoinMapCollection adds a join to the map.
type JoinMapCollection struct {
	MapCollection[string, Join]

	// Order of the joins to apply.
	Order []string
}

func NewJoinMapCollection() *JoinMapCollection {
	p := &JoinMapCollection{}
	p.Map = make(map[string]Join)
	return p
}

func (c *JoinMapCollection) Add(key string, item Join) {
	c.MapCollection.Add(key, item)
	c.Order = append(c.Order, key)
}

func (c *JoinMapCollection) Apply(o ConditionOption, item Join) (Join, squirrel.Sqlizer) {
	item.JoinSql = o.ModifyCondition(item.JoinSql)
	return item, item.JoinSql
}
