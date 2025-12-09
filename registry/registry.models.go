package registry

func (c *Registry[TData, TResponse, TRequest]) ToModel(data *TData) *TResponse {
	if data == nil {
		return nil
	}
	return c.resource(data)
}

func (c *Registry[TData, TResponse, TRequest]) ToModels(data []*TData) []*TResponse {
	if data == nil {
		return []*TResponse{}
	}
	out := make([]*TResponse, 0, len(data))
	for _, item := range data {
		if m := c.ToModel(item); m != nil {
			out = append(out, m)
		}
	}
	return out
}
