package queue

type Queue struct {
	Items []*[]byte
}

func New() *Queue {
	q := new(Queue)
	q.Items = make([]*[]byte, 0)
	return q
}

func (q *Queue) Set(value *[]byte) {
	q.Items = append(q.Items, value)
}

func (q *Queue) Get() *[]byte {
	if len(q.Items) == 0 {
		return nil
	}

	item, items := q.Items[0], q.Items[1:]
	q.Items = items
	return item
}
