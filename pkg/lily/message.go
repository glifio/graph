package lily

type MessageItem struct {
	Cid    string  `json:"cid"`
	Height float64 `json:"height"`
	From   string  `json:"from"`
	To     string  `json:"to"`
	Value  float64 `json:"value"`
	Method string  `json:"method"`
	Params *string `json:"params"`
}

type MessageInterface interface {
//	Init() error
//	Create(text string, isDone bool) (*string, error)
//	Update(id string, text string, isDone bool) error
	Get(id string) (*MessageItem, error)
	List(limit int, offset int) ([]MessageItem, error)
}
