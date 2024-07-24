package recorder

type Query struct {
	sqlMode bool
	sql     string
}

type WithQuery func(query *Query)

func WithSQL(sql string) WithQuery {
	return func(query *Query) {
		query.sqlMode = true
		query.sql = sql
	}
}

func NewQuery(query ...WithQuery) *Query {
	q := &Query{}

	for _, o := range query {
		o(q)
	}

	return q
}
