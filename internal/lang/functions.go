// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lang

import (
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/hashicorp/hcl/v2/ext/tryfunc"
	ctyyaml "github.com/zclconf/go-cty-yaml"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"

	"github.com/opentofu/opentofu/internal/addrs"
	"github.com/opentofu/opentofu/internal/experiments"
	"github.com/opentofu/opentofu/internal/lang/funcs"
)

var impureFunctions = []string{
	"bcrypt",
	"timestamp",
	"uuid",
}

var coreFunctions = map[string]function.Function{
	"abs":              stdlib.AbsoluteFunc,
	"abspath":          funcs.AbsPathFunc,
	"alltrue":          funcs.AllTrueFunc,
	"anytrue":          funcs.AnyTrueFunc,
	"basename":         funcs.BasenameFunc,
	"base64decode":     funcs.Base64DecodeFunc,
	"base64encode":     funcs.Base64EncodeFunc,
	"base64gzip":       funcs.Base64GzipFunc,
	"base64gunzip":     funcs.Base64GunzipFunc,
	"base64sha256":     funcs.Base64Sha256Func,
	"base64sha512":     funcs.Base64Sha512Func,
	"bcrypt":           funcs.BcryptFunc,
	"can":              tryfunc.CanFunc,
	"ceil":             stdlib.CeilFunc,
	"chomp":            stdlib.ChompFunc,
	"cidrcontains":     funcs.CidrContainsFunc,
	"cidrhost":         funcs.CidrHostFunc,
	"cidrnetmask":      funcs.CidrNetmaskFunc,
	"cidrsubnet":       funcs.CidrSubnetFunc,
	"cidrsubnets":      funcs.CidrSubnetsFunc,
	"coalesce":         funcs.CoalesceFunc,
	"coalescelist":     stdlib.CoalesceListFunc,
	"compact":          stdlib.CompactFunc,
	"concat":           stdlib.ConcatFunc,
	"contains":         stdlib.ContainsFunc,
	"csvdecode":        stdlib.CSVDecodeFunc,
	"dirname":          funcs.DirnameFunc,
	"distinct":         stdlib.DistinctFunc,
	"element":          stdlib.ElementFunc,
	"endswith":         funcs.EndsWithFunc,
	"chunklist":        stdlib.ChunklistFunc,
	"flatten":          stdlib.FlattenFunc,
	"floor":            stdlib.FloorFunc,
	"format":           stdlib.FormatFunc,
	"formatdate":       stdlib.FormatDateFunc,
	"formatlist":       stdlib.FormatListFunc,
	"indent":           stdlib.IndentFunc,
	"index":            funcs.IndexFunc, // stdlib.IndexFunc is not compatible
	"join":             stdlib.JoinFunc,
	"jsondecode":       stdlib.JSONDecodeFunc,
	"jsonencode":       stdlib.JSONEncodeFunc,
	"keys":             stdlib.KeysFunc,
	"length":           funcs.LengthFunc,
	"list":             funcs.ListFunc,
	"log":              stdlib.LogFunc,
	"lookup":           funcs.LookupFunc,
	"lower":            stdlib.LowerFunc,
	"map":              funcs.MapFunc,
	"matchkeys":        funcs.MatchkeysFunc,
	"max":              stdlib.MaxFunc,
	"md5":              funcs.Md5Func,
	"merge":            stdlib.MergeFunc,
	"min":              stdlib.MinFunc,
	"one":              funcs.OneFunc,
	"parseint":         stdlib.ParseIntFunc,
	"pathexpand":       funcs.PathExpandFunc,
	"pow":              stdlib.PowFunc,
	"range":            stdlib.RangeFunc,
	"regex":            stdlib.RegexFunc,
	"regexall":         stdlib.RegexAllFunc,
	"replace":          funcs.ReplaceFunc,
	"reverse":          stdlib.ReverseListFunc,
	"rsadecrypt":       funcs.RsaDecryptFunc,
	"sensitive":        funcs.SensitiveFunc,
	"nonsensitive":     funcs.NonsensitiveFunc,
	"issensitive":      funcs.IsSensitiveFunc,
	"setintersection":  stdlib.SetIntersectionFunc,
	"setproduct":       stdlib.SetProductFunc,
	"setsubtract":      stdlib.SetSubtractFunc,
	"setunion":         stdlib.SetUnionFunc,
	"sha1":             funcs.Sha1Func,
	"sha256":           funcs.Sha256Func,
	"sha512":           funcs.Sha512Func,
	"signum":           stdlib.SignumFunc,
	"slice":            stdlib.SliceFunc,
	"sort":             stdlib.SortFunc,
	"split":            stdlib.SplitFunc,
	"startswith":       funcs.StartsWithFunc,
	"strcontains":      funcs.StrContainsFunc,
	"strrev":           stdlib.ReverseFunc,
	"substr":           stdlib.SubstrFunc,
	"sum":              funcs.SumFunc,
	"textdecodebase64": funcs.TextDecodeBase64Func,
	"textencodebase64": funcs.TextEncodeBase64Func,
	"timestamp":        funcs.TimestampFunc,
	"timeadd":          stdlib.TimeAddFunc,
	"timecmp":          funcs.TimeCmpFunc,
	"title":            stdlib.TitleFunc,
	"tostring":         funcs.MakeToFunc(cty.String),
	"tonumber":         funcs.MakeToFunc(cty.Number),
	"tobool":           funcs.MakeToFunc(cty.Bool),
	"toset":            funcs.MakeToFunc(cty.Set(cty.DynamicPseudoType)),
	"tolist":           funcs.MakeToFunc(cty.List(cty.DynamicPseudoType)),
	"tomap":            funcs.MakeToFunc(cty.Map(cty.DynamicPseudoType)),
	"transpose":        funcs.TransposeFunc,
	"trim":             stdlib.TrimFunc,
	"trimprefix":       stdlib.TrimPrefixFunc,
	"trimspace":        stdlib.TrimSpaceFunc,
	"trimsuffix":       stdlib.TrimSuffixFunc,
	"try":              tryfunc.TryFunc,
	"upper":            stdlib.UpperFunc,
	"urlencode":        funcs.URLEncodeFunc,
	"urldecode":        funcs.URLDecodeFunc,
	"uuid":             funcs.UUIDFunc,
	"uuidv5":           funcs.UUIDV5Func,
	"values":           stdlib.ValuesFunc,
	"yamldecode":       ctyyaml.YAMLDecodeFunc,
	"yamlencode":       ctyyaml.YAMLEncodeFunc,
	"zipmap":           stdlib.ZipmapFunc,
}

