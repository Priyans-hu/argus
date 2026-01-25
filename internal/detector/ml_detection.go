package detector

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// MLFramework represents a detected ML framework
type MLFramework struct {
	Name        string
	Category    string // deep-learning, ml, nlp, cv, mlops, data
	Description string
	Version     string
	FileCount   int
	Examples    []string
}

// MLProjectInfo contains ML-specific project information
type MLProjectInfo struct {
	IsMLProject    bool
	Frameworks     []MLFramework
	ModelFiles     []string
	DatasetPaths   []string
	TrainingFiles  []string
	InferenceFiles []string
	NotebookCount  int
	ProjectType    string // research, production, tutorial, library
}

// MLDetector detects ML frameworks and patterns
type MLDetector struct {
	rootPath string
	files    []types.FileInfo
}

// NewMLDetector creates a new ML detector
func NewMLDetector(rootPath string, files []types.FileInfo) *MLDetector {
	return &MLDetector{rootPath: rootPath, files: files}
}

// Detect performs ML-specific detection
func (d *MLDetector) Detect() *MLProjectInfo {
	info := &MLProjectInfo{}

	// Detect frameworks from imports and dependencies
	info.Frameworks = d.detectFrameworks()

	// Detect model files
	info.ModelFiles = d.detectModelFiles()

	// Detect dataset patterns
	info.DatasetPaths = d.detectDatasetPaths()

	// Detect training and inference files
	info.TrainingFiles, info.InferenceFiles = d.detectMLScripts()

	// Count notebooks
	info.NotebookCount = d.countNotebooks()

	// Determine if this is an ML project
	info.IsMLProject = len(info.Frameworks) > 0 ||
		len(info.ModelFiles) > 0 ||
		info.NotebookCount > 5 ||
		len(info.TrainingFiles) > 0

	// Infer project type
	info.ProjectType = d.inferMLProjectType(info)

	return info
}

