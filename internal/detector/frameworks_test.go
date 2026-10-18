package detector

import "testing"

func TestReduxNeedsConnectCallAndDeclaredRedux(t *testing.T) {
	const (
		reactOnly  = `"react": "18.0.0"`
		withRedux  = `"react": "18.0.0", "react-redux": "9.0.0"`
		disconnect = "export class ResizeObserverMock {\n  observe() {}\n  disconnect() {}\n}\n"
		connect    = "export default connect(mapState)(TodoList)\n"
	)
	tests := []struct {
		name   string
		deps   string
		source string
		want   bool
	}{
		{"disconnect method", reactOnly, disconnect, false},
		{"disconnect method with react-redux", withRedux, disconnect, false},
		{"connect call without redux declared", reactOnly, connect, false},
		{"connect call with react-redux", withRedux, connect, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, infos := writeProject(t, map[string]string{
				"package.json":    `{"dependencies": {` + tt.deps + `}}`,
				"src/adapter.tsx": tt.source,
			})
			conventions, err := NewFrameworkDetector(root, ProjectSourceFiles(infos), ReadDependencies(root, infos)).Detect()
			if err != nil {
				t.Fatal(err)
			}
			got := false
			for _, c := range conventions {
				got = got || c.Description == "Redux for global state management"
			}
			if got != tt.want {
				t.Errorf("Redux convention = %v, want %v; conventions %+v", got, tt.want, conventions)
			}
		})
	}
}