// This should probably be replaced with addrs.Function everywhere
const CoreNamespace = addrs.FunctionNamespaceCore + "::"

// Functions returns the set of functions that should be used to when evaluating
// expressions in the receiving scope.
func (s *Scope) Functions() *map[string]function.Function {
	s.funcsLock.Lock()
	defer s.funcsLock.Unlock()
	if s.funcs == nil {
		s.funcs = cachedFuncs.functions(s)
	}
	return s.funcs
}

func functions(s *Scope) map[string]function.Function {
	// fmt.Printf("Functions(%t, %t, %q, %s)\n", s.PureOnly, s.ConsoleMode, s.BaseDir, s.PlanTimestamp)
	// Some of our functions are just directly the cty stdlib functions.
	// Others are implemented in the subdirectory "funcs" here in this
	// repository. New functions should generally start out their lives
	// in the "funcs" directory and potentially graduate to cty stdlib
	// later if the functionality seems to be something domain-agnostic
	// that would be useful to all applications using cty functions.
	fns := make(map[string]function.Function, 2*(len(coreFunctions)+14))
	maps.Copy(fns, coreFunctions)

	fns["file"] = funcs.MakeFileFunc(s.BaseDir, false)
	fns["fileexists"] = funcs.MakeFileExistsFunc(s.BaseDir)
	fns["fileset"] = funcs.MakeFileSetFunc(s.BaseDir)
	fns["filebase64"] = funcs.MakeFileFunc(s.BaseDir, true)
	fns["filebase64sha256"] = funcs.MakeFileBase64Sha256Func(s.BaseDir)
	fns["filebase64sha512"] = funcs.MakeFileBase64Sha512Func(s.BaseDir)
	fns["filemd5"] = funcs.MakeFileMd5Func(s.BaseDir)
	fns["filesha1"] = funcs.MakeFileSha1Func(s.BaseDir)
	fns["filesha256"] = funcs.MakeFileSha256Func(s.BaseDir)
	fns["filesha512"] = funcs.MakeFileSha512Func(s.BaseDir)

	fns["templatefile"] = funcs.MakeTemplateFileFunc(s.BaseDir, func() map[string]function.Function {
		// The templatefile function prevents recursive calls to itself
		// by copying this map and overwriting the "templatefile" entry.
		return fns
	})

	// Registers "templatestring" function in function map.
	fns["templatestring"] = funcs.MakeTemplateStringFunc(s.BaseDir, func() map[string]function.Function {
		// This anonymous function returns the existing map of functions for initialization.
		return fns
	})

	if s.ConsoleMode {
		// The type function is only available in OpenTofu console.
		fns["type"] = funcs.TypeFunc
	}

	if !s.ConsoleMode {
		// The plantimestamp function doesn't make sense in the OpenTofu
		// console.
		fns["plantimestamp"] = funcs.MakeStaticTimestampFunc(s.PlanTimestamp)
	}

	if s.PureOnly {
		// Force our few impure functions to return unknown so that we
		// can defer evaluating them until a later pass.
		for _, name := range impureFunctions {
			fns[name] = function.Unpredictable(fns[name])
		}
	}

	coreNames := make([]string, 0, len(fns))
	// Add a description to each function and parameter based on the
	// contents of descriptionList.
	// One must create a matching description entry whenever a new
	// function is introduced.
	for name, f := range fns {
		fns[name] = funcs.WithDescription(name, f)
		coreNames = append(coreNames, name)
	}
	// Copy all stdlib funcs into core:: namespace
	for _, name := range coreNames {
		fns[CoreNamespace+name] = fns[name]
	}

	return fns
}