// detectFrameworks detects ML frameworks from file content
func (d *MLDetector) detectFrameworks() []MLFramework {
	var frameworks []MLFramework
	frameworkCounts := make(map[string]int)
	frameworkExamples := make(map[string][]string)

	// Framework patterns with import detection
	patterns := []struct {
		name        string
		category    string
		description string
		imports     []string
	}{
		// Deep Learning
		{"TensorFlow", "deep-learning", "Google's deep learning framework", []string{"import tensorflow", "from tensorflow", "import tf"}},
		{"PyTorch", "deep-learning", "Facebook's deep learning framework", []string{"import torch", "from torch"}},
		{"Keras", "deep-learning", "High-level neural networks API", []string{"import keras", "from keras", "from tensorflow.keras"}},
		{"JAX", "deep-learning", "Google's autodiff and XLA library", []string{"import jax", "from jax"}},
		{"Flax", "deep-learning", "Neural network library for JAX", []string{"import flax", "from flax"}},
		{"MXNet", "deep-learning", "Apache deep learning framework", []string{"import mxnet", "from mxnet"}},
		{"PaddlePaddle", "deep-learning", "Baidu's deep learning platform", []string{"import paddle", "from paddle"}},

		// NLP
		{"Transformers", "nlp", "Hugging Face Transformers library", []string{"from transformers", "import transformers"}},
		{"spaCy", "nlp", "Industrial NLP library", []string{"import spacy", "from spacy"}},
		{"NLTK", "nlp", "Natural Language Toolkit", []string{"import nltk", "from nltk"}},
		{"Gensim", "nlp", "Topic modeling library", []string{"import gensim", "from gensim"}},
		{"LangChain", "nlp", "LLM application framework", []string{"from langchain", "import langchain"}},
		{"LlamaIndex", "nlp", "LLM data framework", []string{"from llama_index", "import llama_index"}},

		// Computer Vision
		{"OpenCV", "cv", "Computer vision library", []string{"import cv2", "from cv2"}},
		{"Pillow", "cv", "Python Imaging Library", []string{"from PIL", "import PIL"}},
		{"torchvision", "cv", "PyTorch vision library", []string{"import torchvision", "from torchvision"}},
		{"Detectron2", "cv", "Facebook's detection library", []string{"import detectron2", "from detectron2"}},
		{"MMDetection", "cv", "OpenMMLab detection toolbox", []string{"import mmdet", "from mmdet"}},
		{"YOLO", "cv", "Real-time object detection", []string{"from ultralytics", "import ultralytics"}},
		{"Diffusers", "cv", "Hugging Face diffusion models", []string{"from diffusers", "import diffusers"}},

		// Traditional ML
		{"scikit-learn", "ml", "Machine learning library", []string{"from sklearn", "import sklearn"}},
		{"XGBoost", "ml", "Gradient boosting library", []string{"import xgboost", "from xgboost"}},
		{"LightGBM", "ml", "Light gradient boosting", []string{"import lightgbm", "from lightgbm"}},
		{"CatBoost", "ml", "Gradient boosting on decision trees", []string{"import catboost", "from catboost"}},

		// Data Processing
		{"NumPy", "data", "Numerical computing library", []string{"import numpy", "from numpy"}},
		{"Pandas", "data", "Data analysis library", []string{"import pandas", "from pandas"}},
		{"Polars", "data", "Fast DataFrame library", []string{"import polars", "from polars"}},
		{"Dask", "data", "Parallel computing library", []string{"import dask", "from dask"}},

		// MLOps
		{"MLflow", "mlops", "ML lifecycle platform", []string{"import mlflow", "from mlflow"}},
		{"Weights & Biases", "mlops", "ML experiment tracking", []string{"import wandb", "from wandb"}},
		{"DVC", "mlops", "Data version control", []string{"import dvc", "from dvc"}},
		{"Optuna", "mlops", "Hyperparameter optimization", []string{"import optuna", "from optuna"}},
		{"Ray", "mlops", "Distributed computing framework", []string{"import ray", "from ray"}},
		{"Kubeflow", "mlops", "ML on Kubernetes", []string{"import kfp", "from kfp"}},
		{"BentoML", "mlops", "ML model serving", []string{"import bentoml", "from bentoml"}},
		{"Triton", "mlops", "NVIDIA inference server", []string{"import tritonclient", "from tritonclient"}},

		// Audio
		{"Librosa", "audio", "Audio analysis library", []string{"import librosa", "from librosa"}},
		{"torchaudio", "audio", "PyTorch audio library", []string{"import torchaudio", "from torchaudio"}},
		{"SpeechBrain", "audio", "Speech processing toolkit", []string{"import speechbrain", "from speechbrain"}},

		// Reinforcement Learning
		{"Gymnasium", "rl", "RL environments library", []string{"import gymnasium", "from gymnasium", "import gym", "from gym"}},
		{"Stable-Baselines3", "rl", "RL algorithms library", []string{"from stable_baselines3", "import stable_baselines3"}},
		{"RLlib", "rl", "Ray's RL library", []string{"from ray.rllib", "import rllib"}},
	}

	// Scan Python files for imports
	for _, f := range d.files {
		if !strings.HasSuffix(f.Name, ".py") {
			continue
		}

		content, err := os.ReadFile(f.Path)
		if err != nil {
			continue
		}
		contentStr := string(content)

		for _, p := range patterns {
			for _, imp := range p.imports {
				if strings.Contains(contentStr, imp) {
					frameworkCounts[p.name]++
					if len(frameworkExamples[p.name]) < 3 {
						frameworkExamples[p.name] = append(frameworkExamples[p.name], f.Path)
					}
					break
				}
			}
		}
	}

	// Convert to MLFramework slice
	for _, p := range patterns {
		if count, ok := frameworkCounts[p.name]; ok && count > 0 {
			frameworks = append(frameworks, MLFramework{
				Name:        p.name,
				Category:    p.category,
				Description: p.description,
				FileCount:   count,
				Examples:    frameworkExamples[p.name],
			})
		}
	}

	return frameworks
}

