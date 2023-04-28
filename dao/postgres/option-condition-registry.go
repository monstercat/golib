package postgres

import "github.com/Masterminds/squirrel"

// Collection is the data type that is passed into OptionConditionRegistry. Specific to this registry, it requires
// an applier to apply the condition option to an item before the item is added to the collection.
type Collection[K comparable, T any] interface {
	// Add adds the item into the collection. Key is for allowing for map types. Slices should ignore the key param.
	Add(key K, item T)

	// Size of the collection
	Size() int

	// Apply adds the condition to the item and returns a new item,
	// as well as the corresponding sqlizer.
	Apply(o ConditionOption, item T) (T, squirrel.Sqlizer)
}

// OptionConditionRegistry defines a container for items which may require preprocessing.
type OptionConditionRegistry[T any, K comparable, C Collection[K, T]] struct {
	// A pointer to the StatementBuilder that is used to perform preprocessing on the items the registry.
	p *StatementBuilder

	// Collection for items that may be in the registry.
	Data C

	// List of items the requires preprocessing.
	Registry []ConditionOptionPreprocess
}

func NewOptionConditionRegistry[T any, K comparable, C Collection[K, T]](p *StatementBuilder, col C) *OptionConditionRegistry[T, K, C] {
	return &OptionConditionRegistry[T, K, C]{
		p:    p,
		Data: col,
	}
}

// Preprocess runs the preprocessor for each item in the registry
func (r *OptionConditionRegistry[T, K, C]) Preprocess(p *ConditionOptionPreprocessParams) {
	for _, item := range r.Registry {
		item.Preprocess(r.p, p)
	}
}

// Add adds an item to the Collection, while applying the ConditionOptions and registering them as necessary.
func (r *OptionConditionRegistry[T, K, C]) Add(key K, item T, xs ...ConditionOption) {
	var sql squirrel.Sqlizer
	for _, x := range xs {
		item, sql = r.Data.Apply(x, item)
		if pre, ok := sql.(ConditionOptionPreprocess); ok {
			r.Registry = append(r.Registry, pre)
		}
	}
	r.Data.Add(key, item)
}

// SliceOptionConditionRegistry is syntactic sugar for collections that must be slices.
type SliceOptionConditionRegistry[T any, C Collection[string, T]] struct {
	*OptionConditionRegistry[T, string, C]
}

func NewSliceConditionRegistry[T any, C Collection[string, T]](p *StatementBuilder, col C) *SliceOptionConditionRegistry[T, C] {
	opt := NewOptionConditionRegistry[T, string, C](p, col)
	return &SliceOptionConditionRegistry[T, C]{opt}
}

func (s *SliceOptionConditionRegistry[T, C]) Add(item T, xs ...ConditionOption) {
	s.OptionConditionRegistry.Add("", item, xs...)
}

// SliceCollection contains a slice that implements Collection's Add and Size functions.
type SliceCollection[T any] struct {
	Slice []T
}

func (c *SliceCollection[T]) Add(_ string, item T) {
	c.Slice = append(c.Slice, item)
}

func (c *SliceCollection[T]) Size() int {
	return len(c.Slice)
}

// MapCollection contains a map that implements Collection's Add and Size functions.
type MapCollection[K comparable, T any] struct {
	Map map[K]T
}

func (c *MapCollection[K, T]) Add(key K, item T) {
	c.Map[key] = item
}

func (c *MapCollection[K, T]) Size() int {
	return len(c.Map)
}

func (c *MapCollection[K, T]) Has(key K) bool {
	_, ok := c.Map[key]
	return ok
}