// experimentalFunction checks whether the given experiment is enabled for
// the recieving scope. If so, it will return the given function verbatim.
// If not, it will return a placeholder function that just returns an
// error explaining that the function requires the experiment to be enabled.
//
//lint:ignore U1000 Ignore unused function error for now
func (s *Scope) experimentalFunction(experiment experiments.Experiment, fn function.Function) function.Function {
	if s.activeExperiments.Has(experiment) {
		return fn
	}

	err := fmt.Errorf(
		"this function is experimental and available only when the experiment keyword %s is enabled for the current module",
		experiment.Keyword(),
	)

	return function.New(&function.Spec{
		Params:   fn.Params(),
		VarParam: fn.VarParam(),
		Type: func(args []cty.Value) (cty.Type, error) {
			return cty.DynamicPseudoType, err
		},
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			// It would be weird to get here because the Type function always
			// fails, but we'll return an error here too anyway just to be
			// robust.
			return cty.DynamicVal, err
		},
	})
}

type funcCacheEntry struct {
	planTimestamp time.Time
	functions     *map[string]function.Function
}

type providerFuncCacheEntry struct {
	base *map[string]function.Function
	prov map[string]function.Function
	both map[string]function.Function
}

type funcCache struct {
	sync.Mutex
	cache         map[bool]map[bool]map[string]*funcCacheEntry
	providerCache []providerFuncCacheEntry
}

var cachedFuncs = &funcCache{
	cache:         map[bool]map[bool]map[string]*funcCacheEntry{},
	providerCache: []providerFuncCacheEntry{},
}

func (f *funcCache) functions(s *Scope) *map[string]function.Function {
	f.Lock()
	defer f.Unlock()

	byPurity, ok := f.cache[s.PureOnly]
	if !ok {
		byPurity = map[bool]map[string]*funcCacheEntry{}
		f.cache[s.PureOnly] = byPurity
	}

	byConsole, ok := byPurity[s.ConsoleMode]
	if !ok {
		byConsole = map[string]*funcCacheEntry{}
		byPurity[s.ConsoleMode] = byConsole
	}

	byBaseDir, ok := byConsole[s.BaseDir]
	if !ok {
		fns := functions(s)
		byBaseDir = &funcCacheEntry{
			functions:     &fns,
			planTimestamp: s.PlanTimestamp,
		}
		byConsole[s.BaseDir] = byBaseDir
	} else if !byBaseDir.planTimestamp.Equal(s.PlanTimestamp) {
		fns := functions(s)
		byBaseDir.functions = &fns
		byBaseDir.planTimestamp = s.PlanTimestamp
	}

	return byBaseDir.functions
}

func matchSet(a, b map[string]function.Function) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

func (f *funcCache) providerFunctions(base *map[string]function.Function, prov map[string]function.Function) map[string]function.Function {
	f.Lock()
	defer f.Unlock()

	// See if we've already got a cache entry for this set of provider functions.
	for _, pc := range f.providerCache {
		if pc.base != base {
			continue
		}
		if matchSet(pc.prov, prov) {
			return pc.both
		}
	}

	// We didn't find a match, so we need to create a new entry.
	both := maps.Clone(*base)
	maps.Copy(both, prov)

	f.providerCache = append(f.providerCache, providerFuncCacheEntry{
		base: base,
		prov: prov,
		both: both,
	})

	// fmt.Printf("len(base) == %d\n", len(*base))
	// fmt.Printf("len(prov) == %d\n", len(prov))
	// fmt.Printf("len(both) == %d\n", len(both))
	// fmt.Printf("len(f.providerCache) == %d\n", len(f.providerCache))

	return both
}
