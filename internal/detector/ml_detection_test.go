package detector

import "testing"

func TestMLProjectNeedsImportsOrWeights(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  bool
	}{
		{"tsconfig and config.json files", map[string]string{"tsconfig.json": "{}", "docs/tsconfig.json": "{}", ".changeset/config.json": "{}"}, false},
		{"safetensors weights", map[string]string{"model.safetensors": "w"}, true},
		{"gguf weights", map[string]string{"models/llama.gguf": "w"}, true},
		{"Hugging Face layout", map[string]string{"models/config.json": "{}", "models/pytorch_model.bin": "w"}, true},
		{"bin one directory away from config.json", map[string]string{"config.json": "{}", "models/pytorch_model.bin": "w"}, false},
		{"bin alone", map[string]string{"data/blob.bin": "w"}, false},
		{"joblib alone", map[string]string{"model.joblib": "w"}, false},
		{"pkl alone", map[string]string{"cache.pkl": "w"}, false},
		{"numpy and pandas imports", map[string]string{"src/a.py": "import numpy as np\nimport pandas as pd\n"}, false},
		{"torch import in src", map[string]string{"src/train.py": "import torch\n"}, true},
		{"torch import in docs", map[string]string{"docs/train.py": "import torch\n"}, false},
		{"notebooks and train scripts", map[string]string{"train.py": "x = 1\n", "a.ipynb": "{}", "b.ipynb": "{}", "c.ipynb": "{}",
			"d.ipynb": "{}", "e.ipynb": "{}", "f.ipynb": "{}", "data/train.csv": "a,b\n"}, false},
	}
	for _, tt := range tests {
		// writeProject roots the files in a temporary directory, so the
		// import is read relative to the root, not the working directory.
		root, files := writeProject(t, tt.files)
		got := NewMLDetector(root, ProjectSourceFiles(files)).GetMLPatterns()
		if (len(got) > 0) != tt.want {
			t.Errorf("%s: ML patterns %+v, want ML project %v", tt.name, got, tt.want)
		}
	}
}
