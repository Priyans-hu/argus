package analyzer

import (
	"cmp"
	"slices"

	"github.com/Priyans-hu/argus/pkg/types"
)

// normalize gives every slice whose order comes from map iteration, file
// walks or goroutine timing a total order, so the same project always yields
// the same analysis. Lists whose order means something (Commands,
// Conventions, KeyFiles, architecture layers) are ordered where they are
// built and left alone.
func normalize(a *types.Analysis) {
	if a == nil {
		return
	}

	slices.SortFunc(a.TechStack.Languages, func(x, y types.Language) int {
		return cmp.Or(
			cmp.Compare(y.Percentage, x.Percentage),
			cmp.Compare(x.Name, y.Name),
			cmp.Compare(x.Version, y.Version),
		)
	})
	slices.SortFunc(a.TechStack.Frameworks, func(x, y types.Framework) int {
		return cmp.Or(
			cmp.Compare(x.Category, y.Category),
			cmp.Compare(x.Name, y.Name),
			cmp.Compare(x.Version, y.Version),
		)
	})
	a.TechStack.Frameworks = slices.Compact(a.TechStack.Frameworks)
	a.TechStack.Databases = sortUnique(a.TechStack.Databases)
	a.TechStack.Tools = sortUnique(a.TechStack.Tools)

	slices.SortFunc(a.Structure.Directories, func(x, y types.Directory) int {
		return cmp.Or(
			cmp.Compare(x.Path, y.Path),
			cmp.Compare(x.Purpose, y.Purpose),
			cmp.Compare(x.FileCount, y.FileCount),
		)
	})
	slices.Sort(a.Structure.RootFiles)

	slices.SortFunc(a.Dependencies, func(x, y types.Dependency) int {
		return cmp.Or(
			cmp.Compare(x.Type, y.Type),
			cmp.Compare(x.Name, y.Name),
			cmp.Compare(x.Version, y.Version),
		)
	})

	if p := a.CodePatterns; p != nil {
		for _, list := range [][]types.PatternInfo{
			p.StateManagement, p.DataFetching, p.Routing, p.Forms, p.Testing,
			p.Styling, p.Authentication, p.APIPatterns, p.DatabaseORM, p.Utilities,
			p.GoPatterns, p.RustPatterns, p.PythonPatterns, p.MLPatterns,
		} {
			slices.SortFunc(list, comparePatterns)
		}
	}

	// the order the Configuration section renders
	slices.SortFunc(a.ConfigFiles, func(x, y types.ConfigFileInfo) int {
		return cmp.Or(cmp.Compare(x.Type, y.Type), cmp.Compare(x.Path, y.Path), cmp.Compare(x.Purpose, y.Purpose))
	})

	slices.SortFunc(a.Endpoints, func(x, y types.Endpoint) int {
		return cmp.Or(
			cmp.Compare(x.File, y.File),
			cmp.Compare(x.Line, y.Line),
			cmp.Compare(x.Method, y.Method),
			cmp.Compare(x.Path, y.Path),
			cmp.Compare(x.Handler, y.Handler),
			cmp.Compare(x.Auth, y.Auth),
			cmp.Compare(x.Description, y.Description),
		)
	})
}

// comparePatterns orders the most widespread patterns first.
func comparePatterns(x, y types.PatternInfo) int {
	return cmp.Or(
		cmp.Compare(y.FileCount, x.FileCount),
		cmp.Compare(x.Category, y.Category),
		cmp.Compare(x.Name, y.Name),
		cmp.Compare(x.Description, y.Description),
		cmp.Compare(x.Usage, y.Usage),
		slices.Compare(x.Examples, y.Examples),
	)
}

func sortUnique(s []string) []string {
	slices.Sort(s)
	return slices.Compact(s)
}
