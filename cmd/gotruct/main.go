package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/dave/jennifer/jen"
)

type Converter struct {
	Name string
	Type string
}

type Config struct {
	PackageName     string
	Path            string
	DefaultPrefix   string
	Output          string
	CommentOnStruct string
}

var (
	packageName     *string
	path            *string
	defaultPrefix   *string
	output          *string
	commentOnStruct *string
	config          Config
)

func init() {
	packageName = flag.String("package", "converter", "package name")
	defaultPrefix = flag.String("prefix", "unknow", "default prefix")
	output = flag.String("output", "", "output file")
	commentOnStruct = flag.String("comment", "", "comment on struct")

	flag.Parse()
	if *commentOnStruct == "" {
		*commentOnStruct = fmt.Sprintf("@autowire(set=%s)", *packageName)
	}

	args := flag.Args()

	if len(args) < 0 {
		panic("no input file")
	}

	config = Config{
		PackageName:     *packageName,
		DefaultPrefix:   *defaultPrefix,
		Output:          *output,
		CommentOnStruct: *commentOnStruct,
		Path:            args[0],
	}
}

func main() {
	fs := token.NewFileSet()
	converters := make(map[string][]Converter)
	groupRegExp := regexp.MustCompile(`gotruct:group (\w+)`)
	dirs, _ := os.ReadDir(config.Path)
	for _, dir := range dirs {
		if dir.IsDir() {
			continue
		}
		p := fmt.Sprintf("%s/%s", config.Path, dir.Name())
		f, err := parser.ParseFile(fs, p, nil, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}
		for _, decl := range f.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok {
				docs := genDecl.Doc.Text()
				if strings.Contains(docs, "goverter:converter") {
					group := groupRegExp.FindStringSubmatch(docs)
					if len(group) == 0 {
						group = []string{config.DefaultPrefix}
					}
					for _, spec := range genDecl.Specs {
						if typeSpec, ok := spec.(*ast.TypeSpec); ok {
							converters[strings.Title(group[1])] = append(converters[group[1]], Converter{
								Name: typeSpec.Name.String(),
								Type: typeSpec.Name.String(),
							})
						}
					}
				}
			}
		}
	}

	output, _ := os.Create(config.Output)
	defer output.Close()
	ff := jen.NewFile(config.PackageName)
	for k, c := range converters {
		ff.Comment(config.CommentOnStruct)
		ff.Type().Id(fmt.Sprintf("%sConverter", k)).StructFunc(func(g *jen.Group) {
			for _, converter := range c {
				g.Id(converter.Name).Id(converter.Type)
			}
		})
	}
	ff.Render(output)
}