// detectModelFiles finds model weight and checkpoint files
func (d *MLDetector) detectModelFiles() []string {
	var modelFiles []string
	modelExtensions := map[string]bool{
		".pt":          true, // PyTorch
		".pth":         true, // PyTorch
		".ckpt":        true, // Checkpoint
		".h5":          true, // Keras/HDF5
		".hdf5":        true, // HDF5
		".pb":          true, // TensorFlow
		".onnx":        true, // ONNX
		".safetensors": true, // Safe Tensors
		".bin":         true, // Hugging Face
		".pkl":         true, // Pickle (often models)
		".joblib":      true, // Joblib (sklearn)
		".model":       true, // Generic model
		".weights":     true, // Weights
	}

	modelPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)model.*\.(pt|pth|ckpt|h5|onnx|safetensors|bin)$`),
		regexp.MustCompile(`(?i)checkpoint.*\.(pt|pth|ckpt)$`),
		regexp.MustCompile(`(?i)weights.*\.(pt|pth|h5|bin)$`),
		regexp.MustCompile(`(?i)pytorch_model\.bin$`),
		regexp.MustCompile(`(?i)tf_model\.h5$`),
		regexp.MustCompile(`(?i)config\.json$`), // Hugging Face model config
	}

	for _, f := range d.files {
		ext := strings.ToLower(filepath.Ext(f.Name))
		if modelExtensions[ext] {
			modelFiles = append(modelFiles, f.Path)
			continue
		}

		for _, pattern := range modelPatterns {
			if pattern.MatchString(f.Name) {
				modelFiles = append(modelFiles, f.Path)
				break
			}
		}
	}

	return modelFiles
}

// detectDatasetPaths finds dataset directories and files
func (d *MLDetector) detectDatasetPaths() []string {
	var datasetPaths []string
	seen := make(map[string]bool)

	datasetDirPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)^(data|dataset|datasets|train|test|val|validation)$`),
		regexp.MustCompile(`(?i)^(images|labels|annotations)$`),
		regexp.MustCompile(`(?i)^(raw|processed|interim|external)$`),
	}

	datasetFilePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\.(csv|parquet|jsonl|tfrecord|arrow)$`),
		regexp.MustCompile(`(?i)^(train|test|val|validation)\.(csv|json|txt)$`),
	}

	for _, f := range d.files {
		// Check directory names
		dir := filepath.Dir(f.Path)
		dirName := filepath.Base(dir)
		for _, pattern := range datasetDirPatterns {
			if pattern.MatchString(dirName) && !seen[dir] {
				datasetPaths = append(datasetPaths, dir)
				seen[dir] = true
				break
			}
		}

		// Check file patterns
		for _, pattern := range datasetFilePatterns {
			if pattern.MatchString(f.Name) {
				if !seen[f.Path] {
					datasetPaths = append(datasetPaths, f.Path)
					seen[f.Path] = true
				}
				break
			}
		}
	}

	return datasetPaths
}

// detectMLScripts finds training and inference scripts
func (d *MLDetector) detectMLScripts() ([]string, []string) {
	var trainingFiles, inferenceFiles []string

	trainingPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)^train(ing)?\.py$`),
		regexp.MustCompile(`(?i)^(run_)?train(ing|er)?\.py$`),
		regexp.MustCompile(`(?i)^finetune\.py$`),
		regexp.MustCompile(`(?i)^pretrain\.py$`),
	}

	inferencePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)^(run_)?infer(ence)?\.py$`),
		regexp.MustCompile(`(?i)^predict(ion)?\.py$`),
		regexp.MustCompile(`(?i)^eval(uate)?\.py$`),
		regexp.MustCompile(`(?i)^test\.py$`),
		regexp.MustCompile(`(?i)^serve\.py$`),
		regexp.MustCompile(`(?i)^demo\.py$`),
	}

	for _, f := range d.files {
		if !strings.HasSuffix(f.Name, ".py") {
			continue
		}

		for _, pattern := range trainingPatterns {
			if pattern.MatchString(f.Name) {
				trainingFiles = append(trainingFiles, f.Path)
				break
			}
		}

		for _, pattern := range inferencePatterns {
			if pattern.MatchString(f.Name) {
				inferenceFiles = append(inferenceFiles, f.Path)
				break
			}
		}
	}

	return trainingFiles, inferenceFiles
}

// countNotebooks counts Jupyter notebooks
func (d *MLDetector) countNotebooks() int {
	count := 0
	for _, f := range d.files {
		if strings.HasSuffix(f.Name, ".ipynb") {
			count++
		}
	}
	return count
}

// inferMLProjectType determines the type of ML project
func (d *MLDetector) inferMLProjectType(info *MLProjectInfo) string {
	// Check for research indicators
	hasArxivRef := false
	hasPaperRef := false
	hasExperiments := false

	for _, f := range d.files {
		nameLower := strings.ToLower(f.Name)
		if strings.Contains(nameLower, "arxiv") || strings.Contains(nameLower, "paper") {
			hasPaperRef = true
		}
		if strings.Contains(nameLower, "experiment") {
			hasExperiments = true
		}
	}

	// Check README for arxiv links
	readmePath := filepath.Join(d.rootPath, "README.md")
	if content, err := os.ReadFile(readmePath); err == nil {
		contentStr := strings.ToLower(string(content))
		if strings.Contains(contentStr, "arxiv.org") {
			hasArxivRef = true
		}
		if strings.Contains(contentStr, "paper") && strings.Contains(contentStr, "citation") {
			hasPaperRef = true
		}
	}

	// Determine project type
	if hasArxivRef || (hasPaperRef && hasExperiments) {
		return "research"
	}

	// Check for production indicators
	hasDockerfile := false
	hasAPIEndpoint := false
	hasServing := false

	for _, f := range d.files {
		nameLower := strings.ToLower(f.Name)
		if nameLower == "dockerfile" {
			hasDockerfile = true
		}
		if strings.Contains(nameLower, "api") || strings.Contains(nameLower, "serve") {
			hasAPIEndpoint = true
		}
		if strings.Contains(nameLower, "deploy") || strings.Contains(nameLower, "serving") {
			hasServing = true
		}
	}

	if hasDockerfile && (hasAPIEndpoint || hasServing) {
		return "production"
	}

	// Check for tutorial indicators
	if info.NotebookCount > 3 && len(info.TrainingFiles) <= 1 {
		return "tutorial"
	}

	// Check for library indicators
	hasSetupPy := false
	hasPyprojectToml := false
	for _, f := range d.files {
		if f.Name == "setup.py" {
			hasSetupPy = true
		}
		if f.Name == "pyproject.toml" {
			hasPyprojectToml = true
		}
	}

	if (hasSetupPy || hasPyprojectToml) && len(info.Frameworks) > 0 {
		// Check if it exports models/utils
		srcDir := filepath.Join(d.rootPath, "src")
		if _, err := os.Stat(srcDir); err == nil {
			return "library"
		}
	}

	return "project"
}

// GetMLPatterns converts ML detection to pattern info for code patterns
func (d *MLDetector) GetMLPatterns() []types.PatternInfo {
	info := d.Detect()
	if !info.IsMLProject {
		return nil
	}

	var patterns []types.PatternInfo

	// Add framework patterns
	for _, fw := range info.Frameworks {
		patterns = append(patterns, types.PatternInfo{
			Name:        fw.Name,
			Category:    "ml-" + fw.Category,
			Description: fw.Description,
			FileCount:   fw.FileCount,
			Examples:    fw.Examples,
		})
	}

	// Add project type
	if info.ProjectType != "" {
		// Capitalize first letter without deprecated strings.Title
		projectTypeTitle := info.ProjectType
		if len(projectTypeTitle) > 0 {
			projectTypeTitle = strings.ToUpper(projectTypeTitle[:1]) + projectTypeTitle[1:]
		}
		patterns = append(patterns, types.PatternInfo{
			Name:        "ML " + projectTypeTitle,
			Category:    "ml-project-type",
			Description: "This appears to be an ML " + info.ProjectType + " project",
			FileCount:   1,
		})
	}

	// Add model files info
	if len(info.ModelFiles) > 0 {
		patterns = append(patterns, types.PatternInfo{
			Name:        "Model Files",
			Category:    "ml-artifacts",
			Description: "Contains model weight/checkpoint files",
			FileCount:   len(info.ModelFiles),
			Examples:    info.ModelFiles[:min(3, len(info.ModelFiles))],
		})
	}

	return patterns
}
