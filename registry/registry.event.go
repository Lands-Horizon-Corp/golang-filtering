package registry

import "context"

func (r *Registry[TData, TResponse, TRequest]) OnCreate(context context.Context, data *TData) {
	go func() {
		<-context.Done()
		if r.dispatch != nil {
			topics := r.created(data)
			payload := r.ToModel(data)
			r.dispatch(topics, payload)
		}
	}()
}

func (r *Registry[TData, TResponse, TRequest]) OnUpdate(context context.Context, data *TData) {
	go func() {
		<-context.Done()
		if r.dispatch != nil {
			topics := r.created(data)
			payload := r.ToModel(data)
			r.dispatch(topics, payload)
		}
	}()
}

func (r *Registry[TData, TResponse, TRequest]) OnDelete(context context.Context, data *TData) {
	go func() {
		<-context.Done()
		if r.dispatch != nil {
			topics := r.created(data)
			payload := r.ToModel(data)
			r.dispatch(topics, payload)
		}
	}()
}
