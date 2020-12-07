package mysql

const GenIdSql = "UPDATE id_gen SET max_id = last_insert_id(max_id + 1) WHERE name = ?"

type IdGenerator struct {
	client *Client
}

func NewIdGenerator(client *Client) *IdGenerator {
	return &IdGenerator{
		client: client,
	}
}

func (ig *IdGenerator) SetClient(client *Client) *IdGenerator {
	ig.client = client
	return ig
}

func (ig *IdGenerator) GenerateId(name string) (int64, error) {
	r, err := ig.client.Exec(GenIdSql, name)
	if err != nil {
		return 0, err
	}

	id, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}
