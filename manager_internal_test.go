package loafergo

import "testing"

func TestNewManager(t *testing.T) {
	configWithLogger := &Config{Logger: newDefaultLogger()}
	configWithoutLogger := &Config{}

	testCases := []struct {
		name   string
		input  *Config
		expect func(t *testing.T, m *Manager)
	}{
		{
			name:  "WithLogger",
			input: configWithLogger,
			expect: func(t *testing.T, m *Manager) {
				if m.config.Logger == nil {
					t.Errorf("Test %v failed: expected a logger but got nil", t.Name())
				}
			},
		},
		{
			name:  "WithoutLogger",
			input: configWithoutLogger,
			expect: func(t *testing.T, m *Manager) {
				if m.config.Logger == nil {
					t.Errorf("Test %v failed: expected a logger but got nil", t.Name())
				}
			},
		},
		{
			name:  "WithConfigNil",
			input: nil,
			expect: func(t *testing.T, m *Manager) {
				if m.config.Logger == nil {
					t.Errorf("Test %v failed: expected a logger but got nil", t.Name())
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager := NewManager(tc.input)
			tc.expect(t, manager)
		})
	}
}
