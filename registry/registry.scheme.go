package registry

import (
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
)

func (r *Registry[TData, TResponse, TRequest]) ToSchemaModel() (string, error) {
	reflector := &jsonschema.Reflector{
		DoNotReference: true,
	}
	schema := reflector.Reflect(new(TData))
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}
	return string(schemaJSON), nil
}

func (r *Registry[TData, TResponse, TRequest]) ToSchemaResponse() (string, error) {
	reflector := &jsonschema.Reflector{
		DoNotReference: true,
	}
	schema := reflector.Reflect(new(TResponse))
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}
	return string(schemaJSON), nil
}

func (r *Registry[TData, TResponse, TRequest]) ToSchemaRequest() (string, error) {
	reflector := &jsonschema.Reflector{
		DoNotReference: true,
	}
	schema := reflector.Reflect(new(TRequest))
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal schema: %w", err)
	}
	return string(schemaJSON), nil
}
