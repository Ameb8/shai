# Testing Patterns in shai

This reference provides examples of the standard testing patterns used in the shai project.

## Table-Driven Tests with Testify

Prefer table-driven tests for functions with multiple logic paths.

```go
func TestExample(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "success case",
			input:    "valid",
			expected: "VALID",
			wantErr:  false,
		},
		{
			name:     "error case",
			input:    "",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Process(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}
```

## Mocking with Testify

Use `testify/mock` for interface dependencies.

```go
type MockService struct {
	mock.Mock
}

func (m *MockService) DoSomething(ctx context.Context, data string) error {
	args := m.Called(ctx, data)
	return args.Error(0)
}

func TestWithMock(t *testing.T) {
	ms := new(MockService)
	ms.On("DoSomething", mock.Anything, "trigger").Return(nil)

	err := RunBusinessLogic(ms)
	assert.NoError(t, err)
	ms.AssertExpectations(t)
}
```

## Testing IO and Environment

When testing code that relies on environment variables or global state, ensure you cleanup.

```go
func TestEnvDependent(t *testing.T) {
	original := os.Getenv("KEY")
	defer os.Setenv("KEY", original)

	os.Setenv("KEY", "test-value")
	// ... run test
}
```
