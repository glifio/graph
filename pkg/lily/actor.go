package lily

type ActorItem struct {
	ID        string
	Code      string
	Head      string
	Nonce     string
	Balance   string
	StateRoot string
	Height    string
}

type ActorInterface interface {
//	Init() error
//	Create(text string, isDone bool) (*string, error)
//	Update(id string, text string, isDone bool) error
	Get(id string) (*ActorItem, error)
	List() ([]ActorItem, error)
}
